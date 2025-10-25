"""
Document validation for the LIV Python SDK
"""

from pathlib import Path
from typing import Union, List, Optional
import time

from .models import ValidationResult
from .cli_interface import CLIInterface
from .exceptions import ValidationError, LIVError
from .config_manager import ConfigManager


class LIVValidator:
    """Validates LIV documents using CLI tools and built-in checks."""
    
    def __init__(self, config_manager: Optional[ConfigManager] = None):
        """
        Initialize validator.
        
        Args:
            config_manager: Configuration manager instance
        """
        self.config_manager = config_manager or ConfigManager()
        self.cli = CLIInterface(config_manager=self.config_manager)
    
    def validate_file(self, file_path: Union[str, Path], 
                     strict: Optional[bool] = None,
                     check_signatures: Optional[bool] = None) -> ValidationResult:
        """
        Validate a LIV document file.
        
        Args:
            file_path: Path to .liv file
            strict: Use strict validation (from config if not specified)
            check_signatures: Check signatures (from config if not specified)
            
        Returns:
            ValidationResult object
        """
        file_path = Path(file_path)
        
        # Get validation config
        validation_config = self.config_manager.get_validation_config()
        if strict is None:
            strict = validation_config.get("strict_mode", False)
        if check_signatures is None:
            check_signatures = validation_config.get("check_signatures", True)
        
        # Basic file checks
        basic_errors = self._basic_file_validation(file_path)
        if basic_errors:
            return ValidationResult(
                is_valid=False,
                errors=basic_errors,
                warnings=[],
                file_path=file_path
            )
        
        # Use CLI for detailed validation
        try:
            return self.cli.validate(file_path, strict=strict, check_signatures=check_signatures)
        except Exception as e:
            return ValidationResult(
                is_valid=False,
                errors=[f"Validation failed: {e}"],
                warnings=[],
                file_path=file_path
            )
    
    def _basic_file_validation(self, file_path: Path) -> List[str]:
        """Perform basic file validation checks."""
        errors = []
        
        # Check file exists
        if not file_path.exists():
            errors.append(f"File not found: {file_path}")
            return errors
        
        # Check file extension
        if file_path.suffix.lower() != '.liv':
            errors.append(f"Invalid file extension: {file_path.suffix} (expected .liv)")
        
        # Check file size
        try:
            file_size = file_path.stat().st_size
            if file_size == 0:
                errors.append("File is empty")
            elif file_size > 500 * 1024 * 1024:  # 500MB limit
                errors.append(f"File is too large: {file_size} bytes (max 500MB)")
        except Exception as e:
            errors.append(f"Cannot read file stats: {e}")
        
        # Check if it's a valid ZIP file (basic check)
        try:
            with open(file_path, 'rb') as f:
                header = f.read(4)
                if header != b'PK\x03\x04':  # ZIP file signature
                    errors.append("File is not a valid ZIP archive")
        except Exception as e:
            errors.append(f"Cannot read file header: {e}")
        
        return errors
    
    def validate_multiple(self, file_paths: List[Union[str, Path]],
                         strict: Optional[bool] = None,
                         check_signatures: Optional[bool] = None) -> List[ValidationResult]:
        """
        Validate multiple LIV documents.
        
        Args:
            file_paths: List of file paths to validate
            strict: Use strict validation
            check_signatures: Check signatures
            
        Returns:
            List of ValidationResult objects
        """
        results = []
        
        for file_path in file_paths:
            try:
                result = self.validate_file(file_path, strict=strict, check_signatures=check_signatures)
                results.append(result)
            except Exception as e:
                # Create error result for failed validation
                result = ValidationResult(
                    is_valid=False,
                    errors=[f"Validation error: {e}"],
                    warnings=[],
                    file_path=Path(file_path)
                )
                results.append(result)
        
        return results
    
    def validate_directory(self, directory: Union[str, Path],
                          pattern: str = "*.liv",
                          recursive: bool = False,
                          strict: Optional[bool] = None,
                          check_signatures: Optional[bool] = None) -> List[ValidationResult]:
        """
        Validate all LIV documents in a directory.
        
        Args:
            directory: Directory to search
            pattern: File pattern to match
            recursive: Search recursively
            strict: Use strict validation
            check_signatures: Check signatures
            
        Returns:
            List of ValidationResult objects
        """
        directory = Path(directory)
        
        if not directory.exists():
            raise LIVError(f"Directory not found: {directory}")
        
        if not directory.is_dir():
            raise LIVError(f"Path is not a directory: {directory}")
        
        # Find LIV files
        if recursive:
            file_paths = list(directory.rglob(pattern))
        else:
            file_paths = list(directory.glob(pattern))
        
        if not file_paths:
            return []
        
        return self.validate_multiple(file_paths, strict=strict, check_signatures=check_signatures)
    
    def get_validation_summary(self, results: List[ValidationResult]) -> dict:
        """
        Get summary statistics for validation results.
        
        Args:
            results: List of validation results
            
        Returns:
            Summary dictionary
        """
        total = len(results)
        valid = sum(1 for r in results if r.is_valid)
        invalid = total - valid
        
        total_errors = sum(len(r.errors) for r in results)
        total_warnings = sum(len(r.warnings) for r in results)
        
        avg_time = None
        if results and any(r.validation_time for r in results):
            times = [r.validation_time for r in results if r.validation_time]
            avg_time = sum(times) / len(times) if times else None
        
        return {
            "total_files": total,
            "valid_files": valid,
            "invalid_files": invalid,
            "success_rate": (valid / total * 100) if total > 0 else 0,
            "total_errors": total_errors,
            "total_warnings": total_warnings,
            "average_validation_time": avg_time,
            "files_with_errors": [r.file_path for r in results if r.errors],
            "files_with_warnings": [r.file_path for r in results if r.warnings]
        }
    
    def validate_content_structure(self, html_content: str, css_content: str = "",
                                 js_content: str = "") -> List[str]:
        """
        Validate content structure without creating a full document.
        
        Args:
            html_content: HTML content to validate
            css_content: CSS content to validate
            js_content: JavaScript content to validate
            
        Returns:
            List of validation errors
        """
        errors = []
        
        # Basic HTML validation
        if html_content:
            html_errors = self._validate_html_content(html_content)
            errors.extend(html_errors)
        
        # Basic CSS validation
        if css_content:
            css_errors = self._validate_css_content(css_content)
            errors.extend(css_errors)
        
        # Basic JS validation
        if js_content:
            js_errors = self._validate_js_content(js_content)
            errors.extend(js_errors)
        
        return errors
    
    def _validate_html_content(self, html: str) -> List[str]:
        """Basic HTML content validation."""
        errors = []
        
        if not html.strip():
            errors.append("HTML content is empty")
            return errors
        
        # Check for basic HTML structure
        html_lower = html.lower()
        
        # Check for dangerous content
        dangerous_tags = ['<script', '<object', '<embed', '<applet']
        for tag in dangerous_tags:
            if tag in html_lower:
                errors.append(f"Potentially dangerous HTML tag found: {tag}")
        
        # Check for inline event handlers
        event_handlers = ['onclick', 'onload', 'onerror', 'onmouseover']
        for handler in event_handlers:
            if handler in html_lower:
                errors.append(f"Inline event handler found: {handler}")
        
        # Basic structure checks
        if '<html' not in html_lower and '<!doctype' not in html_lower:
            errors.append("HTML content may be missing DOCTYPE or html tag")
        
        return errors
    
    def _validate_css_content(self, css: str) -> List[str]:
        """Basic CSS content validation."""
        errors = []
        
        if not css.strip():
            return errors  # Empty CSS is valid
        
        # Check for dangerous CSS
        css_lower = css.lower()
        
        dangerous_properties = ['behavior:', '-moz-binding:', 'javascript:']
        for prop in dangerous_properties:
            if prop in css_lower:
                errors.append(f"Potentially dangerous CSS property found: {prop}")
        
        # Basic syntax check (count braces)
        open_braces = css.count('{')
        close_braces = css.count('}')
        if open_braces != close_braces:
            errors.append(f"Mismatched CSS braces: {open_braces} open, {close_braces} close")
        
        return errors
    
    def _validate_js_content(self, js: str) -> List[str]:
        """Basic JavaScript content validation."""
        errors = []
        
        if not js.strip():
            return errors  # Empty JS is valid
        
        # Check for dangerous functions
        js_lower = js.lower()
        
        dangerous_functions = ['eval(', 'function(', 'settimeout(', 'setinterval(']
        for func in dangerous_functions:
            if func in js_lower:
                errors.append(f"Potentially dangerous JavaScript function found: {func}")
        
        # Check for DOM access
        if 'document.' in js_lower or 'window.' in js_lower:
            errors.append("Direct DOM/window access found (may be restricted in sandbox)")
        
        return errors
    
    def is_valid_liv_file(self, file_path: Union[str, Path]) -> bool:
        """
        Quick check if a file is a valid LIV document.
        
        Args:
            file_path: Path to file
            
        Returns:
            True if file appears to be a valid LIV document
        """
        try:
            result = self.validate_file(file_path)
            return result.is_valid
        except Exception:
            return False


# Utility functions
def quick_validate(file_path: Union[str, Path]) -> bool:
    """Quick validation check for a LIV file."""
    validator = LIVValidator()
    return validator.is_valid_liv_file(file_path)


def validate_content(html: str, css: str = "", js: str = "") -> List[str]:
    """Validate content without creating a document."""
    validator = LIVValidator()
    return validator.validate_content_structure(html, css, js)


__all__ = [
    "LIVValidator",
    "quick_validate",
    "validate_content",
]