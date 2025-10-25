"""
Exception classes for the LIV Python SDK
"""

from typing import List, Optional, Dict, Any


class LIVError(Exception):
    """Base exception class for all LIV-related errors."""
    
    def __init__(self, message: str, details: Optional[Dict[str, Any]] = None):
        super().__init__(message)
        self.message = message
        self.details = details or {}
    
    def __str__(self) -> str:
        if self.details:
            return f"{self.message} (Details: {self.details})"
        return self.message


class ValidationError(LIVError):
    """Raised when document validation fails."""
    
    def __init__(self, message: str, errors: Optional[List[str]] = None, 
                 warnings: Optional[List[str]] = None):
        super().__init__(message)
        self.errors = errors or []
        self.warnings = warnings or []
    
    def __str__(self) -> str:
        parts = [self.message]
        if self.errors:
            parts.append(f"Errors: {', '.join(self.errors)}")
        if self.warnings:
            parts.append(f"Warnings: {', '.join(self.warnings)}")
        return " | ".join(parts)


class ConversionError(LIVError):
    """Raised when format conversion fails."""
    
    def __init__(self, message: str, source_format: Optional[str] = None,
                 target_format: Optional[str] = None, source_file: Optional[str] = None):
        super().__init__(message)
        self.source_format = source_format
        self.target_format = target_format
        self.source_file = source_file
    
    def __str__(self) -> str:
        parts = [self.message]
        if self.source_format and self.target_format:
            parts.append(f"Converting from {self.source_format} to {self.target_format}")
        if self.source_file:
            parts.append(f"Source: {self.source_file}")
        return " | ".join(parts)


class CLIError(LIVError):
    """Raised when CLI command execution fails."""
    
    def __init__(self, message: str, command: Optional[str] = None, 
                 exit_code: Optional[int] = None, stderr: Optional[str] = None):
        super().__init__(message)
        self.command = command
        self.exit_code = exit_code
        self.stderr = stderr
    
    def __str__(self) -> str:
        parts = [self.message]
        if self.command:
            parts.append(f"Command: {self.command}")
        if self.exit_code is not None:
            parts.append(f"Exit code: {self.exit_code}")
        if self.stderr:
            parts.append(f"Error output: {self.stderr}")
        return " | ".join(parts)


class ConfigurationError(LIVError):
    """Raised when configuration is invalid or missing."""
    
    def __init__(self, message: str, config_key: Optional[str] = None,
                 config_file: Optional[str] = None):
        super().__init__(message)
        self.config_key = config_key
        self.config_file = config_file
    
    def __str__(self) -> str:
        parts = [self.message]
        if self.config_key:
            parts.append(f"Config key: {self.config_key}")
        if self.config_file:
            parts.append(f"Config file: {self.config_file}")
        return " | ".join(parts)


class AssetError(LIVError):
    """Raised when asset operations fail."""
    
    def __init__(self, message: str, asset_path: Optional[str] = None,
                 asset_type: Optional[str] = None):
        super().__init__(message)
        self.asset_path = asset_path
        self.asset_type = asset_type
    
    def __str__(self) -> str:
        parts = [self.message]
        if self.asset_type:
            parts.append(f"Asset type: {self.asset_type}")
        if self.asset_path:
            parts.append(f"Asset path: {self.asset_path}")
        return " | ".join(parts)


class SecurityError(LIVError):
    """Raised when security validation or policy enforcement fails."""
    
    def __init__(self, message: str, policy_violation: Optional[str] = None,
                 security_level: Optional[str] = None):
        super().__init__(message)
        self.policy_violation = policy_violation
        self.security_level = security_level
    
    def __str__(self) -> str:
        parts = [self.message]
        if self.policy_violation:
            parts.append(f"Policy violation: {self.policy_violation}")
        if self.security_level:
            parts.append(f"Security level: {self.security_level}")
        return " | ".join(parts)


class WASMError(LIVError):
    """Raised when WASM module operations fail."""
    
    def __init__(self, message: str, module_name: Optional[str] = None,
                 wasm_error_code: Optional[str] = None):
        super().__init__(message)
        self.module_name = module_name
        self.wasm_error_code = wasm_error_code
    
    def __str__(self) -> str:
        parts = [self.message]
        if self.module_name:
            parts.append(f"WASM module: {self.module_name}")
        if self.wasm_error_code:
            parts.append(f"WASM error code: {self.wasm_error_code}")
        return " | ".join(parts)


class NetworkError(LIVError):
    """Raised when network operations fail."""
    
    def __init__(self, message: str, url: Optional[str] = None,
                 status_code: Optional[int] = None):
        super().__init__(message)
        self.url = url
        self.status_code = status_code
    
    def __str__(self) -> str:
        parts = [self.message]
        if self.url:
            parts.append(f"URL: {self.url}")
        if self.status_code:
            parts.append(f"Status code: {self.status_code}")
        return " | ".join(parts)


class TimeoutError(LIVError):
    """Raised when operations timeout."""
    
    def __init__(self, message: str, timeout_seconds: Optional[float] = None,
                 operation: Optional[str] = None):
        super().__init__(message)
        self.timeout_seconds = timeout_seconds
        self.operation = operation
    
    def __str__(self) -> str:
        parts = [self.message]
        if self.operation:
            parts.append(f"Operation: {self.operation}")
        if self.timeout_seconds:
            parts.append(f"Timeout: {self.timeout_seconds}s")
        return " | ".join(parts)


# Exception hierarchy for easy catching
__all__ = [
    "LIVError",
    "ValidationError", 
    "ConversionError",
    "CLIError",
    "ConfigurationError",
    "AssetError",
    "SecurityError",
    "WASMError",
    "NetworkError",
    "TimeoutError",
]