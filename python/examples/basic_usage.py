#!/usr/bin/env python3
"""
Basic usage examples for the LIV Python SDK
"""

import json
from pathlib import Path
from liv import LIVBuilder, LIVHelpers, LIVBatchProcessor, DocumentMetadata, SecurityPolicy


def example_simple_document():
    """Create a simple text document."""
    print("Creating a simple document...")
    
    # Method 1: Using the builder
    builder = LIVBuilder()
    document = (builder
               .set_metadata(
                   title="My First LIV Document",
                   author="Python SDK User",
                   description="A simple document created with the Python SDK"
               )
               .set_content(
                   html="""
                   <h1>Welcome to LIV</h1>
                   <p>This is my first document created with the LIV Python SDK.</p>
                   <p>It includes some basic HTML content and styling.</p>
                   """,
                   css="""
                   body {
                       font-family: Arial, sans-serif;
                       max-width: 800px;
                       margin: 0 auto;
                       padding: 20px;
                       line-height: 1.6;
                   }
                   h1 {
                       color: #2c3e50;
                       border-bottom: 2px solid #3498db;
                       padding-bottom: 10px;
                   }
                   """
               )
               .build())
    
    print(f"Document created: {document.metadata.title}")
    print(f"Size: {document.get_size_info()['total_size']} bytes")
    
    return document


def example_helper_functions():
    """Demonstrate helper functions."""
    print("\nUsing helper functions...")
    
    # Method 2: Using helper functions
    text_doc = LIVHelpers.createTextDocument(
        title="Quick Text Document",
        content="This document was created using a helper function.\n\nIt's much simpler for basic text documents.",
        author="Helper Function"
    )
    
    print(f"Helper document created: {text_doc.metadata.title}")
    
    return text_doc


def example_interactive_document():
    """Create an interactive document with JavaScript."""
    print("\nCreating interactive document...")
    
    builder = LIVBuilder()
    document = (builder
               .set_metadata(
                   title="Interactive Counter",
                   author="SDK Demo",
                   description="A simple interactive counter using JavaScript"
               )
               .set_content(
                   html="""
                   <div class="container">
                       <h1>Interactive Counter</h1>
                       <div class="counter">
                           <button id="decrease">-</button>
                           <span id="count">0</span>
                           <button id="increase">+</button>
                       </div>
                       <p>Click the buttons to change the counter value.</p>
                   </div>
                   """,
                   css="""
                   .container {
                       text-align: center;
                       padding: 40px;
                       font-family: Arial, sans-serif;
                   }
                   .counter {
                       margin: 30px 0;
                       font-size: 24px;
                   }
                   button {
                       font-size: 24px;
                       padding: 10px 20px;
                       margin: 0 10px;
                       border: none;
                       background: #3498db;
                       color: white;
                       border-radius: 5px;
                       cursor: pointer;
                   }
                   button:hover {
                       background: #2980b9;
                   }
                   #count {
                       display: inline-block;
                       min-width: 50px;
                       font-weight: bold;
                   }
                   """,
                   js="""
                   let count = 0;
                   
                   function updateDisplay() {
                       document.getElementById('count').textContent = count;
                   }
                   
                   document.getElementById('increase').addEventListener('click', function() {
                       count++;
                       updateDisplay();
                   });
                   
                   document.getElementById('decrease').addEventListener('click', function() {
                       count--;
                       updateDisplay();
                   });
                   
                   // Initialize
                   updateDisplay();
                   """
               )
               .enable_features(interactivity=True)
               .build())
    
    print(f"Interactive document created: {document.metadata.title}")
    print(f"Features enabled: {[name for name, enabled in document.feature_flags.to_dict().items() if enabled]}")
    
    return document


def example_document_with_assets():
    """Create a document with assets."""
    print("\nCreating document with assets...")
    
    builder = LIVBuilder()
    
    # Add some data assets
    config_data = {
        "theme": "dark",
        "version": "1.0",
        "features": ["animations", "interactivity"]
    }
    
    user_data = [
        {"name": "Alice", "score": 95},
        {"name": "Bob", "score": 87},
        {"name": "Charlie", "score": 92}
    ]
    
    document = (builder
               .set_metadata(
                   title="Document with Assets",
                   author="Asset Demo",
                   description="Demonstrates adding various assets to a document"
               )
               .set_content(
                   html="""
                   <h1>Document with Assets</h1>
                   <p>This document includes several data assets:</p>
                   <ul>
                       <li>Configuration data</li>
                       <li>User scores data</li>
                       <li>Custom styling</li>
                   </ul>
                   <div id="data-display"></div>
                   """,
                   css="""
                   body { font-family: Arial, sans-serif; padding: 20px; }
                   #data-display { 
                       background: #f0f0f0; 
                       padding: 15px; 
                       border-radius: 5px; 
                       margin-top: 20px;
                   }
                   """,
                   js="""
                   // This would load and display the asset data
                   document.addEventListener('DOMContentLoaded', function() {
                       const display = document.getElementById('data-display');
                       display.innerHTML = '<p>Asset data would be loaded and displayed here.</p>';
                   });
                   """
               )
               .add_data("config.json", config_data)
               .add_data("users.json", user_data)
               .enable_features(interactivity=True)
               .build())
    
    print(f"Document with assets created: {document.metadata.title}")
    print(f"Assets: {list(document.assets.keys())}")
    
    return document


