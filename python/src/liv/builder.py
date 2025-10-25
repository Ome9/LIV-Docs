"""
LIV Document Builder for the Python SDK

Provides a fluent API for creating LIV documents programmatically.
"""

from pathlib import Path
from typing import Dict, List, Optional, Any, Union
import mimetypes
import json

from .document import LIVDocument
from .models import (
    DocumentMetadata, SecurityPolicy, WASMPermissions, JSPermissions,
    NetworkPolicy, StoragePolicy, AssetInfo, WASMModuleInfo, FeatureFlags
)
from .exceptions import LIVError, AssetError, ValidationError


class LIVBuilder:
    """Builder class for creating LIV documents with a fluent API."""
    
    def __init__(self):
        """Initialize the builder."""
        self.document = LIVDocument()
        self._reset_to_defaults()
    
    def _reset_to_defaults(self) -> None:
        """Reset document to default values."""
        # Set default metadata
        self.document.metadata = DocumentMetadata(
            title="New LIV Document",
            author="Unknown",
            description="",
            version="1.0",
            language="en"
        )
        
        # Set default security policy
        self.document.security_policy = SecurityPolicy(
            wasm_permissions=WASMPermissions(
                memory_limit=64 * 1024 * 1024,  # 64MB
                cpu_time_limit=5000,  # 5 seconds
                allow_networking=False,
                allow_file_system=False,
                allowed_imports=["env"]
            ),
            js_permissions=JSPermissions(
                execution_mode="sandboxed",
                allowed_apis=["dom"],
                dom_access="read"
            ),
            network_policy=NetworkPolicy(
                allow_outbound=False,
                allowed_hosts=[],
                allowed_ports=[]
            ),
            storage_policy=StoragePolicy(
                allow_local_storage=False,
                allow_session_storage=False,
                allow_indexed_db=False,
                allow_cookies=False
            )
        )
        
        # Set default feature flags
        self.document.feature_flags = FeatureFlags()
        
        # Initialize empty content
        self.document.html_content = ""
        self.document.css_content = ""
        self.document.js_content = ""
        self.document.static_fallback = ""
        
        # Clear assets and modules
        self.document.assets = {}
        self.document.wasm_modules = {}
    
    def set_metadata(self, title: Optional[str] = None, author: Optional[str] = None,
                    description: Optional[str] = None, version: Optional[str] = None,
                    language: Optional[str] = None, keywords: Optional[List[str]] = None,
                    **custom_fields) -> 'LIVBuilder':
        """
        Set document metadata.
        
        Args:
            title: Document title
            author: Document author
            description: Document description
            version: Document version
            language: Document language
            keywords: List of keywords
            **custom_fields: Additional custom metadata fields
            
        Returns:
            Builder instance for chaining
        """
        if not self.document.metadata:
            self.document.metadata = DocumentMetadata(title="", author="")
        
        if title is not None:
            self.document.metadata.title = title
        if author is not None:
            self.document.metadata.author = author
        if description is not None:
            self.document.metadata.description = description
        if version is not None:
            self.document.metadata.version = version
        if language is not None:
            self.document.metadata.language = language
        if keywords is not None:
            self.document.metadata.keywords = keywords
        
        # Add custom fields
        self.document.metadata.custom_fields.update(custom_fields)
        
        return self
    
    def set_content(self, html: Optional[str] = None, css: Optional[str] = None,
                   js: Optional[str] = None, static_fallback: Optional[str] = None) -> 'LIVBuilder':
        """
        Set document content.
        
        Args:
            html: HTML content
            css: CSS content
            js: JavaScript content
            static_fallback: Static fallback HTML
            
        Returns:
            Builder instance for chaining
        """
        if html is not None:
            self.document.html_content = html
        if css is not None:
            self.document.css_content = css
        if js is not None:
            self.document.js_content = js
            # Enable interactivity if JS is provided
            if js and self.document.feature_flags:
                self.document.feature_flags.interactivity = True
        if static_fallback is not None:
            self.document.static_fallback = static_fallback
        
        return self
    
    def set_html(self, html: str) -> 'LIVBuilder':
        """Set HTML content."""
        self.document.html_content = html
        return self
    
    def set_css(self, css: str) -> 'LIVBuilder':
        """Set CSS content."""
        self.document.css_content = css
        # Check for animations in CSS
        if css and self.document.feature_flags:
            if any(keyword in css.lower() for keyword in ['@keyframes', 'animation:', 'transition:']):
                self.document.feature_flags.animations = True
        return self
    
    def set_javascript(self, js: str) -> 'LIVBuilder':
        """Set JavaScript content."""
        self.document.js_content = js
        if js and self.document.feature_flags:
            self.document.feature_flags.interactivity = True
        return self
    
    def set_static_fallback(self, fallback: str) -> 'LIVBuilder':
        """Set static fallback content."""
        self.document.static_fallback = fallback
        return self
    
    def load_content_from_files(self, html_file: Optional[Union[str, Path]] = None,
                               css_file: Optional[Union[str, Path]] = None,
                               js_file: Optional[Union[str, Path]] = None) -> 'LIVBuilder':
        """
        Load content from files.
        
        Args:
            html_file: Path to HTML file
            css_file: Path to CSS file
            js_file: Path to JavaScript file
            
        Returns:
            Builder instance for chaining
        """
        if html_file:
            html_path = Path(html_file)
            if html_path.exists():
                with open(html_path, 'r', encoding='utf-8') as f:
                    self.set_html(f.read())
            else:
                raise LIVError(f"HTML file not found: {html_file}")
        
        if css_file:
            css_path = Path(css_file)
            if css_path.exists():
                with open(css_path, 'r', encoding='utf-8') as f:
                    self.set_css(f.read())
            else:
                raise LIVError(f"CSS file not found: {css_file}")
        
        if js_file:
            js_path = Path(js_file)
            if js_path.exists():
                with open(js_path, 'r', encoding='utf-8') as f:
                    self.set_javascript(f.read())
            else:
                raise LIVError(f"JavaScript file not found: {js_file}")
        
        return self
    
    def set_security_policy(self, policy: SecurityPolicy) -> 'LIVBuilder':
        """Set security policy."""
        self.document.security_policy = policy
        return self
    
    def set_wasm_permissions(self, memory_limit: Optional[int] = None,
                           cpu_time_limit: Optional[int] = None,
                           allow_networking: Optional[bool] = None,
                           allow_file_system: Optional[bool] = None,
                           allowed_imports: Optional[List[str]] = None) -> 'LIVBuilder':
        """
        Set WASM permissions.
        
        Args:
            memory_limit: Memory limit in bytes
            cpu_time_limit: CPU time limit in milliseconds
            allow_networking: Whether to allow networking
            allow_file_system: Whether to allow file system access
            allowed_imports: List of allowed imports
            
        Returns:
            Builder instance for chaining
        """
        if not self.document.security_policy:
            self.document.security_policy = SecurityPolicy()
        
        wasm_perms = self.document.security_policy.wasm_permissions
        
        if memory_limit is not None:
            wasm_perms.memory_limit = memory_limit
        if cpu_time_limit is not None:
            wasm_perms.cpu_time_limit = cpu_time_limit
        if allow_networking is not None:
            wasm_perms.allow_networking = allow_networking
        if allow_file_system is not None:
            wasm_perms.allow_file_system = allow_file_system
        if allowed_imports is not None:
            wasm_perms.allowed_imports = allowed_imports
        
        return self
    
    def set_js_permissions(self, execution_mode: Optional[str] = None,
                          allowed_apis: Optional[List[str]] = None,
                          dom_access: Optional[str] = None) -> 'LIVBuilder':
        """
        Set JavaScript permissions.
        
        Args:
            execution_mode: Execution mode (none, sandboxed, trusted)
            allowed_apis: List of allowed APIs
            dom_access: DOM access level (none, read, write)
            
        Returns:
            Builder instance for chaining
        """
        if not self.document.security_policy:
            self.document.security_policy = SecurityPolicy()
        
        js_perms = self.document.security_policy.js_permissions
        
        if execution_mode is not None:
            js_perms.execution_mode = execution_mode
        if allowed_apis is not None:
            js_perms.allowed_apis = allowed_apis
        if dom_access is not None:
            js_perms.dom_access = dom_access
        
        return self
    
    def add_asset(self, name: str, asset_type: str, 
                 file_path: Optional[Union[str, Path]] = None,
                 data: Optional[bytes] = None,
                 mime_type: Optional[str] = None) -> 'LIVBuilder':
        """
        Add an asset to the document.
        
        Args:
            name: Asset name
            asset_type: Asset type (image, font, data, audio, video)
            file_path: Path to asset file
            data: Asset data (if not loading from file)
            mime_type: MIME type (auto-detected if not provided)
            
        Returns:
            Builder instance for chaining
        """
        if file_path and data:
            raise AssetError("Cannot specify both file_path and data")
        
        if not file_path and not data:
            raise AssetError("Must specify either file_path or data")
        
        # Load from file if path provided
        if file_path:
            path = Path(file_path)
            if not path.exists():
                raise AssetError(f"Asset file not found: {file_path}", asset_path=str(file_path))
            
            with open(path, 'rb') as f:
                data = f.read()
            
            # Auto-detect MIME type if not provided
            if not mime_type:
                mime_type, _ = mimetypes.guess_type(str(path))
        
        # Create asset info
        asset_info = AssetInfo(
            name=name,
            asset_type=asset_type,
            path=Path(file_path) if file_path else None,
            data=data,
            mime_type=mime_type,
            size=len(data) if data else None
        )
        
        self.document.assets[name] = asset_info
        
        # Update feature flags based on asset type
        if self.document.feature_flags:
            if asset_type == "audio":
                self.document.feature_flags.audio = True
            elif asset_type == "video":
                self.document.feature_flags.video = True
        
        return self
    
    def add_image(self, name: str, file_path: Union[str, Path], 
                 mime_type: Optional[str] = None) -> 'LIVBuilder':
        """Add an image asset."""
        return self.add_asset(name, "image", file_path=file_path, mime_type=mime_type)
    
    def add_font(self, name: str, file_path: Union[str, Path],
                mime_type: Optional[str] = None) -> 'LIVBuilder':
        """Add a font asset."""
        return self.add_asset(name, "font", file_path=file_path, mime_type=mime_type)
    
    def add_data(self, name: str, data: Union[str, bytes, Dict[str, Any]],
                mime_type: str = "application/octet-stream") -> 'LIVBuilder':
        """
        Add a data asset.
        
        Args:
            name: Asset name
            data: Data content (string, bytes, or dict for JSON)
            mime_type: MIME type
            
        Returns:
            Builder instance for chaining
        """
        if isinstance(data, dict):
            data_bytes = json.dumps(data, indent=2).encode('utf-8')
            mime_type = "application/json"
        elif isinstance(data, str):
            data_bytes = data.encode('utf-8')
            if mime_type == "application/octet-stream":
                mime_type = "text/plain"
        else:
            data_bytes = data
        
        return self.add_asset(name, "data", data=data_bytes, mime_type=mime_type)
    
    def add_wasm_module(self, name: str, file_path: Optional[Union[str, Path]] = None,
                       data: Optional[bytes] = None, version: str = "1.0",
                       entry_point: str = "main",
                       permissions: Optional[WASMPermissions] = None) -> 'LIVBuilder':
        """
        Add a WASM module to the document.
        
        Args:
            name: Module name
            file_path: Path to WASM file
            data: WASM module data (if not loading from file)
            version: Module version
            entry_point: Entry point function name
            permissions: Module-specific permissions
            
        Returns:
            Builder instance for chaining
        """
        if file_path and data:
            raise LIVError("Cannot specify both file_path and data")
        
        if not file_path and not data:
            raise LIVError("Must specify either file_path or data")
        
        # Load from file if path provided
        if file_path:
            path = Path(file_path)
            if not path.exists():
                raise LIVError(f"WASM file not found: {file_path}")
            
            with open(path, 'rb') as f:
                data = f.read()
        
        # Create module info
        module_info = WASMModuleInfo(
            name=name,
            path=Path(file_path) if file_path else None,
            data=data,
            version=version,
            entry_point=entry_point,
            permissions=permissions
        )
        
        self.document.wasm_modules[name] = module_info
        
        # Enable WASM feature flag
        if self.document.feature_flags:
            self.document.feature_flags.webassembly = True
        
        return self
    
    def enable_features(self, animations: Optional[bool] = None,
                       interactivity: Optional[bool] = None,
                       charts: Optional[bool] = None,
                       forms: Optional[bool] = None,
                       audio: Optional[bool] = None,
                       video: Optional[bool] = None,
                       webgl: Optional[bool] = None,
                       webassembly: Optional[bool] = None) -> 'LIVBuilder':
        """
        Enable or disable document features.
        
        Args:
            animations: Enable animations
            interactivity: Enable interactivity
            charts: Enable charts
            forms: Enable forms
            audio: Enable audio
            video: Enable video
            webgl: Enable WebGL
            webassembly: Enable WebAssembly
            
        Returns:
            Builder instance for chaining
        """
        if not self.document.feature_flags:
            self.document.feature_flags = FeatureFlags()
        
        flags = self.document.feature_flags
        
        if animations is not None:
            flags.animations = animations
        if interactivity is not None:
            flags.interactivity = interactivity
        if charts is not None:
            flags.charts = charts
        if forms is not None:
            flags.forms = forms
        if audio is not None:
            flags.audio = audio
        if video is not None:
            flags.video = video
        if webgl is not None:
            flags.webgl = webgl
        if webassembly is not None:
            flags.webassembly = webassembly
        
        return self
    
    def validate(self) -> List[str]:
        """
        Validate the current document configuration.
        
        Returns:
            List of validation errors (empty if valid)
        """
        errors = []
        
        # Check required metadata
        if not self.document.metadata:
            errors.append("Document metadata is required")
        else:
            if not self.document.metadata.title:
                errors.append("Document title is required")
            if not self.document.metadata.author:
                errors.append("Document author is required")
        
        # Check content
        if not self.document.html_content and not self.document.static_fallback:
            errors.append("Document must have HTML content or static fallback")
        
        # Check WASM modules have valid data
        for name, module in self.document.wasm_modules.items():
            if not module.data and not (module.path and module.path.exists()):
                errors.append(f"WASM module '{name}' has no data or invalid path")
        
        # Check assets have valid data
        for name, asset in self.document.assets.items():
            if not asset.data and not (asset.path and asset.path.exists()):
                errors.append(f"Asset '{name}' has no data or invalid path")
        
        return errors
    
    def build(self) -> LIVDocument:
        """
        Build and return the LIV document.
        
        Returns:
            Completed LIVDocument instance
            
        Raises:
            ValidationError: If document validation fails
        """
        # Validate before building
        errors = self.validate()
        if errors:
            raise ValidationError("Document validation failed", errors=errors)
        
        return self.document
    
    def build_and_save(self, output_path: Union[str, Path],
                      sign: bool = False, key_path: Optional[Union[str, Path]] = None) -> LIVDocument:
        """
        Build document and save to file.
        
        Args:
            output_path: Output file path
            sign: Whether to sign the document
            key_path: Path to signing key
            
        Returns:
            Completed LIVDocument instance
        """
        document = self.build()
        document.save(output_path, sign=sign, key_path=key_path)
        return document
    
    def reset(self) -> 'LIVBuilder':
        """Reset builder to initial state."""
        self.document = LIVDocument()
        self._reset_to_defaults()
        return self
    
    def clone(self) -> 'LIVBuilder':
        """Create a copy of the current builder."""
        new_builder = LIVBuilder()
        
        # Copy metadata
        if self.document.metadata:
            new_builder.document.metadata = DocumentMetadata.from_dict(
                self.document.metadata.to_dict()
            )
        
        # Copy security policy
        if self.document.security_policy:
            new_builder.document.security_policy = SecurityPolicy.from_dict(
                self.document.security_policy.to_dict()
            )
        
        # Copy feature flags
        if self.document.feature_flags:
            new_builder.document.feature_flags = FeatureFlags.from_dict(
                self.document.feature_flags.to_dict()
            )
        
        # Copy content
        new_builder.document.html_content = self.document.html_content
        new_builder.document.css_content = self.document.css_content
        new_builder.document.js_content = self.document.js_content
        new_builder.document.static_fallback = self.document.static_fallback
        
        # Copy assets and modules (shallow copy)
        new_builder.document.assets = self.document.assets.copy()
        new_builder.document.wasm_modules = self.document.wasm_modules.copy()
        
        return new_builder


# Convenience functions
def create_simple_document(title: str, author: str, html_content: str,
                          css_content: str = "", js_content: str = "") -> LIVDocument:
    """Create a simple LIV document with basic content."""
    builder = LIVBuilder()
    return (builder
            .set_metadata(title=title, author=author)
            .set_content(html=html_content, css=css_content, js=js_content)
            .build())


def create_from_html_file(html_file: Union[str, Path], title: str, author: str,
                         css_file: Optional[Union[str, Path]] = None,
                         js_file: Optional[Union[str, Path]] = None) -> LIVDocument:
    """Create a LIV document from HTML file."""
    builder = LIVBuilder()
    return (builder
            .set_metadata(title=title, author=author)
            .load_content_from_files(html_file=html_file, css_file=css_file, js_file=js_file)
            .build())


__all__ = [
    "LIVBuilder",
    "create_simple_document",
    "create_from_html_file",
]