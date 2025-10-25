"""
Batch processing utilities for the LIV Python SDK
"""

import concurrent.futures
import time
from pathlib import Path
from typing import List, Union, Optional, Dict, Any, Callable, Iterator
import logging

from .models import ConversionResult, BatchProcessingResult
from .converter import LIVConverter
from .validator import LIVValidator
from .builder import LIVBuilder
from .exceptions import LIVError, ConversionError
from .config_manager import ConfigManager

logger = logging.getLogger(__name__)


class LIVBatchProcessor:
    """Handles batch processing operations for LIV documents."""
    
    def __init__(self, config_manager: Optional[ConfigManager] = None):
        """
        Initialize batch processor.
        
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
    
    def convert_multiple(self, conversions: List[Dict[str, Any]],
                        progress_callback: Optional[Callable[[int, int], None]] = None) -> BatchProcessingResult:
        """
        Convert multiple files in batch.
        
        Args:
            conversions: List of conversion specifications, each containing:
                - input_path: Input file path
                - output_path: Output file path
                - target_format: Target format (optional, auto-detected)
                - options: Conversion options (optional)
            progress_callback: Optional callback for progress updates (current, total)
            
        Returns:
            BatchProcessingResult object
        """
        start_time = time.time()
        results = []
        successful = 0
        failed = 0
        
        logger.info(f"Starting batch conversion of {len(conversions)} files")
        
        # Process conversions with thread pool
        with concurrent.futures.ThreadPoolExecutor(max_workers=self.max_concurrent) as executor:
            # Submit all conversion tasks
            future_to_conversion = {}
            for i, conversion in enumerate(conversions):
                future = executor.submit(self._convert_single_with_retry, conversion)
                future_to_conversion[future] = (i, conversion)
            
            # Collect results as they complete
            for future in concurrent.futures.as_completed(future_to_conversion, timeout=self.timeout):
                i, conversion = future_to_conversion[future]
                
                try:
                    result = future.result()
                    results.append(result)
                    
                    if result.success:
                        successful += 1
                        logger.debug(f"Conversion {i+1} successful: {conversion['input_path']}")
                    else:
                        failed += 1
                        logger.warning(f"Conversion {i+1} failed: {conversion['input_path']}")
                    
                    # Call progress callback
                    if progress_callback:
                        progress_callback(len(results), len(conversions))
                        
                except Exception as e:
                    # Create error result
                    error_result = ConversionResult(
                        success=False,
                        input_path=Path(conversion['input_path']),
                        errors=[f"Conversion error: {e}"]
                    )
                    results.append(error_result)
                    failed += 1
                    logger.error(f"Conversion {i+1} exception: {e}")
        
        processing_time = time.time() - start_time
        
        logger.info(f"Batch conversion completed: {successful} successful, {failed} failed, {processing_time:.2f}s")
        
        return BatchProcessingResult(
            total_files=len(conversions),
            successful=successful,
            failed=failed,
            results=results,
            processing_time=processing_time
        )
    
    def _convert_single_with_retry(self, conversion: Dict[str, Any]) -> ConversionResult:
        """Convert a single file with retry logic."""
        input_path = conversion['input_path']
        output_path = conversion['output_path']
        target_format = conversion.get('target_format')
        options = conversion.get('options', {})
        
        last_error = None
        
        for attempt in range(self.retry_attempts):
            try:
                if target_format:
                    return self.converter._convert_with_cli(input_path, output_path, target_format, options)
                else:
                    return self.converter.convert_auto(input_path, output_path, **options)
                    
            except Exception as e:
                last_error = e
                if attempt < self.retry_attempts - 1:
                    logger.debug(f"Conversion attempt {attempt + 1} failed for {input_path}, retrying...")
                    time.sleep(1)  # Brief delay before retry
                else:
                    logger.error(f"All conversion attempts failed for {input_path}: {e}")
        
        # All attempts failed
        return ConversionResult(
            success=False,
            input_path=Path(input_path),
            errors=[f"All {self.retry_attempts} attempts failed: {last_error}"]
        )
    
    def convert_directory(self, input_dir: Union[str, Path], output_dir: Union[str, Path],
                         target_format: str, pattern: str = "*",
                         recursive: bool = False,
                         progress_callback: Optional[Callable[[int, int], None]] = None) -> BatchProcessingResult:
        """
        Convert all files in a directory.
        
        Args:
            input_dir: Input directory
            output_dir: Output directory
            target_format: Target format for conversion
            pattern: File pattern to match
            recursive: Search recursively
            progress_callback: Optional progress callback
            
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
            logger.warning(f"No files found matching pattern '{pattern}' in {input_dir}")
            return BatchProcessingResult(total_files=0, successful=0, failed=0)
        
        # Create conversion specifications
        conversions = []
        for input_file in input_files:
            # Calculate relative path to preserve directory structure
            rel_path = input_file.relative_to(input_dir)
            output_file = output_dir / rel_path.with_suffix(f'.{target_format}')
            
            # Ensure output subdirectory exists
            output_file.parent.mkdir(parents=True, exist_ok=True)
            
            conversions.append({
                'input_path': input_file,
                'output_path': output_file,
                'target_format': target_format
            })
        
        return self.convert_multiple(conversions, progress_callback)
    
    def validate_multiple(self, file_paths: List[Union[str, Path]],
                         strict: bool = False,
                         progress_callback: Optional[Callable[[int, int], None]] = None) -> List[Any]:
        """
        Validate multiple LIV documents in batch.
        
        Args:
            file_paths: List of file paths to validate
            strict: Use strict validation
            progress_callback: Optional progress callback
            
        Returns:
            List of ValidationResult objects
        """
        logger.info(f"Starting batch validation of {len(file_paths)} files")
        
        results = []
        
        # Process validations with thread pool
        with concurrent.futures.ThreadPoolExecutor(max_workers=self.max_concurrent) as executor:
            # Submit all validation tasks
            future_to_path = {}
            for i, file_path in enumerate(file_paths):
                future = executor.submit(self.validator.validate_file, file_path, strict)
                future_to_path[future] = (i, file_path)
            
            # Collect results as they complete
            for future in concurrent.futures.as_completed(future_to_path, timeout=self.timeout):
                i, file_path = future_to_path[future]
                
                try:
                    result = future.result()
                    results.append(result)
                    
                    if progress_callback:
                        progress_callback(len(results), len(file_paths))
                        
                except Exception as e:
                    logger.error(f"Validation failed for {file_path}: {e}")
                    # Create error result
                    from .models import ValidationResult
                    error_result = ValidationResult(
                        is_valid=False,
                        errors=[f"Validation error: {e}"],
                        file_path=Path(file_path)
                    )
                    results.append(error_result)
        
        logger.info(f"Batch validation completed")
        return results
    
    def create_documents_from_html(self, html_files: List[Union[str, Path]],
                                  output_dir: Union[str, Path],
                                  default_author: str = "Batch Processor",
                                  progress_callback: Optional[Callable[[int, int], None]] = None) -> BatchProcessingResult:
        """
        Create LIV documents from HTML files.
        
        Args:
            html_files: List of HTML file paths
            output_dir: Output directory for .liv files
            default_author: Default author for documents
            progress_callback: Optional progress callback
            
        Returns:
            BatchProcessingResult object
        """
        output_dir = Path(output_dir)
        output_dir.mkdir(parents=True, exist_ok=True)
        
        conversions = []
        
        for html_file in html_files:
            html_path = Path(html_file)
            if not html_path.exists():
                continue
            
            # Generate output path
            output_path = output_dir / html_path.with_suffix('.liv').name
            
            # Use filename as title
            title = html_path.stem.replace('_', ' ').replace('-', ' ').title()
            
            conversions.append({
                'input_path': html_path,
                'output_path': output_path,
                'target_format': 'liv',
                'options': {
                    'title': title,
                    'author': default_author
                }
            })
        
        return self.convert_multiple(conversions, progress_callback)
    
    def process_with_template(self, data_files: List[Union[str, Path]],
                             template_builder: Callable[[Dict[str, Any]], LIVBuilder],
                             output_dir: Union[str, Path],
                             progress_callback: Optional[Callable[[int, int], None]] = None) -> BatchProcessingResult:
        """
        Process data files using a template builder function.
        
        Args:
            data_files: List of data files (JSON, CSV, etc.)
            template_builder: Function that takes data and returns a LIVBuilder
            output_dir: Output directory
            progress_callback: Optional progress callback
            
        Returns:
            BatchProcessingResult object
        """
        import json
        
        output_dir = Path(output_dir)
        output_dir.mkdir(parents=True, exist_ok=True)
        
        start_time = time.time()
        results = []
        successful = 0
        failed = 0
        
        for i, data_file in enumerate(data_files):
            try:
                data_path = Path(data_file)
                
                # Load data
                if data_path.suffix.lower() == '.json':
                    with open(data_path, 'r', encoding='utf-8') as f:
                        data = json.load(f)
                else:
                    # For other formats, pass the file path
                    data = {'file_path': str(data_path)}
                
                # Create document using template
                builder = template_builder(data)
                document = builder.build()
                
                # Save document
                output_path = output_dir / data_path.with_suffix('.liv').name
                document.save(output_path)
                
                # Create success result
                result = ConversionResult(
                    success=True,
                    input_path=data_path,
                    output_path=output_path,
                    source_format=data_path.suffix.lstrip('.'),
                    target_format='liv'
                )
                results.append(result)
                successful += 1
                
            except Exception as e:
                # Create error result
                result = ConversionResult(
                    success=False,
                    input_path=Path(data_file),
                    errors=[f"Template processing error: {e}"]
                )
                results.append(result)
                failed += 1
                logger.error(f"Template processing failed for {data_file}: {e}")
            
            # Call progress callback
            if progress_callback:
                progress_callback(i + 1, len(data_files))
        
        processing_time = time.time() - start_time
        
        return BatchProcessingResult(
            total_files=len(data_files),
            successful=successful,
            failed=failed,
            results=results,
            processing_time=processing_time
        )
    
    def get_processing_stats(self) -> Dict[str, Any]:
        """Get processing statistics and configuration."""
        return {
            "max_concurrent": self.max_concurrent,
            "timeout": self.timeout,
            "retry_attempts": self.retry_attempts,
            "supported_formats": self.converter.get_supported_formats()
        }


# Utility functions
def batch_convert(input_files: List[Union[str, Path]], output_dir: Union[str, Path],
                 target_format: str) -> BatchProcessingResult:
    """Quick batch conversion utility."""
    processor = LIVBatchProcessor()
    
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
    
    return processor.convert_multiple(conversions)


def batch_validate(file_paths: List[Union[str, Path]]) -> List[Any]:
    """Quick batch validation utility."""
    processor = LIVBatchProcessor()
    return processor.validate_multiple(file_paths)


__all__ = [
    "LIVBatchProcessor",
    "batch_convert",
    "batch_validate",
]