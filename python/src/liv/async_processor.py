"""
Async processing utilities for the LIV Python SDK
"""

import asyncio
import aiofiles
from pathlib import Path
from typing import List, Union, Optional, Dict, Any, Callable, Awaitable
import time

from .models import ConversionResult, BatchProcessingResult
from .converter import LIVConverter
from .validator import LIVValidator
from .exceptions import LIVError, ConversionError
from .config_manager import ConfigManager


class AsyncLIVProcessor:
    """Async processor for LIV documents."""
    
    def __init__(self, config_manager: Optional[ConfigManager] = None):
        """
        Initialize async processor.
        
        Args:
            config_manager: Configuration manager instance
        """
        self.config_manager = config_manager or ConfigManager()
        self.converter = LIVConverter(config_manager=self.config_manager)
        self.validator = LIVValidator(config_manager=self.config_manager)
        
        # Get batch processing configuration
        batch_config = self.config_manager.get_batch_config()
        self.max_concurrent = batch_config.get("max_concurrent", 4)
        self.timeout = batch_config.get("timeout", 300)
        self.retry_attempts = batch_config.get("retry_attempts", 3)
    
    async def convert_html_to_liv_async(self, html_path: Union[str, Path], 
                                       output_path: Union[str, Path],
                                       title: Optional[str] = None,
                                       author: Optional[str] = None) -> ConversionResult:
        """
        Async HTML to LIV conversion.
        
        Args:
            html_path: Input HTML file path
            output_path: Output LIV file path
            title: Document title
            author: Document author
            
        Returns:
            ConversionResult object
        """
        # Run conversion in thread pool to avoid blocking
        loop = asyncio.get_event_loop()
        
        def sync_convert():
            return self.converter.html_to_liv(
                html_path, output_path, 
                title=title, author=author
            )
        
        return await loop.run_in_executor(None, sync_convert)
    
    async def convert_liv_to_pdf_async(self, liv_path: Union[str, Path],
                                      pdf_path: Union[str, Path],
                                      quality: str = "high") -> ConversionResult:
        """
        Async LIV to PDF conversion.
        
        Args:
            liv_path: Input LIV file path
            pdf_path: Output PDF file path
            quality: PDF quality
            
        Returns:
            ConversionResult object
        """
        loop = asyncio.get_event_loop()
        
        def sync_convert():
            return self.converter.liv_to_pdf(liv_path, pdf_path, quality=quality)
        
        return await loop.run_in_executor(None, sync_convert)
    
    async def validate_async(self, file_path: Union[str, Path],
                           strict: bool = False) -> Any:
        """
        Async document validation.
        
        Args:
            file_path: Path to LIV file
            strict: Use strict validation
            
        Returns:
            ValidationResult object
        """
        loop = asyncio.get_event_loop()
        
        def sync_validate():
            return self.validator.validate_file(file_path, strict=strict)
        
        return await loop.run_in_executor(None, sync_validate)
    
    async def convert_multiple_async(self, conversions: List[Dict[str, Any]],
                                   progress_callback: Optional[Callable[[int, int], Awaitable[None]]] = None) -> BatchProcessingResult:
        """
        Convert multiple files asynchronously.
        
        Args:
            conversions: List of conversion specifications
            progress_callback: Optional async progress callback
            
        Returns:
            BatchProcessingResult object
        """
        start_time = time.time()
        results = []
        successful = 0
        failed = 0
        
        # Create semaphore to limit concurrent operations
        semaphore = asyncio.Semaphore(self.max_concurrent)
        
        async def convert_single(conversion: Dict[str, Any]) -> ConversionResult:
            """Convert a single file with semaphore."""
            async with semaphore:
                return await self._convert_single_async_with_retry(conversion)
        
        # Create tasks for all conversions
        tasks = [convert_single(conv) for conv in conversions]
        
        # Process tasks as they complete
        for i, task in enumerate(asyncio.as_completed(tasks)):
            try:
                result = await task
                results.append(result)
                
                if result.success:
                    successful += 1
                else:
                    failed += 1
                
                # Call progress callback if provided
                if progress_callback:
                    await progress_callback(len(results), len(conversions))
                    
            except Exception as e:
                # Create error result
                error_result = ConversionResult(
                    success=False,
                    input_path=Path(conversions[i]['input_path']),
                    errors=[f"Conversion error: {e}"]
                )
                results.append(error_result)
                failed += 1
        
        processing_time = time.time() - start_time
        
        return BatchProcessingResult(
            total_files=len(conversions),
            successful=successful,
            failed=failed,
            results=results,
            processing_time=processing_time
        )
    
    async def _convert_single_async_with_retry(self, conversion: Dict[str, Any]) -> ConversionResult:
        """Convert a single file with retry logic (async)."""
        input_path = conversion['input_path']
        output_path = conversion['output_path']
        target_format = conversion.get('target_format')
        options = conversion.get('options', {})
        
        last_error = None
        
        for attempt in range(self.retry_attempts):
            try:
                # Run conversion in thread pool
                loop = asyncio.get_event_loop()
                
                def sync_convert():
                    if target_format:
                        return self.converter._convert_with_cli(input_path, output_path, target_format, options)
                    else:
                        return self.converter.convert_auto(input_path, output_path, **options)
                
                return await loop.run_in_executor(None, sync_convert)
                
            except Exception as e:
                last_error = e
                if attempt < self.retry_attempts - 1:
                    await asyncio.sleep(1)  # Brief delay before retry
        
        # All attempts failed
        return ConversionResult(
            success=False,
            input_path=Path(input_path),
            errors=[f"All {self.retry_attempts} attempts failed: {last_error}"]
        )
    
    async def validate_multiple_async(self, file_paths: List[Union[str, Path]],
                                    strict: bool = False,
                                    progress_callback: Optional[Callable[[int, int], Awaitable[None]]] = None) -> List[Any]:
        """
        Validate multiple files asynchronously.
        
        Args:
            file_paths: List of file paths to validate
            strict: Use strict validation
            progress_callback: Optional async progress callback
            
        Returns:
            List of ValidationResult objects
        """
        results = []
        
        # Create semaphore to limit concurrent operations
        semaphore = asyncio.Semaphore(self.max_concurrent)
        
        async def validate_single(file_path: Union[str, Path]) -> Any:
            """Validate a single file with semaphore."""
            async with semaphore:
                return await self.validate_async(file_path, strict=strict)
        
        # Create tasks for all validations
        tasks = [validate_single(path) for path in file_paths]
        
        # Process tasks as they complete
        for i, task in enumerate(asyncio.as_completed(tasks)):
            try:
                result = await task
                results.append(result)
                
                # Call progress callback if provided
                if progress_callback:
                    await progress_callback(len(results), len(file_paths))
                    
            except Exception as e:
                # Create error result
                from .models import ValidationResult
                error_result = ValidationResult(
                    is_valid=False,
                    errors=[f"Validation error: {e}"],
                    file_path=Path(file_paths[i])
                )
                results.append(error_result)
        
        return results
    
    async def process_directory_async(self, input_dir: Union[str, Path],
                                    output_dir: Union[str, Path],
                                    target_format: str,
                                    pattern: str = "*",
                                    recursive: bool = False,
                                    progress_callback: Optional[Callable[[int, int], Awaitable[None]]] = None) -> BatchProcessingResult:
        """
        Process all files in a directory asynchronously.
        
        Args:
            input_dir: Input directory
            output_dir: Output directory
            target_format: Target format
            pattern: File pattern to match
            recursive: Search recursively
            progress_callback: Optional async progress callback
            
        Returns:
            BatchProcessingResult object
        """
        input_dir = Path(input_dir)
        output_dir = Path(output_dir)
        
        if not input_dir.exists():
            raise LIVError(f"Input directory not found: {input_dir}")
        
        # Create output directory
        output_dir.mkdir(parents=True, exist_ok=True)
        
        # Find input files
        if recursive:
            input_files = list(input_dir.rglob(pattern))
        else:
            input_files = list(input_dir.glob(pattern))
        
        if not input_files:
            return BatchProcessingResult(total_files=0, successful=0, failed=0)
        
        # Create conversion specifications
        conversions = []
        for input_file in input_files:
            rel_path = input_file.relative_to(input_dir)
            output_file = output_dir / rel_path.with_suffix(f'.{target_format}')
            
            # Ensure output subdirectory exists
            output_file.parent.mkdir(parents=True, exist_ok=True)
            
            conversions.append({
                'input_path': input_file,
                'output_path': output_file,
                'target_format': target_format
            })
        
        return await self.convert_multiple_async(conversions, progress_callback)
    
    async def read_file_async(self, file_path: Union[str, Path]) -> str:
        """
        Read file content asynchronously.
        
        Args:
            file_path: Path to file
            
        Returns:
            File content as string
        """
        async with aiofiles.open(file_path, 'r', encoding='utf-8') as f:
            return await f.read()
    
    async def write_file_async(self, file_path: Union[str, Path], content: str) -> None:
        """
        Write file content asynchronously.
        
        Args:
            file_path: Path to file
            content: Content to write
        """
        # Ensure parent directory exists
        Path(file_path).parent.mkdir(parents=True, exist_ok=True)
        
        async with aiofiles.open(file_path, 'w', encoding='utf-8') as f:
            await f.write(content)
    
    async def create_documents_from_templates_async(self, template_data: List[Dict[str, Any]],
                                                  output_dir: Union[str, Path],
                                                  template_func: Callable[[Dict[str, Any]], str],
                                                  progress_callback: Optional[Callable[[int, int], Awaitable[None]]] = None) -> BatchProcessingResult:
        """
        Create documents from templates asynchronously.
        
        Args:
            template_data: List of data for template processing
            output_dir: Output directory
            template_func: Function that takes data and returns HTML content
            progress_callback: Optional async progress callback
            
        Returns:
            BatchProcessingResult object
        """
        output_dir = Path(output_dir)
        output_dir.mkdir(parents=True, exist_ok=True)
        
        start_time = time.time()
        results = []
        successful = 0
        failed = 0
        
        # Create semaphore to limit concurrent operations
        semaphore = asyncio.Semaphore(self.max_concurrent)
        
        async def process_template(i: int, data: Dict[str, Any]) -> ConversionResult:
            """Process a single template with semaphore."""
            async with semaphore:
                try:
                    # Generate content using template function
                    loop = asyncio.get_event_loop()
                    html_content = await loop.run_in_executor(None, template_func, data)
                    
                    # Create output file
                    output_file = output_dir / f"document_{i+1}.html"
                    await self.write_file_async(output_file, html_content)
                    
                    return ConversionResult(
                        success=True,
                        input_path=Path(f"template_data_{i+1}"),
                        output_path=output_file,
                        source_format="template",
                        target_format="html"
                    )
                    
                except Exception as e:
                    return ConversionResult(
                        success=False,
                        input_path=Path(f"template_data_{i+1}"),
                        errors=[f"Template processing error: {e}"]
                    )
        
        # Create tasks for all templates
        tasks = [process_template(i, data) for i, data in enumerate(template_data)]
        
        # Process tasks as they complete
        for i, task in enumerate(asyncio.as_completed(tasks)):
            try:
                result = await task
                results.append(result)
                
                if result.success:
                    successful += 1
                else:
                    failed += 1
                
                # Call progress callback if provided
                if progress_callback:
                    await progress_callback(len(results), len(template_data))
                    
            except Exception as e:
                error_result = ConversionResult(
                    success=False,
                    input_path=Path(f"template_data_{i+1}"),
                    errors=[f"Task error: {e}"]
                )
                results.append(error_result)
                failed += 1
        
        processing_time = time.time() - start_time
        
        return BatchProcessingResult(
            total_files=len(template_data),
            successful=successful,
            failed=failed,
            results=results,
            processing_time=processing_time
        )


# Utility functions for async operations
async def async_batch_convert(input_files: List[Union[str, Path]], 
                             output_dir: Union[str, Path],
                             target_format: str) -> BatchProcessingResult:
    """Async batch conversion utility."""
    processor = AsyncLIVProcessor()
    
    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    
    conversions = []
    for input_file in input_files:
        input_path = Path(input_file)
        output_path = output_dir / input_path.with_suffix(f'.{target_format}').name
        conversions.append({
            'input_path': input_path,
            'output_path': output_path,
            'target_format': target_format
        })
    
    return await processor.convert_multiple_async(conversions)


async def async_batch_validate(file_paths: List[Union[str, Path]]) -> List[Any]:
    """Async batch validation utility."""
    processor = AsyncLIVProcessor()
    return await processor.validate_multiple_async(file_paths)


__all__ = [
    "AsyncLIVProcessor",
    "async_batch_convert",
    "async_batch_validate",
]