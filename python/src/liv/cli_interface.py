"""
CLI Interface for the LIV Python SDK

This module provides a Python interface to the existing Go CLI tools,
allowing seamless integration with the LIV command-line utilities.
"""

import subprocess
import json
import shutil
import os
from pathlib import Path
from typing import Dict, List, Optional, Any, Union
import tempfile
import time

from .exceptions import CLIError, ConfigurationError
from .models import ValidationResult, ConversionResult
from .config_manager import ConfigManager


class CLIInterface:
    """Interface to the LIV CLI tools."""
    
    def __init__(self, cli_path: Optional[str] = None, config_manager: Optional[ConfigManager] = None):
        """
        Initialize CLI interface.
        
        Args:
            cli_path: Path to the LIV CLI executable
            config_manager: Configuration manager instance
        """
        self.config_manager = config_manager or ConfigManager()
        self.cli_path = cli_path or self._find_cli_executable()
        self.temp_dir = Path(tempfile.gettempdir()) / "liv-python"
        self.temp_dir.mkdir(exist_ok=True)
        
        # Verify CLI is available
        self._verify_cli_available()
    
    def _find_cli_executable(self) -> str:
        """Find the LIV CLI executable in PATH or config."""
        # Check config first
        cli_path = self.config_manager.get("cli_path")
        if cli_path and Path(cli_path).exists():
            return cli_path
        
        # Check environment variable
        env_path = os.environ.get("LIV_CLI_PATH")
        if env_path and Path(env_path).exists():
            return env_path
        
        # Search in PATH
        cli_names = ["liv", "liv.exe", "liv-cli", "liv-cli.exe"]
        for name in cli_names:
            path = shutil.which(name)
            if path:
                return path
        
        raise ConfigurationError(
            "LIV CLI executable not found. Please install the LIV CLI tools or set LIV_CLI_PATH environment variable."
        )
    
    def _verify_cli_available(self) -> None:
        """Verify that the CLI is available and working."""
        try:
            result = self._run_command(["--version"], capture_output=True)
            if result.returncode != 0:
                raise CLIError(f"CLI version check failed: {result.stderr}")
        except FileNotFoundError:
            raise ConfigurationError(f"CLI executable not found at: {self.cli_path}")
        except Exception as e:
            raise CLIError(f"Failed to verify CLI availability: {e}")
    
    def _run_command(self, args: List[str], capture_output: bool = True, 
                    timeout: Optional[float] = None, cwd: Optional[Path] = None) -> subprocess.CompletedProcess:
        """
        Run a CLI command.
        
        Args:
            args: Command arguments
            capture_output: Whether to capture stdout/stderr
            timeout: Command timeout in seconds
            cwd: Working directory
            
        Returns:
            CompletedProcess result
        """
        cmd = [self.cli_path] + args
        
        try:
            result = subprocess.run(
                cmd,
                capture_output=capture_output,
                text=True,
                timeout=timeout or self.config_manager.get("command_timeout", 300),
                cwd=cwd
            )
            return result
        except subprocess.TimeoutExpired as e:
            raise CLIError(f"Command timed out after {timeout}s", command=" ".join(cmd))
        except subprocess.CalledProcessError as e:
            raise CLIError(
                f"Command failed with exit code {e.returncode}",
                command=" ".join(cmd),
                exit_code=e.returncode,
                stderr=e.stderr
            )
        except Exception as e:
            raise CLIError(f"Failed to execute command: {e}", command=" ".join(cmd))
    
    def build(self, source_dir: Union[str, Path], output_path: Union[str, Path],
              manifest_path: Optional[Union[str, Path]] = None,
              sign: bool = False, key_path: Optional[Union[str, Path]] = None,
              compress: bool = True, validate: bool = True) -> bool:
        """
        Build a LIV document from source directory.
        
        Args:
            source_dir: Source directory containing document files
            output_path: Output path for the .liv file
            manifest_path: Optional manifest file path
            sign: Whether to sign the document
            key_path: Path to signing key
            compress: Whether to compress assets
            validate: Whether to validate the result
            
        Returns:
            True if successful
        """
        args = ["build"]
        
        # Add source directory
        args.extend(["--source", str(source_dir)])
        
        # Add output path
        args.extend(["--output", str(output_path)])
        
        # Add optional arguments
        if manifest_path:
            args.extend(["--manifest", str(manifest_path)])
        
        if sign and key_path:
            args.extend(["--sign", "--key", str(key_path)])
        
        if not compress:
            args.append("--no-compress")
        
        if not validate:
            args.append("--no-validate")
        
        result = self._run_command(args)
        
        if result.returncode != 0:
            raise CLIError(
                f"Build command failed: {result.stderr}",
                command=" ".join(args),
                exit_code=result.returncode,
                stderr=result.stderr
            )
        
        return True
    
    def validate(self, file_path: Union[str, Path], 
                strict: bool = False, check_signatures: bool = True) -> ValidationResult:
        """
        Validate a LIV document.
        
        Args:
            file_path: Path to the .liv file
            strict: Whether to use strict validation
            check_signatures: Whether to verify signatures
            
        Returns:
            ValidationResult object
        """
        args = ["validate", str(file_path)]
        
        if strict:
            args.append("--strict")
        
        if not check_signatures:
            args.append("--no-signatures")
        
        # Add JSON output for easier parsing
        args.extend(["--format", "json"])
        
        start_time = time.time()
        result = self._run_command(args)
        validation_time = time.time() - start_time
        
        if result.returncode == 0:
            # Parse JSON output
            try:
                data = json.loads(result.stdout)
                return ValidationResult(
                    is_valid=data.get("valid", True),
                    errors=data.get("errors", []),
                    warnings=data.get("warnings", []),
                    validation_time=validation_time,
                    file_path=Path(file_path)
                )
            except json.JSONDecodeError:
                # Fallback to simple parsing
                return ValidationResult(
                    is_valid=True,
                    errors=[],
                    warnings=[],
                    validation_time=validation_time,
                    file_path=Path(file_path)
                )
        else:
            # Parse error output
            errors = [result.stderr] if result.stderr else ["Validation failed"]
            return ValidationResult(
                is_valid=False,
                errors=errors,
                warnings=[],
                validation_time=validation_time,
                file_path=Path(file_path)
            )
    
    def convert(self, input_path: Union[str, Path], output_path: Union[str, Path],
               target_format: str, options: Optional[Dict[str, Any]] = None) -> ConversionResult:
        """
        Convert a document to another format.
        
        Args:
            input_path: Input file path
            output_path: Output file path
            target_format: Target format (pdf, html, markdown, epub)
            options: Additional conversion options
            
        Returns:
            ConversionResult object
        """
        args = ["convert"]
        args.extend(["--input", str(input_path)])
        args.extend(["--output", str(output_path)])
        args.extend(["--format", target_format])
        
        # Add options
        if options:
            for key, value in options.items():
                if isinstance(value, bool):
                    if value:
                        args.append(f"--{key}")
                else:
                    args.extend([f"--{key}", str(value)])
        
        input_path = Path(input_path)
        output_path = Path(output_path)
        
        # Get input file size
        file_size_before = input_path.stat().st_size if input_path.exists() else None
        
        start_time = time.time()
        result = self._run_command(args)
        conversion_time = time.time() - start_time
        
        # Get output file size
        file_size_after = output_path.stat().st_size if output_path.exists() else None
        
        success = result.returncode == 0 and output_path.exists()
        
        return ConversionResult(
            success=success,
            input_path=input_path,
            output_path=output_path if success else None,
            source_format=input_path.suffix.lstrip('.'),
            target_format=target_format,
            conversion_time=conversion_time,
            file_size_before=file_size_before,
            file_size_after=file_size_after,
            errors=[result.stderr] if result.stderr and not success else [],
            warnings=[]
        )
    
    def sign(self, file_path: Union[str, Path], key_path: Union[str, Path],
            output_path: Optional[Union[str, Path]] = None) -> bool:
        """
        Sign a LIV document.
        
        Args:
            file_path: Path to the .liv file
            key_path: Path to the signing key
            output_path: Optional output path (defaults to overwriting input)
            
        Returns:
            True if successful
        """
        args = ["sign"]
        args.extend(["--file", str(file_path)])
        args.extend(["--key", str(key_path)])
        
        if output_path:
            args.extend(["--output", str(output_path)])
        
        result = self._run_command(args)
        
        if result.returncode != 0:
            raise CLIError(
                f"Sign command failed: {result.stderr}",
                command=" ".join(args),
                exit_code=result.returncode,
                stderr=result.stderr
            )
        
        return True
    
    def verify(self, file_path: Union[str, Path], 
              public_key_path: Optional[Union[str, Path]] = None) -> bool:
        """
        Verify signatures in a LIV document.
        
        Args:
            file_path: Path to the .liv file
            public_key_path: Optional path to public key
            
        Returns:
            True if signatures are valid
        """
        args = ["verify", str(file_path)]
        
        if public_key_path:
            args.extend(["--key", str(public_key_path)])
        
        result = self._run_command(args)
        return result.returncode == 0
    
    def view(self, file_path: Union[str, Path], port: int = 8080, 
            open_browser: bool = True) -> subprocess.Popen:
        """
        Start the LIV viewer server.
        
        Args:
            file_path: Path to the .liv file
            port: Server port
            open_browser: Whether to open browser automatically
            
        Returns:
            Popen process object
        """
        args = ["view", str(file_path)]
        args.extend(["--port", str(port)])
        
        if not open_browser:
            args.append("--no-browser")
        
        # Start as background process
        process = subprocess.Popen(
            [self.cli_path] + args,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        
        return process
    
    def get_version(self) -> str:
        """Get CLI version."""
        result = self._run_command(["--version"])
        if result.returncode == 0:
            return result.stdout.strip()
        else:
            raise CLIError("Failed to get CLI version")
    
    def get_help(self, command: Optional[str] = None) -> str:
        """Get help text for CLI commands."""
        args = ["--help"]
        if command:
            args = [command, "--help"]
        
        result = self._run_command(args)
        return result.stdout if result.returncode == 0 else result.stderr
    
    def list_formats(self) -> List[str]:
        """List supported conversion formats."""
        result = self._run_command(["convert", "--list-formats"])
        if result.returncode == 0:
            # Parse format list from output
            formats = []
            for line in result.stdout.split('\n'):
                line = line.strip()
                if line and not line.startswith('#'):
                    formats.append(line)
            return formats
        else:
            # Return default formats if command not available
            return ["pdf", "html", "markdown", "epub"]
    
    def cleanup_temp_files(self) -> None:
        """Clean up temporary files created during operations."""
        try:
            if self.temp_dir.exists():
                import shutil
                shutil.rmtree(self.temp_dir)
                self.temp_dir.mkdir(exist_ok=True)
        except Exception:
            pass  # Ignore cleanup errors
    
    def __del__(self):
        """Cleanup when object is destroyed."""
        self.cleanup_temp_files()


# Utility functions for common CLI operations
def quick_validate(file_path: Union[str, Path]) -> bool:
    """Quick validation check for a LIV file."""
    cli = CLIInterface()
    result = cli.validate(file_path)
    return result.is_valid


def quick_convert(input_path: Union[str, Path], output_path: Union[str, Path], 
                 target_format: str) -> bool:
    """Quick conversion between formats."""
    cli = CLIInterface()
    result = cli.convert(input_path, output_path, target_format)
    return result.success


def get_cli_version() -> str:
    """Get the CLI version."""
    cli = CLIInterface()
    return cli.get_version()


__all__ = [
    "CLIInterface",
    "quick_validate",
    "quick_convert", 
    "get_cli_version",
]