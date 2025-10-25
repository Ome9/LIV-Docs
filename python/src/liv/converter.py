"""
Format conversion utilities for the LIV Python SDK
"""

from pathlib import Path
from typing import Union, Optional, Dict, Any, List
import time

from .models import ConversionResult
from .cli_interface import CLIInterface
from .exceptions import ConversionError, LIVError
from .config_manager import ConfigManager


class LIVConverter:
    """Handles conversion between LIV and other document formats."""
    
    def __init__(self, config_manager: Optional[ConfigManager] = None):
        """
        Initialize converter.
        
        Args:
            config_manager: Configuration manager instance
        """
        self.config_manager = config_manager or ConfigManager()
        self.cli = CLIInterface(config_manager=self.config_manager)
        
        # Get conversion configuration
        self.conversion_config = self.config_manager.get_conversion_config()
    
    def liv_to_pdf(self, input_path: Union[str, Path], output_path: Union[str, Path],
                   quality: Optional[str] = None, include_assets: Optional[bool] = None,
                   page_size: str = "A4", orientation: str = "portrait") -> ConversionResult:
        """
        Convert LIV document to PDF.
        
        Args:
            input_path: Input .liv file path
            output_path: Output PDF file path
            quality: PDF quality (low, medium, high)
            include_assets: Whether to include assets
            page_size: Page size (A4, Letter, etc.)
            orientation: Page orientation (portrait, landscape)
            
        Returns:
            ConversionResult object
        """
        options = {}
        
        # Use config defaults if not specified
        if quality is None:
            quality = self.conversion_config.get("pdf_quality", "high")
        options["quality"] = quality
        
        if include_assets is None:
            include_assets = self.conversion_config.get("pdf_include_assets", True)
        if include_assets:
            options["include-assets"] = True
        
        options["page-size"] = page_size
        options["orientation"] = orientation
        
        return self._convert_with_cli(input_path, output_path, "pdf", options)
    
    def liv_to_html(self, input_path: Union[str, Path], output_path: Union[str, Path],
                    include_assets: Optional[bool] = None, 
                    standalone: bool = True,
                    include_css: bool = True,
                    include_js: bool = True) -> ConversionResult:
        """
        Convert LIV document to HTML.
        
        Args:
            input_path: Input .liv file path
            output_path: Output HTML file path
            include_assets: Whether to include assets
            standalone: Create standalone HTML file
            include_css: Include CSS in output
            include_js: Include JavaScript in output
            
        Returns:
            ConversionResult object
        """
        options = {}
        
        if include_assets is None:
            include_assets = self.conversion_config.get("html_include_assets", True)
        if include_assets:
            options["include-assets"] = True
        
        if standalone:
            options["standalone"] = True
        
        if not include_css:
            options["no-css"] = True
        
        if not include_js:
            options["no-js"] = True
        
        return self._convert_with_cli(input_path, output_path, "html", options)
    
    def liv_to_markdown(self, input_path: Union[str, Path], output_path: Union[str, Path],
                       preserve_formatting: Optional[bool] = None,
                       include_images: bool = True) -> ConversionResult:
        """
        Convert LIV document to Markdown.
        
        Args:
            input_path: Input .liv file path
            output_path: Output Markdown file path
            preserve_formatting: Whether to preserve formatting
            include_images: Whether to include images
            
        Returns:
            ConversionResult object
        """
        options = {}
        
        if preserve_formatting is None:
            preserve_formatting = self.conversion_config.get("markdown_preserve_formatting", True)
        if preserve_formatting:
            options["preserve-formatting"] = True
        
        if include_images:
            options["include-images"] = True
        
        return self._convert_with_cli(input_path, output_path, "markdown", options)
    
    def liv_to_epub(self, input_path: Union[str, Path], output_path: Union[str, Path],
                    title: Optional[str] = None, author: Optional[str] = None,
                    cover_image: Optional[Union[str, Path]] = None) -> ConversionResult:
        """
        Convert LIV document to EPUB.
        
        Args:
            input_path: Input .liv file path
            output_path: Output EPUB file path
            title: EPUB title (uses document title if not specified)
            author: EPUB author (uses document author if not specified)
            cover_image: Path to cover image
            
        Returns:
            ConversionResult object
        """
        options = {}
        
        if title:
            options["title"] = title
        
        if author:
            options["author"] = author
        
        if cover_image:
            options["cover"] = str(cover_image)
        
        return self._convert_with_cli(input_path, output_path, "epub", options)
    
    def html_to_liv(self, input_path: Union[str, Path], output_path: Union[str, Path],
                   title: Optional[str] = None, author: Optional[str] = None,
                   css_file: Optional[Union[str, Path]] = None,
                   js_file: Optional[Union[str, Path]] = None,
                   assets_dir: Optional[Union[str, Path]] = None) -> ConversionResult:
        """
        Convert HTML file to LIV document.
        
        Args:
            input_path: Input HTML file path
            output_path: Output .liv file path
            title: Document title
            author: Document author
            css_file: Optional CSS file to include
            js_file: Optional JavaScript file to include
            assets_dir: Directory containing assets to include
            
        Returns:
            ConversionResult object
        """
        options = {}
        
        if title:
            options["title"] = title
        
        if author:
            options["author"] = author
        
        if css_file:
            options["css"] = str(css_file)
        
        if js_file:
            options["js"] = str(js_file)
        
        if assets_dir:
            options["assets"] = str(assets_dir)
        
        return self._convert_with_cli(input_path, output_path, "liv", options)
    
    def markdown_to_liv(self, input_path: Union[str, Path], output_path: Union[str, Path],
                       title: Optional[str] = None, author: Optional[str] = None,
                       css_file: Optional[Union[str, Path]] = None,
                       template: Optional[str] = None) -> ConversionResult:
        """
        Convert Markdown file to LIV document.
        
        Args:
            input_path: Input Markdown file path
            output_path: Output .liv file path
            title: Document title
            author: Document author
            css_file: Optional CSS file for styling
            template: HTML template to use
            
        Returns:
            ConversionResult object
        """
        options = {}
        
        if title:
            options["title"] = title
        
        if author:
            options["author"] = author
        
        if css_file:
            options["css"] = str(css_file)
        
        if template:
            options["template"] = template
        
        return self._convert_with_cli(input_path, output_path, "liv", options)
    
    def pdf_to_liv(self, input_path: Union[str, Path], output_path: Union[str, Path],
                  title: Optional[str] = None, author: Optional[str] = None,
                  extract_images: bool = True) -> ConversionResult:
        """
        Convert PDF file to LIV document.
        
        Args:
            input_path: Input PDF file path
            output_path: Output .liv file path
            title: Document title
            author: Document author
            extract_images: Whether to extract images from PDF
            
        Returns:
            ConversionResult object
        """
        options = {}
        
        if title:
            options["title"] = title
        
        if author:
            options["author"] = author
        
        if extract_images:
            options["extract-images"] = True
        
        return self._convert_with_cli(input_path, output_path, "liv", options)
    
    def epub_to_liv(self, input_path: Union[str, Path], output_path: Union[str, Path],
                   extract_metadata: bool = True) -> ConversionResult:
        """
        Convert EPUB file to LIV document.
        
        Args:
            input_path: Input EPUB file path
            output_path: Output .liv file path
            extract_metadata: Whether to extract EPUB metadata
            
        Returns:
            ConversionResult object
        """
        options = {}
        
        if extract_metadata:
            options["extract-metadata"] = True
        
        return self._convert_with_cli(input_path, output_path, "liv", options)
    
    def _convert_with_cli(self, input_path: Union[str, Path], output_path: Union[str, Path],
                         target_format: str, options: Dict[str, Any]) -> ConversionResult:
        """
        Perform conversion using CLI interface.
        
        Args:
            input_path: Input file path
            output_path: Output file path
            target_format: Target format
            options: Conversion options
            
        Returns:
            ConversionResult object
        """
        try:
            return self.cli.convert(input_path, output_path, target_format, options)
        except Exception as e:
            # Create error result
            return ConversionResult(
                success=False,
                input_path=Path(input_path),
                output_path=None,
                source_format=Path(input_path).suffix.lstrip('.'),
                target_format=target_format,
                errors=[str(e)]
            )
    
    def get_supported_formats(self) -> Dict[str, List[str]]:
        """
        Get supported conversion formats.
        
        Returns:
            Dictionary mapping source formats to supported target formats
        """
        try:
            formats = self.cli.list_formats()
            
            # Parse format information (this would depend on CLI output format)
            # For now, return known supported formats
            return {
                "liv": ["pdf", "html", "markdown", "epub"],
                "html": ["liv"],
                "markdown": ["liv"],
                "pdf": ["liv"],
                "epub": ["liv"]
            }
        except Exception:
            # Return default formats if CLI query fails
            return {
                "liv": ["pdf", "html", "markdown", "epub"],
                "html": ["liv"],
                "markdown": ["liv"],
                "pdf": ["liv"],
                "epub": ["liv"]
            }
    
    def is_conversion_supported(self, source_format: str, target_format: str) -> bool:
        """
        Check if conversion between formats is supported.
        
        Args:
            source_format: Source format
            target_format: Target format
            
        Returns:
            True if conversion is supported
        """
        supported = self.get_supported_formats()
        return target_format in supported.get(source_format, [])
    
    def convert_auto(self, input_path: Union[str, Path], output_path: Union[str, Path],
                    **options) -> ConversionResult:
        """
        Auto-detect formats and convert.
        
        Args:
            input_path: Input file path
            output_path: Output file path
            **options: Conversion options
            
        Returns:
            ConversionResult object
        """
        input_path = Path(input_path)
        output_path = Path(output_path)
        
        source_format = input_path.suffix.lstrip('.').lower()
        target_format = output_path.suffix.lstrip('.').lower()
        
        if not self.is_conversion_supported(source_format, target_format):
            raise ConversionError(
                f"Conversion from {source_format} to {target_format} is not supported",
                source_format=source_format,
                target_format=target_format
            )
        
        # Route to appropriate conversion method
        if source_format == "liv":
            if target_format == "pdf":
                return self.liv_to_pdf(input_path, output_path, **options)
            elif target_format == "html":
                return self.liv_to_html(input_path, output_path, **options)
            elif target_format == "markdown" or target_format == "md":
                return self.liv_to_markdown(input_path, output_path, **options)
            elif target_format == "epub":
                return self.liv_to_epub(input_path, output_path, **options)
        
        elif target_format == "liv":
            if source_format == "html":
                return self.html_to_liv(input_path, output_path, **options)
            elif source_format == "markdown" or source_format == "md":
                return self.markdown_to_liv(input_path, output_path, **options)
            elif source_format == "pdf":
                return self.pdf_to_liv(input_path, output_path, **options)
            elif source_format == "epub":
                return self.epub_to_liv(input_path, output_path, **options)
        
        # Fallback to CLI conversion
        return self._convert_with_cli(input_path, output_path, target_format, options)


# Utility functions
def quick_convert(input_path: Union[str, Path], output_path: Union[str, Path]) -> bool:
    """Quick conversion with auto-detection."""
    converter = LIVConverter()
    result = converter.convert_auto(input_path, output_path)
    return result.success


def convert_to_pdf(input_path: Union[str, Path], output_path: Union[str, Path]) -> bool:
    """Quick LIV to PDF conversion."""
    converter = LIVConverter()
    result = converter.liv_to_pdf(input_path, output_path)
    return result.success


def convert_from_html(input_path: Union[str, Path], output_path: Union[str, Path],
                     title: str = "Converted Document", author: str = "Unknown") -> bool:
    """Quick HTML to LIV conversion."""
    converter = LIVConverter()
    result = converter.html_to_liv(input_path, output_path, title=title, author=author)
    return result.success


__all__ = [
    "LIVConverter",
    "quick_convert",
    "convert_to_pdf",
    "convert_from_html",
]