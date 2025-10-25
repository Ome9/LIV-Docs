"""
Asset management utilities for the LIV Python SDK
"""

import hashlib
import mimetypes
import shutil
from pathlib import Path
from typing import Dict, List, Optional, Union, Any, BinaryIO
import tempfile
import zipfile

from .models import AssetInfo
from .exceptions import AssetError, LIVError


class AssetManager:
    """Manages assets for LIV documents."""
    
    def __init__(self, temp_dir: Optional[Union[str, Path]] = None):
        """
        Initialize asset manager.
        
        Args:
            temp_dir: Temporary directory for asset processing
        """
        self.temp_dir = Path(temp_dir) if temp_dir else Path(tempfile.gettempdir()) / "liv_assets"
        self.temp_dir.mkdir(exist_ok=True)
        
        # Asset registry
        self.assets: Dict[str, AssetInfo] = {}
    
    def add_asset_from_file(self, name: str, file_path: Union[str, Path],
                           asset_type: Optional[str] = None,
                           mime_type: Optional[str] = None) -> AssetInfo:
        """
        Add an asset from a file.
        
        Args:
            name: Asset name
            file_path: Path to asset file
            asset_type: Asset type (auto-detected if not provided)
            mime_type: MIME type (auto-detected if not provided)
            
        Returns:
            AssetInfo object
        """
        file_path = Path(file_path)
        
        if not file_path.exists():
            raise AssetError(f"Asset file not found: {file_path}", asset_path=str(file_path))
        
        # Auto-detect asset type if not provided
        if not asset_type:
            asset_type = self._detect_asset_type(file_path)
        
        # Auto-detect MIME type if not provided
        if not mime_type:
            mime_type, _ = mimetypes.guess_type(str(file_path))
        
        # Read file data
        with open(file_path, 'rb') as f:
            data = f.read()
        
        # Calculate hash
        file_hash = self._calculate_hash(data)
        
        # Create asset info
        asset_info = AssetInfo(
            name=name,
            asset_type=asset_type,
            path=file_path,
            data=data,
            mime_type=mime_type,
            size=len(data),
            hash=file_hash
        )
        
        self.assets[name] = asset_info
        return asset_info
    
    def add_asset_from_data(self, name: str, data: bytes, asset_type: str,
                           mime_type: Optional[str] = None) -> AssetInfo:
        """
        Add an asset from raw data.
        
        Args:
            name: Asset name
            data: Asset data
            asset_type: Asset type
            mime_type: MIME type
            
        Returns:
            AssetInfo object
        """
        # Calculate hash
        file_hash = self._calculate_hash(data)
        
        # Create asset info
        asset_info = AssetInfo(
            name=name,
            asset_type=asset_type,
            data=data,
            mime_type=mime_type,
            size=len(data),
            hash=file_hash
        )
        
        self.assets[name] = asset_info
        return asset_info
    
    def add_asset_from_url(self, name: str, url: str, asset_type: Optional[str] = None) -> AssetInfo:
        """
        Add an asset by downloading from URL.
        
        Args:
            name: Asset name
            url: Asset URL
            asset_type: Asset type (auto-detected if not provided)
            
        Returns:
            AssetInfo object
        """
        try:
            import urllib.request
            
            # Download the asset
            with urllib.request.urlopen(url) as response:
                data = response.read()
                content_type = response.headers.get('Content-Type')
            
            # Auto-detect asset type if not provided
            if not asset_type:
                if content_type:
                    asset_type = self._mime_type_to_asset_type(content_type)
                else:
                    # Try to detect from URL extension
                    url_path = Path(url)
                    asset_type = self._detect_asset_type(url_path)
            
            return self.add_asset_from_data(name, data, asset_type, content_type)
            
        except Exception as e:
            raise AssetError(f"Failed to download asset from URL: {e}", asset_path=url)
    
    def _detect_asset_type(self, file_path: Path) -> str:
        """Detect asset type from file extension."""
        extension = file_path.suffix.lower()
        
        image_extensions = {'.jpg', '.jpeg', '.png', '.gif', '.bmp', '.svg', '.webp', '.ico'}
        font_extensions = {'.ttf', '.otf', '.woff', '.woff2', '.eot'}
        audio_extensions = {'.mp3', '.wav', '.ogg', '.m4a', '.aac', '.flac'}
        video_extensions = {'.mp4', '.webm', '.ogg', '.avi', '.mov', '.wmv'}
        
        if extension in image_extensions:
            return 'image'
        elif extension in font_extensions:
            return 'font'
        elif extension in audio_extensions:
            return 'audio'
        elif extension in video_extensions:
            return 'video'
        else:
            return 'data'
    
    def _mime_type_to_asset_type(self, mime_type: str) -> str:
        """Convert MIME type to asset type."""
        if mime_type.startswith('image/'):
            return 'image'
        elif mime_type.startswith('font/') or 'font' in mime_type:
            return 'font'
        elif mime_type.startswith('audio/'):
            return 'audio'
        elif mime_type.startswith('video/'):
            return 'video'
        else:
            return 'data'
    
    def _calculate_hash(self, data: bytes) -> str:
        """Calculate SHA-256 hash of data."""
        return hashlib.sha256(data).hexdigest()
    
    def get_asset(self, name: str) -> Optional[AssetInfo]:
        """Get asset by name."""
        return self.assets.get(name)
    
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
    
    def remove_asset(self, name: str) -> bool:
        """
        Remove an asset.
        
        Args:
            name: Asset name
            
        Returns:
            True if asset was removed
        """
        if name in self.assets:
            del self.assets[name]
            return True
        return False
    
    def clear_assets(self) -> None:
        """Clear all assets."""
        self.assets.clear()
    
    def get_total_size(self) -> int:
        """Get total size of all assets in bytes."""
        return sum(asset.size or 0 for asset in self.assets.values())
    
    def get_assets_by_type(self) -> Dict[str, List[AssetInfo]]:
        """Group assets by type."""
        grouped = {}
        for asset in self.assets.values():
            if asset.asset_type not in grouped:
                grouped[asset.asset_type] = []
            grouped[asset.asset_type].append(asset)
        return grouped
    
    def optimize_images(self, quality: int = 85, max_width: Optional[int] = None,
                       max_height: Optional[int] = None) -> Dict[str, int]:
        """
        Optimize image assets.
        
        Args:
            quality: JPEG quality (1-100)
            max_width: Maximum width for resizing
            max_height: Maximum height for resizing
            
        Returns:
            Dictionary mapping asset names to size reduction in bytes
        """
        try:
            from PIL import Image
            import io
        except ImportError:
            raise AssetError("PIL (Pillow) is required for image optimization")
        
        optimizations = {}
        
        for name, asset in self.assets.items():
            if asset.asset_type == 'image' and asset.data:
                try:
                    # Open image
                    image = Image.open(io.BytesIO(asset.data))
                    
                    # Resize if needed
                    if max_width or max_height:
                        image.thumbnail((max_width or image.width, max_height or image.height), Image.Resampling.LANCZOS)
                    
                    # Save optimized image
                    output = io.BytesIO()
                    
                    # Convert to RGB if necessary for JPEG
                    if image.mode in ('RGBA', 'LA', 'P'):
                        if asset.mime_type == 'image/jpeg':
                            # Convert to RGB for JPEG
                            rgb_image = Image.new('RGB', image.size, (255, 255, 255))
                            rgb_image.paste(image, mask=image.split()[-1] if image.mode == 'RGBA' else None)
                            image = rgb_image
                    
                    # Determine format
                    format_map = {
                        'image/jpeg': 'JPEG',
                        'image/png': 'PNG',
                        'image/webp': 'WEBP'
                    }
                    
                    format_name = format_map.get(asset.mime_type, 'JPEG')
                    
                    # Save with optimization
                    if format_name == 'JPEG':
                        image.save(output, format=format_name, quality=quality, optimize=True)
                    else:
                        image.save(output, format=format_name, optimize=True)
                    
                    # Update asset data
                    original_size = len(asset.data)
                    optimized_data = output.getvalue()
                    new_size = len(optimized_data)
                    
                    if new_size < original_size:
                        asset.data = optimized_data
                        asset.size = new_size
                        asset.hash = self._calculate_hash(optimized_data)
                        optimizations[name] = original_size - new_size
                    
                except Exception as e:
                    # Skip assets that can't be optimized
                    continue
        
        return optimizations
    
    def compress_assets(self, compression_level: int = 6) -> Dict[str, int]:
        """
        Compress assets using gzip.
        
        Args:
            compression_level: Compression level (1-9)
            
        Returns:
            Dictionary mapping asset names to size reduction in bytes
        """
        import gzip
        
        compressions = {}
        
        for name, asset in self.assets.items():
            if asset.data and asset.asset_type in ['data', 'font']:  # Only compress certain types
                try:
                    original_size = len(asset.data)
                    compressed_data = gzip.compress(asset.data, compresslevel=compression_level)
                    new_size = len(compressed_data)
                    
                    # Only use compression if it actually reduces size
                    if new_size < original_size:
                        asset.data = compressed_data
                        asset.size = new_size
                        asset.hash = self._calculate_hash(compressed_data)
                        # Update MIME type to indicate compression
                        if asset.mime_type:
                            asset.mime_type = f"{asset.mime_type}+gzip"
                        compressions[name] = original_size - new_size
                        
                except Exception:
                    # Skip assets that can't be compressed
                    continue
        
        return compressions
    
    def export_assets(self, output_dir: Union[str, Path], 
                     preserve_structure: bool = True) -> Dict[str, Path]:
        """
        Export all assets to a directory.
        
        Args:
            output_dir: Output directory
            preserve_structure: Whether to preserve asset type subdirectories
            
        Returns:
            Dictionary mapping asset names to exported file paths
        """
        output_dir = Path(output_dir)
        output_dir.mkdir(parents=True, exist_ok=True)
        
        exported = {}
        
        for name, asset in self.assets.items():
            if asset.data:
                if preserve_structure:
                    # Create subdirectory for asset type
                    type_dir = output_dir / f"{asset.asset_type}s"
                    type_dir.mkdir(exist_ok=True)
                    file_path = type_dir / name
                else:
                    file_path = output_dir / name
                
                # Write asset data
                with open(file_path, 'wb') as f:
                    f.write(asset.data)
                
                exported[name] = file_path
        
        return exported
    
    def import_from_directory(self, input_dir: Union[str, Path],
                            pattern: str = "*", recursive: bool = True) -> List[AssetInfo]:
        """
        Import assets from a directory.
        
        Args:
            input_dir: Input directory
            pattern: File pattern to match
            recursive: Search recursively
            
        Returns:
            List of imported asset info objects
        """
        input_dir = Path(input_dir)
        
        if not input_dir.exists():
            raise AssetError(f"Input directory not found: {input_dir}")
        
        imported = []
        
        # Find files
        if recursive:
            files = list(input_dir.rglob(pattern))
        else:
            files = list(input_dir.glob(pattern))
        
        for file_path in files:
            if file_path.is_file():
                try:
                    asset_info = self.add_asset_from_file(file_path.name, file_path)
                    imported.append(asset_info)
                except Exception as e:
                    # Skip files that can't be imported
                    continue
        
        return imported
    
    def create_asset_archive(self, output_path: Union[str, Path]) -> Path:
        """
        Create a ZIP archive of all assets.
        
        Args:
            output_path: Output ZIP file path
            
        Returns:
            Path to created archive
        """
        output_path = Path(output_path)
        
        with zipfile.ZipFile(output_path, 'w', zipfile.ZIP_DEFLATED) as zip_file:
            for name, asset in self.assets.items():
                if asset.data:
                    # Create path within archive
                    archive_path = f"{asset.asset_type}s/{name}"
                    zip_file.writestr(archive_path, asset.data)
        
        return output_path
    
    def load_from_archive(self, archive_path: Union[str, Path]) -> List[AssetInfo]:
        """
        Load assets from a ZIP archive.
        
        Args:
            archive_path: Path to ZIP archive
            
        Returns:
            List of loaded asset info objects
        """
        archive_path = Path(archive_path)
        
        if not archive_path.exists():
            raise AssetError(f"Archive not found: {archive_path}")
        
        loaded = []
        
        with zipfile.ZipFile(archive_path, 'r') as zip_file:
            for file_info in zip_file.filelist:
                if not file_info.is_dir():
                    # Extract asset type and name from path
                    path_parts = file_info.filename.split('/')
                    if len(path_parts) >= 2:
                        asset_type = path_parts[0].rstrip('s')  # Remove plural
                        name = '/'.join(path_parts[1:])
                        
                        # Read asset data
                        data = zip_file.read(file_info.filename)
                        
                        # Create asset info
                        asset_info = self.add_asset_from_data(name, data, asset_type)
                        loaded.append(asset_info)
        
        return loaded
    
    def get_statistics(self) -> Dict[str, Any]:
        """Get asset statistics."""
        stats = {
            'total_assets': len(self.assets),
            'total_size': self.get_total_size(),
            'by_type': {}
        }
        
        # Group by type
        for asset_type, assets in self.get_assets_by_type().items():
            stats['by_type'][asset_type] = {
                'count': len(assets),
                'size': sum(asset.size or 0 for asset in assets)
            }
        
        return stats
    
    def cleanup_temp_files(self) -> None:
        """Clean up temporary files."""
        if self.temp_dir.exists():
            shutil.rmtree(self.temp_dir)
            self.temp_dir.mkdir(exist_ok=True)
    
    def __del__(self):
        """Cleanup when object is destroyed."""
        try:
            self.cleanup_temp_files()
        except:
            pass


__all__ = [
    "AssetManager",
]