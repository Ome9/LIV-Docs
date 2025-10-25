"""
Configuration management for the LIV Python SDK
"""

import json
import os
from pathlib import Path
from typing import Dict, Any, Optional, Union
import logging

from .exceptions import ConfigurationError


class ConfigManager:
    """Manages configuration for the LIV Python SDK."""
    
    DEFAULT_CONFIG = {
        "cli_path": None,
        "temp_dir": None,
        "command_timeout": 300,  # 5 minutes
        "log_level": "INFO",
        "default_security_policy": {
            "wasm_memory_limit": 67108864,  # 64MB
            "allow_networking": False,
            "allow_file_system": False
        },
        "batch_processing": {
            "max_concurrent": 4,
            "timeout": 300,
            "retry_attempts": 3
        },
        "conversion": {
            "pdf_quality": "high",
            "html_include_assets": True,
            "markdown_preserve_formatting": True
        },
        "validation": {
            "strict_mode": False,
            "check_signatures": True,
            "validate_wasm": True
        }
    }
    
    def __init__(self, config_file: Optional[Union[str, Path]] = None):
        """
        Initialize configuration manager.
        
        Args:
            config_file: Optional path to configuration file
        """
        self.config_file = self._find_config_file(config_file)
        self.config = self.DEFAULT_CONFIG.copy()
        self._load_config()
        self._setup_logging()
    
    def _find_config_file(self, config_file: Optional[Union[str, Path]]) -> Optional[Path]:
        """Find configuration file in standard locations."""
        if config_file:
            path = Path(config_file)
            if path.exists():
                return path
            else:
                raise ConfigurationError(f"Config file not found: {config_file}")
        
        # Search in standard locations
        search_paths = [
            Path.cwd() / "liv.config.json",
            Path.cwd() / ".liv" / "config.json",
            Path.home() / ".liv" / "config.json",
            Path.home() / ".config" / "liv" / "config.json",
        ]
        
        for path in search_paths:
            if path.exists():
                return path
        
        return None
    
    def _load_config(self) -> None:
        """Load configuration from file and environment variables."""
        # Load from file
        if self.config_file:
            try:
                with open(self.config_file, 'r', encoding='utf-8') as f:
                    file_config = json.load(f)
                self._merge_config(self.config, file_config)
            except Exception as e:
                raise ConfigurationError(
                    f"Failed to load config file: {e}",
                    config_file=str(self.config_file)
                )
        
        # Load from environment variables
        self._load_env_config()
    
    def _load_env_config(self) -> None:
        """Load configuration from environment variables."""
        env_mappings = {
            "LIV_CLI_PATH": "cli_path",
            "LIV_TEMP_DIR": "temp_dir",
            "LIV_LOG_LEVEL": "log_level",
            "LIV_COMMAND_TIMEOUT": ("command_timeout", int),
            "LIV_MAX_CONCURRENT": ("batch_processing.max_concurrent", int),
            "LIV_STRICT_VALIDATION": ("validation.strict_mode", bool),
        }
        
        for env_var, config_key in env_mappings.items():
            value = os.environ.get(env_var)
            if value is not None:
                if isinstance(config_key, tuple):
                    key, converter = config_key
                    try:
                        if converter == bool:
                            value = value.lower() in ('true', '1', 'yes', 'on')
                        else:
                            value = converter(value)
                    except ValueError:
                        continue
                    self._set_nested_config(key, value)
                else:
                    self.config[config_key] = value
    
    def _merge_config(self, base: Dict[str, Any], update: Dict[str, Any]) -> None:
        """Recursively merge configuration dictionaries."""
        for key, value in update.items():
            if key in base and isinstance(base[key], dict) and isinstance(value, dict):
                self._merge_config(base[key], value)
            else:
                base[key] = value
    
    def _set_nested_config(self, key: str, value: Any) -> None:
        """Set nested configuration value using dot notation."""
        keys = key.split('.')
        config = self.config
        
        for k in keys[:-1]:
            if k not in config:
                config[k] = {}
            config = config[k]
        
        config[keys[-1]] = value
    
    def _setup_logging(self) -> None:
        """Setup logging based on configuration."""
        log_level = self.config.get("log_level", "INFO").upper()
        logging.basicConfig(
            level=getattr(logging, log_level, logging.INFO),
            format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        )
    
    def get(self, key: str, default: Any = None) -> Any:
        """
        Get configuration value.
        
        Args:
            key: Configuration key (supports dot notation)
            default: Default value if key not found
            
        Returns:
            Configuration value
        """
        keys = key.split('.')
        value = self.config
        
        try:
            for k in keys:
                value = value[k]
            return value
        except (KeyError, TypeError):
            return default
    
    def set(self, key: str, value: Any) -> None:
        """
        Set configuration value.
        
        Args:
            key: Configuration key (supports dot notation)
            value: Value to set
        """
        self._set_nested_config(key, value)
    
    def save(self, config_file: Optional[Union[str, Path]] = None) -> None:
        """
        Save configuration to file.
        
        Args:
            config_file: Optional path to save to (defaults to current config file)
        """
        save_path = Path(config_file) if config_file else self.config_file
        
        if not save_path:
            # Create default config file location
            config_dir = Path.home() / ".liv"
            config_dir.mkdir(exist_ok=True)
            save_path = config_dir / "config.json"
        
        try:
            with open(save_path, 'w', encoding='utf-8') as f:
                json.dump(self.config, f, indent=2)
            self.config_file = save_path
        except Exception as e:
            raise ConfigurationError(f"Failed to save config: {e}", config_file=str(save_path))
    
    def reset_to_defaults(self) -> None:
        """Reset configuration to defaults."""
        self.config = self.DEFAULT_CONFIG.copy()
    
    def validate_config(self) -> None:
        """Validate current configuration."""
        errors = []
        
        # Check CLI path if specified
        cli_path = self.get("cli_path")
        if cli_path and not Path(cli_path).exists():
            errors.append(f"CLI path does not exist: {cli_path}")
        
        # Check temp directory if specified
        temp_dir = self.get("temp_dir")
        if temp_dir:
            temp_path = Path(temp_dir)
            if not temp_path.exists():
                try:
                    temp_path.mkdir(parents=True, exist_ok=True)
                except Exception:
                    errors.append(f"Cannot create temp directory: {temp_dir}")
        
        # Validate numeric values
        timeout = self.get("command_timeout")
        if timeout is not None and (not isinstance(timeout, (int, float)) or timeout <= 0):
            errors.append("command_timeout must be a positive number")
        
        max_concurrent = self.get("batch_processing.max_concurrent")
        if max_concurrent is not None and (not isinstance(max_concurrent, int) or max_concurrent <= 0):
            errors.append("batch_processing.max_concurrent must be a positive integer")
        
        if errors:
            raise ConfigurationError(f"Configuration validation failed: {'; '.join(errors)}")
    
    def get_cli_path(self) -> Optional[str]:
        """Get CLI path from configuration."""
        return self.get("cli_path")
    
    def get_temp_dir(self) -> Path:
        """Get temporary directory path."""
        temp_dir = self.get("temp_dir")
        if temp_dir:
            return Path(temp_dir)
        else:
            import tempfile
            return Path(tempfile.gettempdir()) / "liv-python"
    
    def get_security_policy(self) -> Dict[str, Any]:
        """Get default security policy."""
        return self.get("default_security_policy", {})
    
    def get_batch_config(self) -> Dict[str, Any]:
        """Get batch processing configuration."""
        return self.get("batch_processing", {})
    
    def get_conversion_config(self) -> Dict[str, Any]:
        """Get conversion configuration."""
        return self.get("conversion", {})
    
    def get_validation_config(self) -> Dict[str, Any]:
        """Get validation configuration."""
        return self.get("validation", {})
    
    def to_dict(self) -> Dict[str, Any]:
        """Get configuration as dictionary."""
        return self.config.copy()
    
    def __repr__(self) -> str:
        """String representation of configuration."""
        return f"ConfigManager(config_file={self.config_file})"


# Global configuration instance
_global_config: Optional[ConfigManager] = None


def get_global_config() -> ConfigManager:
    """Get global configuration instance."""
    global _global_config
    if _global_config is None:
        _global_config = ConfigManager()
    return _global_config


def set_global_config(config: ConfigManager) -> None:
    """Set global configuration instance."""
    global _global_config
    _global_config = config


__all__ = [
    "ConfigManager",
    "get_global_config",
    "set_global_config",
]