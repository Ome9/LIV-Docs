"""
Data models for the LIV Python SDK
"""

from dataclasses import dataclass, field
from typing import Dict, List, Optional, Any, Union
from datetime import datetime
from pathlib import Path
import json


@dataclass
class DocumentMetadata:
    """Metadata for a LIV document."""
    title: str
    author: str
    description: str = ""
    version: str = "1.0"
    language: str = "en"
    created: Optional[datetime] = None
    modified: Optional[datetime] = None
    keywords: List[str] = field(default_factory=list)
    custom_fields: Dict[str, Any] = field(default_factory=dict)
    
    def __post_init__(self):
        if self.created is None:
            self.created = datetime.now()
        if self.modified is None:
            self.modified = datetime.now()
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            "title": self.title,
            "author": self.author,
            "description": self.description,
            "version": self.version,
            "language": self.language,
            "created": self.created.isoformat() if self.created else None,
            "modified": self.modified.isoformat() if self.modified else None,
            "keywords": self.keywords,
            **self.custom_fields
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "DocumentMetadata":
        """Create from dictionary."""
        created = None
        if data.get("created"):
            created = datetime.fromisoformat(data["created"])
        
        modified = None
        if data.get("modified"):
            modified = datetime.fromisoformat(data["modified"])
        
        # Extract known fields and put the rest in custom_fields
        known_fields = {
            "title", "author", "description", "version", 
            "language", "created", "modified", "keywords"
        }
        custom_fields = {k: v for k, v in data.items() if k not in known_fields}
        
        return cls(
            title=data["title"],
            author=data["author"],
            description=data.get("description", ""),
            version=data.get("version", "1.0"),
            language=data.get("language", "en"),
            created=created,
            modified=modified,
            keywords=data.get("keywords", []),
            custom_fields=custom_fields
        )


@dataclass
class WASMPermissions:
    """WASM module permissions configuration."""
    memory_limit: int = 64 * 1024 * 1024  # 64MB default
    cpu_time_limit: int = 5000  # 5 seconds default
    allow_networking: bool = False
    allow_file_system: bool = False
    allowed_imports: List[str] = field(default_factory=lambda: ["env"])
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            "memoryLimit": self.memory_limit,
            "cpuTimeLimit": self.cpu_time_limit,
            "allowNetworking": self.allow_networking,
            "allowFileSystem": self.allow_file_system,
            "allowedImports": self.allowed_imports
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "WASMPermissions":
        """Create from dictionary."""
        return cls(
            memory_limit=data.get("memoryLimit", 64 * 1024 * 1024),
            cpu_time_limit=data.get("cpuTimeLimit", 5000),
            allow_networking=data.get("allowNetworking", False),
            allow_file_system=data.get("allowFileSystem", False),
            allowed_imports=data.get("allowedImports", ["env"])
        )


@dataclass
class JSPermissions:
    """JavaScript execution permissions."""
    execution_mode: str = "sandboxed"  # none, sandboxed, trusted
    allowed_apis: List[str] = field(default_factory=lambda: ["dom"])
    dom_access: str = "read"  # none, read, write
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            "executionMode": self.execution_mode,
            "allowedAPIs": self.allowed_apis,
            "domAccess": self.dom_access
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "JSPermissions":
        """Create from dictionary."""
        return cls(
            execution_mode=data.get("executionMode", "sandboxed"),
            allowed_apis=data.get("allowedAPIs", ["dom"]),
            dom_access=data.get("domAccess", "read")
        )


@dataclass
class NetworkPolicy:
    """Network access policy."""
    allow_outbound: bool = False
    allowed_hosts: List[str] = field(default_factory=list)
    allowed_ports: List[int] = field(default_factory=list)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            "allowOutbound": self.allow_outbound,
            "allowedHosts": self.allowed_hosts,
            "allowedPorts": self.allowed_ports
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "NetworkPolicy":
        """Create from dictionary."""
        return cls(
            allow_outbound=data.get("allowOutbound", False),
            allowed_hosts=data.get("allowedHosts", []),
            allowed_ports=data.get("allowedPorts", [])
        )


@dataclass
class StoragePolicy:
    """Storage access policy."""
    allow_local_storage: bool = False
    allow_session_storage: bool = False
    allow_indexed_db: bool = False
    allow_cookies: bool = False
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            "allowLocalStorage": self.allow_local_storage,
            "allowSessionStorage": self.allow_session_storage,
            "allowIndexedDB": self.allow_indexed_db,
            "allowCookies": self.allow_cookies
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "StoragePolicy":
        """Create from dictionary."""
        return cls(
            allow_local_storage=data.get("allowLocalStorage", False),
            allow_session_storage=data.get("allowSessionStorage", False),
            allow_indexed_db=data.get("allowIndexedDB", False),
            allow_cookies=data.get("allowCookies", False)
        )


