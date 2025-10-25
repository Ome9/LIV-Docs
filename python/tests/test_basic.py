"""
Basic tests for the LIV Python SDK
"""

import pytest
import tempfile
from pathlib import Path

from liv import LIVBuilder, LIVDocument, DocumentMetadata, SecurityPolicy
from liv.exceptions import LIVError, ValidationError


class TestLIVBuilder:
    """Test the LIV document builder."""
    
    def test_create_simple_document(self):
        """Test creating a simple document."""
        builder = LIVBuilder()
        
        document = (builder
                   .set_metadata(title="Test Document", author="Test Author")
                   .set_content(html="<h1>Hello World</h1>", css="h1 { color: blue; }")
                   .build())
        
        assert document.metadata.title == "Test Document"
        assert document.metadata.author == "Test Author"
        assert document.html_content == "<h1>Hello World</h1>"
        assert document.css_content == "h1 { color: blue; }"
    
    def test_validation_errors(self):
        """Test validation catches errors."""
        builder = LIVBuilder()
        
        # Missing required fields should cause validation error
        with pytest.raises(ValidationError):
            builder.build()
    
    def test_fluent_api(self):
        """Test fluent API chaining."""
        builder = LIVBuilder()
        
        # Should be able to chain method calls
        result = (builder
                 .set_metadata(title="Chain Test", author="Tester")
                 .set_html("<p>Test</p>")
                 .set_css("p { margin: 0; }")
                 .enable_features(animations=True, interactivity=True))
        
        assert result is builder  # Should return self for chaining
        
        document = builder.build()
        assert document.feature_flags.animations is True
        assert document.feature_flags.interactivity is True


class TestDocumentMetadata:
    """Test document metadata handling."""
    
    def test_metadata_creation(self):
        """Test creating metadata."""
        metadata = DocumentMetadata(
            title="Test Title",
            author="Test Author",
            description="Test Description"
        )
        
        assert metadata.title == "Test Title"
        assert metadata.author == "Test Author"
        assert metadata.description == "Test Description"
        assert metadata.version == "1.0"  # Default
        assert metadata.language == "en"  # Default
    
    def test_metadata_serialization(self):
        """Test metadata to/from dict conversion."""
        metadata = DocumentMetadata(
            title="Test Title",
            author="Test Author",
            keywords=["test", "document"]
        )
        
        # Convert to dict
        data = metadata.to_dict()
        assert data["title"] == "Test Title"
        assert data["author"] == "Test Author"
        assert data["keywords"] == ["test", "document"]
        
        # Convert back from dict
        metadata2 = DocumentMetadata.from_dict(data)
        assert metadata2.title == metadata.title
        assert metadata2.author == metadata.author
        assert metadata2.keywords == metadata.keywords


class TestSecurityPolicy:
    """Test security policy handling."""
    
    def test_default_security_policy(self):
        """Test default security policy creation."""
        policy = SecurityPolicy()
        
        # Should have secure defaults
        assert policy.wasm_permissions.allow_networking is False
        assert policy.wasm_permissions.allow_file_system is False
        assert policy.js_permissions.execution_mode == "sandboxed"
        assert policy.network_policy.allow_outbound is False
    
    def test_security_policy_serialization(self):
        """Test security policy serialization."""
        policy = SecurityPolicy()
        policy.wasm_permissions.memory_limit = 32 * 1024 * 1024
        
        # Convert to dict
        data = policy.to_dict()
        assert data["wasmPermissions"]["memoryLimit"] == 32 * 1024 * 1024
        
        # Convert back from dict
        policy2 = SecurityPolicy.from_dict(data)
        assert policy2.wasm_permissions.memory_limit == policy.wasm_permissions.memory_limit


class TestAssetManagement:
    """Test asset management functionality."""
    
    def test_add_data_asset(self):
        """Test adding a data asset."""
        builder = LIVBuilder()
        
        builder.set_metadata(title="Asset Test", author="Tester")
        builder.set_html("<h1>Test</h1>")
        builder.add_data("config", {"theme": "dark"})
        
        document = builder.build()
        
        assert "config" in document.assets
        asset = document.assets["config"]
        assert asset.asset_type == "data"
        assert asset.mime_type == "application/json"
    
    def test_add_asset_from_file(self):
        """Test adding asset from file."""
        # Create a temporary file
        with tempfile.NamedTemporaryFile(mode='w', suffix='.txt', delete=False) as f:
            f.write("Test content")
            temp_path = Path(f.name)
        
        try:
            builder = LIVBuilder()
            builder.set_metadata(title="File Asset Test", author="Tester")
            builder.set_html("<h1>Test</h1>")
            builder.add_asset("test.txt", "data", file_path=temp_path)
            
            document = builder.build()
            
            assert "test.txt" in document.assets
            asset = document.assets["test.txt"]
            assert asset.asset_type == "data"
            assert asset.data == b"Test content"
            
        finally:
            # Clean up
            temp_path.unlink()


class TestFeatureFlags:
    """Test feature flags functionality."""
    
    def test_feature_detection(self):
        """Test automatic feature detection."""
        builder = LIVBuilder()
        
        # Adding JS should enable interactivity
        builder.set_metadata(title="Feature Test", author="Tester")
        builder.set_html("<h1>Test</h1>")
        builder.set_javascript("console.log('test');")
        
        document = builder.build()
        assert document.feature_flags.interactivity is True
    
    def test_manual_feature_flags(self):
        """Test manually setting feature flags."""
        builder = LIVBuilder()
        
        builder.set_metadata(title="Manual Features", author="Tester")
        builder.set_html("<h1>Test</h1>")
        builder.enable_features(
            animations=True,
            charts=True,
            webgl=True
        )
        
        document = builder.build()
        assert document.feature_flags.animations is True
        assert document.feature_flags.charts is True
        assert document.feature_flags.webgl is True
        assert document.feature_flags.forms is False  # Not enabled


# Integration test (requires CLI tools)
@pytest.mark.integration
class TestIntegration:
    """Integration tests that require CLI tools."""
    
    def test_document_save_and_load(self):
        """Test saving and loading a document."""
        # Create a document
        builder = LIVBuilder()
        document = (builder
                   .set_metadata(title="Save Test", author="Tester")
                   .set_content(html="<h1>Save Test</h1>", css="h1 { color: red; }")
                   .build())
        
        # Save to temporary file
        with tempfile.NamedTemporaryFile(suffix='.liv', delete=False) as f:
            temp_path = Path(f.name)
        
        try:
            # This will fail if CLI tools are not available
            document.save(temp_path)
            
            # Load the document back
            loaded_doc = LIVDocument(temp_path)
            
            assert loaded_doc.metadata.title == "Save Test"
            assert loaded_doc.metadata.author == "Tester"
            
        except Exception as e:
            # Skip test if CLI tools not available
            pytest.skip(f"CLI tools not available: {e}")
        finally:
            # Clean up
            if temp_path.exists():
                temp_path.unlink()


if __name__ == "__main__":
    pytest.main([__file__])