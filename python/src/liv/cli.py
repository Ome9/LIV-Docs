"""
Command-line interface for the LIV Python SDK
"""

import argparse
import sys
import json
from pathlib import Path
from typing import List, Optional

from . import __version__
from .builder import LIVBuilder
from .converter import LIVConverter
from .validator import LIVValidator
from .batch_processor import LIVBatchProcessor
from .document import LIVDocument
from .exceptions import LIVError


def create_parser() -> argparse.ArgumentParser:
    """Create the command-line argument parser."""
    parser = argparse.ArgumentParser(
        prog='liv-python',
        description='LIV Document Format Python SDK CLI',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  liv-python create --title "My Document" --author "John Doe" --html index.html --output doc.liv
  liv-python convert input.liv output.pdf
  liv-python validate document.liv
  liv-python batch-convert *.html --output-dir ./output --format liv
        """
    )
    
    parser.add_argument('--version', action='version', version=f'%(prog)s {__version__}')
    
    subparsers = parser.add_subparsers(dest='command', help='Available commands')
    
    # Create command
    create_parser = subparsers.add_parser('create', help='Create a new LIV document')
    create_parser.add_argument('--title', required=True, help='Document title')
    create_parser.add_argument('--author', required=True, help='Document author')
    create_parser.add_argument('--description', help='Document description')
    create_parser.add_argument('--html', help='HTML file to include')
    create_parser.add_argument('--css', help='CSS file to include')
    create_parser.add_argument('--js', help='JavaScript file to include')
    create_parser.add_argument('--assets-dir', help='Directory containing assets')
    create_parser.add_argument('--output', '-o', required=True, help='Output .liv file path')
    create_parser.add_argument('--sign', action='store_true', help='Sign the document')
    create_parser.add_argument('--key', help='Signing key file (required if --sign)')
    
    # Convert command
    convert_parser = subparsers.add_parser('convert', help='Convert between formats')
    convert_parser.add_argument('input', help='Input file path')
    convert_parser.add_argument('output', help='Output file path')
    convert_parser.add_argument('--format', help='Target format (auto-detected if not specified)')
    convert_parser.add_argument('--quality', help='Conversion quality (for PDF)')
    convert_parser.add_argument('--include-assets', action='store_true', help='Include assets in output')
    
    # Validate command
    validate_parser = subparsers.add_parser('validate', help='Validate LIV documents')
    validate_parser.add_argument('files', nargs='+', help='LIV files to validate')
    validate_parser.add_argument('--strict', action='store_true', help='Use strict validation')
    validate_parser.add_argument('--no-signatures', action='store_true', help='Skip signature validation')
    validate_parser.add_argument('--json', action='store_true', help='Output results as JSON')
    
    # Batch convert command
    batch_convert_parser = subparsers.add_parser('batch-convert', help='Convert multiple files')
    batch_convert_parser.add_argument('files', nargs='+', help='Input files to convert')
    batch_convert_parser.add_argument('--output-dir', '-o', required=True, help='Output directory')
    batch_convert_parser.add_argument('--format', required=True, help='Target format')
    batch_convert_parser.add_argument('--progress', action='store_true', help='Show progress')
    
    # Info command
    info_parser = subparsers.add_parser('info', help='Show document information')
    info_parser.add_argument('file', help='LIV file to analyze')
    info_parser.add_argument('--json', action='store_true', help='Output as JSON')
    
    return parser


def cmd_create(args) -> int:
    """Handle create command."""
    try:
        builder = LIVBuilder()
        
        # Set metadata
        builder.set_metadata(
            title=args.title,
            author=args.author,
            description=args.description or ""
        )
        
        # Load content files
        if args.html or args.css or args.js:
            builder.load_content_from_files(
                html_file=args.html,
                css_file=args.css,
                js_file=args.js
            )
        
        # Add assets from directory
        if args.assets_dir:
            from .asset_manager import AssetManager
            asset_manager = AssetManager()
            assets = asset_manager.import_from_directory(args.assets_dir)
            
            for asset in assets:
                builder.document.assets[asset.name] = asset
        
        # Build and save document
        document = builder.build_and_save(
            args.output,
            sign=args.sign,
            key_path=args.key
        )
        
        print(f"Created LIV document: {args.output}")
        return 0
        
    except Exception as e:
        print(f"Error creating document: {e}", file=sys.stderr)
        return 1


def cmd_convert(args) -> int:
    """Handle convert command."""
    try:
        converter = LIVConverter()
        
        # Prepare options
        options = {}
        if args.quality:
            options['quality'] = args.quality
        if args.include_assets:
            options['include_assets'] = True
        
        # Convert
        if args.format:
            result = converter._convert_with_cli(args.input, args.output, args.format, options)
        else:
            result = converter.convert_auto(args.input, args.output, **options)
        
        if result.success:
            print(f"Converted {args.input} -> {args.output}")
            if result.conversion_time:
                print(f"Conversion time: {result.conversion_time:.2f}s")
            return 0
        else:
            print(f"Conversion failed: {'; '.join(result.errors)}", file=sys.stderr)
            return 1
            
    except Exception as e:
        print(f"Error converting file: {e}", file=sys.stderr)
        return 1


def cmd_validate(args) -> int:
    """Handle validate command."""
    try:
        validator = LIVValidator()
        results = []
        
        for file_path in args.files:
            result = validator.validate_file(
                file_path,
                strict=args.strict,
                check_signatures=not args.no_signatures
            )
            results.append(result)
        
        if args.json:
            # Output as JSON
            json_results = [result.to_dict() for result in results]
            print(json.dumps(json_results, indent=2))
        else:
            # Human-readable output
            all_valid = True
            for result in results:
                status = "VALID" if result.is_valid else "INVALID"
                print(f"{result.file_path}: {status}")
                
                if result.errors:
                    for error in result.errors:
                        print(f"  ERROR: {error}")
                    all_valid = False
                
                if result.warnings:
                    for warning in result.warnings:
                        print(f"  WARNING: {warning}")
                
                if result.validation_time:
                    print(f"  Validation time: {result.validation_time:.3f}s")
        
        return 0 if all_valid else 1
        
    except Exception as e:
        print(f"Error validating files: {e}", file=sys.stderr)
        return 1


def cmd_batch_convert(args) -> int:
    """Handle batch convert command."""
    try:
        processor = LIVBatchProcessor()
        
        # Prepare conversions
        output_dir = Path(args.output_dir)
        output_dir.mkdir(parents=True, exist_ok=True)
        
        conversions = []
        for input_file in args.files:
            input_path = Path(input_file)
            output_path = output_dir / input_path.with_suffix(f'.{args.format}').name
            conversions.append({
                'input_path': input_path,
                'output_path': output_path,
                'target_format': args.format
            })
        
        # Progress callback
        def progress_callback(current: int, total: int):
            if args.progress:
                percent = (current / total) * 100
                print(f"\rProgress: {current}/{total} ({percent:.1f}%)", end='', flush=True)
        
        # Process conversions
        result = processor.convert_multiple(conversions, progress_callback)
        
        if args.progress:
            print()  # New line after progress
        
        print(f"Batch conversion completed:")
        print(f"  Total files: {result.total_files}")
        print(f"  Successful: {result.successful}")
        print(f"  Failed: {result.failed}")
        print(f"  Success rate: {result.success_rate:.1f}%")
        
        if result.processing_time:
            print(f"  Processing time: {result.processing_time:.2f}s")
        
        # Show failed files
        if result.failed > 0:
            print("\nFailed conversions:")
            for conv_result in result.results:
                if not conv_result.success:
                    print(f"  {conv_result.input_path}: {'; '.join(conv_result.errors)}")
        
        return 0 if result.failed == 0 else 1
        
    except Exception as e:
        print(f"Error in batch conversion: {e}", file=sys.stderr)
        return 1


def cmd_info(args) -> int:
    """Handle info command."""
    try:
        document = LIVDocument(args.file)
        
        if args.json:
            # JSON output
            info = {
                'metadata': document.metadata.to_dict() if document.metadata else {},
                'size_info': document.get_size_info(),
                'assets': len(document.assets),
                'wasm_modules': len(document.wasm_modules),
                'features': document.feature_flags.to_dict() if document.feature_flags else {}
            }
            print(json.dumps(info, indent=2))
        else:
            # Human-readable output
            print(f"LIV Document: {args.file}")
            print("=" * 50)
            
            if document.metadata:
                print(f"Title: {document.metadata.title}")
                print(f"Author: {document.metadata.author}")
                print(f"Description: {document.metadata.description}")
                print(f"Version: {document.metadata.version}")
                print(f"Language: {document.metadata.language}")
                print(f"Created: {document.metadata.created}")
                print(f"Modified: {document.metadata.modified}")
            
            size_info = document.get_size_info()
            print(f"\nSize Information:")
            print(f"  Total size: {size_info['total_size']:,} bytes")
            print(f"  Content size: {size_info['content_size']:,} bytes")
            print(f"  Assets size: {size_info['assets_size']:,} bytes")
            print(f"  WASM size: {size_info['wasm_size']:,} bytes")
            
            print(f"\nAssets: {len(document.assets)}")
            for asset_type, assets in document.list_assets():
                type_assets = [a for a in assets if a.asset_type == asset_type]
                if type_assets:
                    print(f"  {asset_type}: {len(type_assets)}")
            
            print(f"\nWASM Modules: {len(document.wasm_modules)}")
            for module in document.list_wasm_modules():
                print(f"  {module.name} (v{module.version})")
            
            if document.feature_flags:
                enabled_features = [name for name, enabled in document.feature_flags.to_dict().items() if enabled]
                if enabled_features:
                    print(f"\nEnabled Features: {', '.join(enabled_features)}")
        
        return 0
        
    except Exception as e:
        print(f"Error reading document info: {e}", file=sys.stderr)
        return 1


def main() -> int:
    """Main CLI entry point."""
    parser = create_parser()
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        return 1
    
    # Route to command handlers
    if args.command == 'create':
        return cmd_create(args)
    elif args.command == 'convert':
        return cmd_convert(args)
    elif args.command == 'validate':
        return cmd_validate(args)
    elif args.command == 'batch-convert':
        return cmd_batch_convert(args)
    elif args.command == 'info':
        return cmd_info(args)
    else:
        print(f"Unknown command: {args.command}", file=sys.stderr)
        return 1


if __name__ == '__main__':
    sys.exit(main())