@dataclass
class SecurityPolicy:
    """Complete security policy for a LIV document."""
    wasm_permissions: WASMPermissions = field(default_factory=WASMPermissions)
    js_permissions: JSPermissions = field(default_factory=JSPermissions)
    network_policy: NetworkPolicy = field(default_factory=NetworkPolicy)
    storage_policy: StoragePolicy = field(default_factory=StoragePolicy)
    content_security_policy: Optional[str] = None
    trusted_domains: List[str] = field(default_factory=list)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            "wasmPermissions": self.wasm_permissions.to_dict(),
            "jsPermissions": self.js_permissions.to_dict(),
            "networkPolicy": self.network_policy.to_dict(),
            "storagePolicy": self.storage_policy.to_dict(),
            "contentSecurityPolicy": self.content_security_policy,
            "trustedDomains": self.trusted_domains
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "SecurityPolicy":
        """Create from dictionary."""
        return cls(
            wasm_permissions=WASMPermissions.from_dict(data.get("wasmPermissions", {})),
            js_permissions=JSPermissions.from_dict(data.get("jsPermissions", {})),
            network_policy=NetworkPolicy.from_dict(data.get("networkPolicy", {})),
            storage_policy=StoragePolicy.from_dict(data.get("storagePolicy", {})),
            content_security_policy=data.get("contentSecurityPolicy"),
            trusted_domains=data.get("trustedDomains", [])
        )


@dataclass
class AssetInfo:
    """Information about a document asset."""
    name: str
    asset_type: str  # image, font, data, audio, video
    path: Optional[Path] = None
    data: Optional[bytes] = None
    mime_type: Optional[str] = None
    size: Optional[int] = None
    hash: Optional[str] = None
    
    def __post_init__(self):
        if self.data and self.size is None:
            self.size = len(self.data)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            "name": self.name,
            "type": self.asset_type,
            "path": str(self.path) if self.path else None,
            "mimeType": self.mime_type,
            "size": self.size,
            "hash": self.hash
        }


@dataclass
class WASMModuleInfo:
    """Information about a WASM module."""
    name: str
    path: Optional[Path] = None
    data: Optional[bytes] = None
    version: str = "1.0"
    entry_point: str = "main"
    permissions: Optional[WASMPermissions] = None
    exports: List[str] = field(default_factory=list)
    imports: List[str] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)
    
    def __post_init__(self):
        if self.permissions is None:
            self.permissions = WASMPermissions()
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            "name": self.name,
            "path": str(self.path) if self.path else None,
            "version": self.version,
            "entryPoint": self.entry_point,
            "permissions": self.permissions.to_dict() if self.permissions else None,
            "exports": self.exports,
            "imports": self.imports,
            "metadata": self.metadata
        }


@dataclass
class ValidationResult:
    """Result of document validation."""
    is_valid: bool
    errors: List[str] = field(default_factory=list)
    warnings: List[str] = field(default_factory=list)
    validation_time: Optional[float] = None
    file_path: Optional[Path] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            "isValid": self.is_valid,
            "errors": self.errors,
            "warnings": self.warnings,
            "validationTime": self.validation_time,
            "filePath": str(self.file_path) if self.file_path else None
        }


@dataclass
class ConversionResult:
    """Result of format conversion."""
    success: bool
    input_path: Path
    output_path: Optional[Path] = None
    source_format: Optional[str] = None
    target_format: Optional[str] = None
    conversion_time: Optional[float] = None
    file_size_before: Optional[int] = None
    file_size_after: Optional[int] = None
    errors: List[str] = field(default_factory=list)
    warnings: List[str] = field(default_factory=list)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            "success": self.success,
            "inputPath": str(self.input_path),
            "outputPath": str(self.output_path) if self.output_path else None,
            "sourceFormat": self.source_format,
            "targetFormat": self.target_format,
            "conversionTime": self.conversion_time,
            "fileSizeBefore": self.file_size_before,
            "fileSizeAfter": self.file_size_after,
            "errors": self.errors,
            "warnings": self.warnings
        }


@dataclass
class BatchProcessingResult:
    """Result of batch processing operation."""
    total_files: int
    successful: int
    failed: int
    results: List[ConversionResult] = field(default_factory=list)
    processing_time: Optional[float] = None
    
    @property
    def success_rate(self) -> float:
        """Calculate success rate as percentage."""
        if self.total_files == 0:
            return 0.0
        return (self.successful / self.total_files) * 100.0
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            "totalFiles": self.total_files,
            "successful": self.successful,
            "failed": self.failed,
            "successRate": self.success_rate,
            "processingTime": self.processing_time,
            "results": [result.to_dict() for result in self.results]
        }


@dataclass
class FeatureFlags:
    """Feature flags for document capabilities."""
    animations: bool = False
    interactivity: bool = False
    charts: bool = False
    forms: bool = False
    audio: bool = False
    video: bool = False
    webgl: bool = False
    webassembly: bool = False
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            "animations": self.animations,
            "interactivity": self.interactivity,
            "charts": self.charts,
            "forms": self.forms,
            "audio": self.audio,
            "video": self.video,
            "webgl": self.webgl,
            "webassembly": self.webassembly
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "FeatureFlags":
        """Create from dictionary."""
        return cls(
            animations=data.get("animations", False),
            interactivity=data.get("interactivity", False),
            charts=data.get("charts", False),
            forms=data.get("forms", False),
            audio=data.get("audio", False),
            video=data.get("video", False),
            webgl=data.get("webgl", False),
            webassembly=data.get("webassembly", False)
        )


# Export all models
__all__ = [
    "DocumentMetadata",
    "WASMPermissions",
    "JSPermissions",
    "NetworkPolicy",
    "StoragePolicy",
    "SecurityPolicy",
    "AssetInfo",
    "WASMModuleInfo",
    "ValidationResult",
    "ConversionResult",
    "BatchProcessingResult",
    "FeatureFlags",
]