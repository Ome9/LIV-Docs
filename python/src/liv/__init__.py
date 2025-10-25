"""
LIV Document Format Python SDK

A Python SDK for automating LIV document creation, validation, and batch processing.
This SDK provides a high-level Python interface to the LIV Document Format CLI tools.
"""

__version__ = "0.1.0"
__author__ = "LIV Document Format Team"
__email__ = "team@liv-format.org"

# Core classes
from .document import LIVDocument
from .builder import LIVBuilder
from .validator import LIVValidator
from .converter import LIVConverter
from .batch_processor import LIVBatchProcessor

# Data models
from .models import (
    DocumentMetadata,
    SecurityPolicy,
    WASMPermissions,
    JSPermissions,
    NetworkPolicy,
    StoragePolicy,
    ValidationResult,
    ConversionResult,
    AssetInfo,
    WASMModuleInfo
)

# Utilities
from .cli_interface import CLIInterface
from .asset_manager import AssetManager
from .config_manager import ConfigManager

# Async support
try:
    from .async_processor import AsyncLIVProcessor
    __all_async__ = ["AsyncLIVProcessor"]
except ImportError:
    __all_async__ = []

# Exceptions
from .exceptions import (
    LIVError,
    ValidationError,
    ConversionError,
    CLIError,
    ConfigurationError,
    AssetError,
    SecurityError
)

# Main exports
__all__ = [
    # Core classes
    "LIVDocument",
    "LIVBuilder", 
    "LIVValidator",
    "LIVConverter",
    "LIVBatchProcessor",
    
    # Data models
    "DocumentMetadata",
    "SecurityPolicy",
    "WASMPermissions",
    "JSPermissions", 
    "NetworkPolicy",
    "StoragePolicy",
    "ValidationResult",
    "ConversionResult",
    "AssetInfo",
    "WASMModuleInfo",
    
    # Utilities
    "CLIInterface",
    "AssetManager",
    "ConfigManager",
    
    # Exceptions
    "LIVError",
    "ValidationError",
    "ConversionError",
    "CLIError",
    "ConfigurationError",
    "AssetError",
    "SecurityError",
    
    # Version info
    "__version__",
    "__author__",
    "__email__",
] + __all_async__

# Package metadata
__package_info__ = {
    "name": "liv-document-format",
    "version": __version__,
    "description": "Python SDK for LIV Document Format automation and batch processing",
    "author": __author__,
    "author_email": __email__,
    "url": "https://github.com/liv-document-format/liv-python",
    "license": "MIT",
    "python_requires": ">=3.8",
}