def example_security_policy():
    """Demonstrate custom security policy."""
    print("\nCreating document with custom security policy...")
    
    # Create a restrictive security policy
    security_policy = SecurityPolicy()
    security_policy.wasm_permissions.memory_limit = 32 * 1024 * 1024  # 32MB
    security_policy.wasm_permissions.allow_networking = False
    security_policy.js_permissions.execution_mode = "sandboxed"
    security_policy.js_permissions.dom_access = "read"
    
    builder = LIVBuilder()
    document = (builder
               .set_metadata(
                   title="Secure Document",
                   author="Security Demo",
                   description="Document with restrictive security policy"
               )
               .set_content(
                   html="<h1>Secure Document</h1><p>This document has restricted permissions.</p>"
               )
               .set_security_policy(security_policy)
               .build())
    
    print(f"Secure document created: {document.metadata.title}")
    print(f"WASM memory limit: {document.security_policy.wasm_permissions.memory_limit} bytes")
    print(f"JS execution mode: {document.security_policy.js_permissions.execution_mode}")
    
    return document


def example_batch_processing():
    """Demonstrate batch processing capabilities."""
    print("\nDemonstrating batch processing...")
    
    # Create multiple documents for batch processing
    documents = []
    
    for i in range(3):
        builder = LIVBuilder()
        doc = (builder
               .set_metadata(
                   title=f"Batch Document {i+1}",
                   author="Batch Processor",
                   description=f"Document {i+1} created for batch processing demo"
               )
               .set_content(
                   html=f"<h1>Document {i+1}</h1><p>This is batch document number {i+1}.</p>"
               )
               .build())
        documents.append(doc)
    
    print(f"Created {len(documents)} documents for batch processing")
    
    # Note: Actual batch processing would require CLI tools
    # This is just demonstrating the document creation part
    
    return documents


def example_document_info():
    """Show how to get document information."""
    print("\nGetting document information...")
    
    # Create a sample document
    builder = LIVBuilder()
    document = (builder
               .set_metadata(
                   title="Info Demo Document",
                   author="Info Demo",
                   keywords=["demo", "information", "metadata"]
               )
               .set_content(
                   html="<h1>Document Info</h1>" * 10,  # Some content
                   css="h1 { color: blue; }" * 5  # Some CSS
               )
               .add_data("sample.json", {"key": "value"})
               .enable_features(animations=True, interactivity=True)
               .build())
    
    # Get size information
    size_info = document.get_size_info()
    print(f"Document: {document.metadata.title}")
    print(f"Total size: {size_info['total_size']:,} bytes")
    print(f"Content size: {size_info['content_size']:,} bytes")
    print(f"Assets size: {size_info['assets_size']:,} bytes")
    
    # Get metadata
    print(f"Author: {document.metadata.author}")
    print(f"Keywords: {', '.join(document.metadata.keywords)}")
    
    # Get features
    enabled_features = [name for name, enabled in document.feature_flags.to_dict().items() if enabled]
    print(f"Enabled features: {', '.join(enabled_features)}")
    
    return document


def main():
    """Run all examples."""
    print("LIV Python SDK - Basic Usage Examples")
    print("=" * 50)
    
    try:
        # Run examples
        example_simple_document()
        example_helper_functions()
        example_interactive_document()
        example_document_with_assets()
        example_security_policy()
        example_batch_processing()
        example_document_info()
        
        print("\n" + "=" * 50)
        print("All examples completed successfully!")
        print("\nNote: To actually save documents as .liv files, you need the LIV CLI tools installed.")
        print("The Python SDK provides the high-level interface, but uses the Go CLI for file operations.")
        
    except Exception as e:
        print(f"\nError running examples: {e}")
        print("This is expected if the LIV CLI tools are not installed.")


if __name__ == "__main__":
    main()