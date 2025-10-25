"""
LIV Document class for the Python SDK
"""

import json
import zipfile
import hashlib
from pathlib import Path
from typing import Dict, List, Optional, Any, Union, BinaryIO
import tempfile
import shutil

from .exceptions import LIVError, ValidationError, AssetError
from .models import DocumentMetadata, SecurityPolicy, AssetInfo, WASMModuleInfo, FeatureFlags
from .cli_interface import CLIInterface
from .validator import LIVValidator


class LIVDocument:
    """Represents a LIV document with all its components."""
    
    def __init__(self, file_path: Optional[Union[str, Path]] = None):
        """
        Initialize LIV document.
        
        Args:
            file_path: Optional path to existing .liv file to load
        """
        self.file_path = Path(file_path) if file_path else None
        self.metadata: Optional[DocumentMetadata] = None
        self.security_policy: Optional[SecurityPolicy] = None
        self.feature_flags: Optional[FeatureFlags] = None
        
        # Content
        self.html_content: str = ""
        self.css_content: str = ""
        self.js_content: str = ""
        self.static_fallback: str = ""
        
        # Assets and modules
        self.assets: Dict[str, AssetInfo] = {}
        self.wasm_modules: Dict[str, WASMModuleInfo] = {}
        
        # Internal data
        self._manifest: Optional[Dict[str, Any]] = None
        self._temp_dir: Optional[Path] = None
        
        # Load document if path provided
        if self.file_path:
            self.load()
    
    def load(self, file_path: Optional[Union[str, Path]] = None) -> None:
        """
        Load document from .liv file.
        
        Args:
            file_path: Optional path to .liv file (uses instance path if not provided)
        """
        if file_path:
            self.file_path = Path(file_path)
        
        if not self.file_path or not self.file_path.exists():
            raise LIVError(f"File not found: {self.file_path}")
        
        # Create temporary directory for extraction
        self._temp_dir = Path(tempfile.mkdtemp(prefix="liv_document_"))
        
        try:
            # Extract ZIP contents
            with zipfile.ZipFile(self.file_path, 'r') as zip_file:
                zip_file.extractall(self._temp_dir)
            
            # Load manifest
            manifest_path = self._temp_dir / "manifest.json"
            if manifest_path.exists():
                with open(manifest_path, 'r', encoding='utf-8') as f:
                    self._manifest = json.load(f)
                self._parse_manifest()
            else:
                raise LIVError("Manifest not found in document")
            
            # Load content files
            self._load_content()
            
            # Load assets
            self._load_assets()
            
            # Load WASM modules
            self._load_wasm_modules()
            
        except Exception as e:
            self._cleanup_temp()
            if isinstance(e, LIVError):
                raise
            else:
                raise LIVError(f"Failed to load document: {e}")
    
    def _parse_manifest(self) -> None:
        """Parse manifest data into structured objects."""
        if not self._manifest:
            return
        
        # Parse metadata
        metadata_data = self._manifest.get("metadata", {})
        self.metadata = DocumentMetadata.from_dict(metadata_data)
        
        # Parse security policy
        security_data = self._manifest.get("security", {})
        self.security_policy = SecurityPolicy.from_dict(security_data)
        
        # Parse feature flags
        features_data = self._manifest.get("features", {})
        self.feature_flags = FeatureFlags.from_dict(features_data)
    
    def _load_content(self) -> None:
        """Load content files from extracted directory."""
        if not self._temp_dir:
            return
        
        content_dir = self._temp_dir / "content"
        
        # Load HTML
        html_path = content_dir / "index.html"
        if html_path.exists():
            with open(html_path, 'r', encoding='utf-8') as f:
                self.html_content = f.read()
        
        # Load CSS
        css_path = content_dir / "styles" / "main.css"
        if css_path.exists():
            with open(css_path, 'r', encoding='utf-8') as f:
                self.css_content = f.read()
        
        # Load JavaScript
        js_path = content_dir / "scripts" / "main.js"
        if js_path.exists():
            with open(js_path, 'r', encoding='utf-8') as f:
                self.js_content = f.read()
        
        # Load static fallback
        fallback_path = content_dir / "static" / "fallback.html"
        if fallback_path.exists():
            with open(fallback_path, 'r', encoding='utf-8') as f:
                self.static_fallback = f.read()
    
    def _load_assets(self) -> None:
        """Load asset information from extracted directory."""
        if not self._temp_dir:
            return
        
        assets_dir = self._temp_dir / "assets"
        if not assets_dir.exists():
            return
        
        # Load different asset types
        for asset_type in ["images", "fonts", "data", "audio", "video"]:
            type_dir = assets_dir / asset_type
            if type_dir.exists():
                for asset_file in type_dir.iterdir():
                    if asset_file.is_file():
                        asset_info = AssetInfo(
                            name=asset_file.name,
                            asset_type=asset_type.rstrip('s'),  # Remove plural
                            path=asset_file,
                            size=asset_file.stat().st_size,
                            hash=self._calculate_file_hash(asset_file)
                        )
                        self.assets[asset_file.name] = asset_info
    
    def _load_wasm_modules(self) -> None:
        """Load WASM module information."""
        if not self._temp_dir:
            return
        
        # Look for WASM files in root directory
        for wasm_file in self._temp_dir.glob("*.wasm"):
            module_name = wasm_file.stem
            
            # Load module data
            with open(wasm_file, 'rb') as f:
                module_data = f.read()
            
            module_info = WASMModuleInfo(
                name=module_name,
                path=wasm_file,
                data=module_data
            )
            
            # Load module metadata from manifest if available
            if self._manifest and "wasmConfig" in self._manifest:
                wasm_config = self._manifest["wasmConfig"]
                if "modules" in wasm_config and module_name in wasm_config["modules"]:
                    module_config = wasm_config["modules"][module_name]
                    module_info.version = module_config.get("version", "1.0")
                    module_info.entry_point = module_config.get("entryPoint", "main")
                    module_info.exports = module_config.get("exports", [])
                    module_info.imports = module_config.get("imports", [])
                    module_info.metadata = module_config.get("metadata", {})
            
            self.wasm_modules[module_name] = module_info
    
    def _calculate_file_hash(self, file_path: Path) -> str:
        """Calculate SHA-256 hash of a file."""
        hash_sha256 = hashlib.sha256()
        with open(file_path, 'rb') as f:
            for chunk in iter(lambda: f.read(4096), b""):
                hash_sha256.update(chunk)
        return hash_sha256.hexdigest()
    
    def save(self, output_path: Optional[Union[str, Path]] = None, 
            sign: bool = False, key_path: Optional[Union[str, Path]] = None) -> None:
        """
        Save document to .liv file.
        
        Args:
            output_path: Output file path (uses instance path if not provided)
            sign: Whether to sign the document
            key_path: Path to signing key (required if sign=True)
        """
        if output_path:
            self.file_path = Path(output_path)
        
        if not self.file_path:
            raise LIVError("No output path specified")
        
        # Ensure parent directory exists
        self.file_path.parent.mkdir(parents=True, exist_ok=True)
        
        # Create temporary build directory
        build_dir = Path(tempfile.mkdtemp(prefix="liv_build_"))
        
        try:
            # Create directory structure
            self._create_build_structure(build_dir)
            
            # Use CLI to build the document
            cli = CLIInterface()
            cli.build(
                source_dir=build_dir,
                output_path=self.file_path,
                sign=sign,
                key_path=key_path
            )
            
        finally:
            # Cleanup build directory
            if build_dir.exists():
                shutil.rmtree(build_dir)
    
    def _create_build_structure(self, build_dir: Path) -> None:
        """Create build directory structure with all document components."""
        # Create directories
        content_dir = build_dir / "content"
        content_dir.mkdir(parents=True)
        (content_dir / "styles").mkdir(exist_ok=True)
        (content_dir / "scripts").mkdir(exist_ok=True)
        (content_dir / "static").mkdir(exist_ok=True)
        
        assets_dir = build_dir / "assets"
        for asset_type in ["images", "fonts", "data", "audio", "video"]:
            (assets_dir / asset_type).mkdir(parents=True, exist_ok=True)
        
        # Write content files
        if self.html_content:
            with open(content_dir / "index.html", 'w', encoding='utf-8') as f:
                f.write(self.html_content)
        
        if self.css_content:
            with open(content_dir / "styles" / "main.css", 'w', encoding='utf-8') as f:
                f.write(self.css_content)
        
        if self.js_content:
            with open(content_dir / "scripts" / "main.js", 'w', encoding='utf-8') as f:
                f.write(self.js_content)
        
        if self.static_fallback:
            with open(content_dir / "static" / "fallback.html", 'w', encoding='utf-8') as f:
                f.write(self.static_fallback)
        
        # Copy assets
        for asset_info in self.assets.values():
            if asset_info.path and asset_info.path.exists():
                asset_type_plural = asset_info.asset_type + 's'
                dest_path = assets_dir / asset_type_plural / asset_info.name
                shutil.copy2(asset_info.path, dest_path)
            elif asset_info.data:
                asset_type_plural = asset_info.asset_type + 's'
                dest_path = assets_dir / asset_type_plural / asset_info.name
                with open(dest_path, 'wb') as f:
                    f.write(asset_info.data)
        
        # Copy WASM modules
        for module_info in self.wasm_modules.values():
            if module_info.path and module_info.path.exists():
                dest_path = build_dir / f"{module_info.name}.wasm"
                shutil.copy2(module_info.path, dest_path)
            elif module_info.data:
                dest_path = build_dir / f"{module_info.name}.wasm"
                with open(dest_path, 'wb') as f:
                    f.write(module_info.data)
        
        # Create manifest
        self._create_manifest(build_dir)
    
    def _create_manifest(self, build_dir: Path) -> None:
        """Create manifest.json file."""
        manifest = {
            "version": "1.0",
            "metadata": self.metadata.to_dict() if self.metadata else {},
            "security": self.security_policy.to_dict() if self.security_policy else {},
            "features": self.feature_flags.to_dict() if self.feature_flags else {},
            "resources": {}
        }
        
        # Add WASM configuration if modules exist
        if self.wasm_modules:
            wasm_config = {
                "modules": {},
                "permissions": self.security_policy.wasm_permissions.to_dict() if self.security_policy else {},
                "memoryLimit": self.security_policy.wasm_permissions.memory_limit if self.security_policy else 67108864
            }
            
            for module_info in self.wasm_modules.values():
                wasm_config["modules"][module_info.name] = module_info.to_dict()
            
            manifest["wasmConfig"] = wasm_config
        
        # Write manifest
        with open(build_dir / "manifest.json", 'w', encoding='utf-8') as f:
            json.dump(manifest, f, indent=2)
    
    def validate(self) -> bool:
        """
        Validate the document.
        
        Returns:
            True if valid, False otherwise
        """
        if not self.file_path or not self.file_path.exists():
            return False
        
        validator = LIVValidator()
        result = validator.validate_file(self.file_path)
        return result.is_valid
    
    def get_asset(self, name: str) -> Optional[AssetInfo]:
        """Get asset by name."""
        return self.assets.get(name)
    
    def get_wasm_module(self, name: str) -> Optional[WASMModuleInfo]:
        """Get WASM module by name."""
        return self.wasm_modules.get(name)
    
    def list_assets(self, asset_type: Optional[str] = None) -> List[AssetInfo]:
        """
        List assets, optionally filtered by type.
        
        Args:
            asset_type: Optional asset type filter
            
        Returns:
            List of asset info objects
        """
        if asset_type:
            return [asset for asset in self.assets.values() if asset.asset_type == asset_type]
        else:
            return list(self.assets.values())
    
    def list_wasm_modules(self) -> List[WASMModuleInfo]:
        """List all WASM modules."""
        return list(self.wasm_modules.values())
    
    def get_size_info(self) -> Dict[str, int]:
        """Get size information for the document."""
        info = {
            "total_size": 0,
            "content_size": 0,
            "assets_size": 0,
            "wasm_size": 0
        }
        
        # Calculate content size
        info["content_size"] = (
            len(self.html_content.encode('utf-8')) +
            len(self.css_content.encode('utf-8')) +
            len(self.js_content.encode('utf-8')) +
            len(self.static_fallback.encode('utf-8'))
        )
        
        # Calculate assets size
        for asset in self.assets.values():
            if asset.size:
                info["assets_size"] += asset.size
        
        # Calculate WASM size
        for module in self.wasm_modules.values():
            if module.data:
                info["wasm_size"] += len(module.data)
        
        info["total_size"] = info["content_size"] + info["assets_size"] + info["wasm_size"]
        
        return info
    
    def _cleanup_temp(self) -> None:
        """Clean up temporary directory."""
        if self._temp_dir and self._temp_dir.exists():
            shutil.rmtree(self._temp_dir)
            self._temp_dir = None
    
    def __del__(self):
        """Cleanup when object is destroyed."""
        self._cleanup_temp()
    
    def __repr__(self) -> str:
        """String representation of document."""
        title = self.metadata.title if self.metadata else "Untitled"
        return f"LIVDocument(title='{title}', file_path={self.file_path})"


# Utility functions
def load_document(file_path: Union[str, Path]) -> LIVDocument:
    """Load a LIV document from file."""
    return LIVDocument(file_path)


def create_document(metadata: DocumentMetadata, html_content: str = "", 
                   css_content: str = "", js_content: str = "") -> LIVDocument:
    """Create a new LIV document with basic content."""
    doc = LIVDocument()
    doc.metadata = metadata
    doc.html_content = html_content
    doc.css_content = css_content
    doc.js_content = js_content
    return doc


__all__ = [
    "LIVDocument",
    "load_document",
    "create_document",
]