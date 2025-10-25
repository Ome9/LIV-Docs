# LIV Security Permission Management

The LIV Security Permission Management system provides a comprehensive interface for managing granular permissions, inheritance, and security policies for LIV documents and WASM modules.

## Features

### 🔍 Permission Evaluation
- Real-time permission evaluation against security policies
- Support for WASM module permissions (memory, CPU, networking, filesystem)
- Automatic inheritance from parent policies
- Detailed security warnings and restrictions

### 📋 Permission Templates
- Pre-configured permission templates for common use cases
- Basic Document, Interactive Content, Data Visualization, and Network Enabled templates
- Custom template creation and management
- Template-based policy generation

### 🛡️ Security Policy Management
- Hierarchical policy inheritance
- Administrative controls (document size limits, WASM module limits, file type restrictions)
- Compliance settings (GDPR, HIPAA, data classification)
- Resource limits and monitoring

### 🔗 Trust Chain Validation
- Digital signature verification
- Certificate authority validation
- Trust chain building and verification
- Revocation status checking

## Getting Started

### Starting the Permission Management Server

```bash
# Basic usage
go run cmd/permission-server/main.go

# With custom configuration
go run cmd/permission-server/main.go \
  -port 8080 \
  -config-dir ./security-config \
  -log-level info

# With TLS enabled
go run cmd/permission-server/main.go \
  -port 8443 \
  -tls \
  -cert server.crt \
  -key server.key
```

### Accessing the Web Interface

Once the server is running, open your browser and navigate to:
- HTTP: `http://localhost:8080`
- HTTPS: `https://localhost:8443`

## Web Interface Guide

### Permission Evaluation Panel

1. **Document Information**
   - Enter the Document ID for the LIV document
   - Optionally specify a WASM Module Name for module-specific evaluation

2. **Security Policy Selection**
   - Choose from available security policies
   - Policies are loaded automatically from the policy manager

3. **Permission Configuration**
   - **Memory Limit**: Set WASM memory limit in bytes (1KB - 128MB)
   - **CPU Time Limit**: Set CPU execution time limit in milliseconds (100ms - 30s)
   - **Allow Networking**: Enable/disable network access
   - **Allow File System**: Enable/disable file system access
   - **Allowed Imports**: Comma-separated list of allowed WASM imports

4. **Justification**
   - Provide a reason for the permission request
   - Used for audit logging and compliance

5. **Evaluation Results**
   - ✅ **Granted**: Permissions approved
   - ❌ **Denied**: Permissions rejected
   - **Warnings**: Security concerns and recommendations
   - **Restrictions**: Applied limitations and constraints
   - **Inheritance**: Information about inherited permissions

### Permission Templates Panel

Pre-configured templates for common scenarios:

#### Basic Document Template
- **Memory**: 4MB
- **CPU Time**: 2 seconds
- **Network**: Disabled
- **File System**: Disabled
- **Use Case**: Static documents with minimal interactivity

#### Interactive Content Template
- **Memory**: 16MB
- **CPU Time**: 10 seconds
- **Network**: Disabled
- **File System**: Disabled
- **Imports**: console, dom, events
- **Use Case**: Interactive forms, games, and dynamic content

#### Data Visualization Template
- **Memory**: 32MB
- **CPU Time**: 15 seconds
- **Network**: Disabled
- **File System**: Disabled
- **Imports**: console, dom, canvas, webgl
- **Use Case**: Charts, graphs, and complex data visualizations

#### Network Enabled Template
- **Memory**: 16MB
- **CPU Time**: 10 seconds
- **Network**: Enabled
- **File System**: Disabled
- **Imports**: console, dom, fetch
- **Use Case**: Documents that fetch external data or communicate with APIs

### Security Policies Panel

View and select from available security policies:

- **Basic Security Policy**: Conservative settings for basic documents
- **High Security Policy**: Strict settings for sensitive documents
- **Interactive Content Policy**: Balanced settings for interactive content

Each policy shows:
- Policy name and description
- Creation date and author
- Parent policy (if inherited)

### Trust Chain Validation Panel

Validate document signatures and trust chains:

1. Enter the Document ID
2. Click "Validate Trust Chain"
3. View the complete trust chain with:
   - Certificate Authority information
   - Validity periods
   - Trust levels (system, organization, user)
   - Revocation status

## API Reference

### Evaluate Permissions

```http
POST /api/permissions/evaluate
Content-Type: application/json

{
  "document_id": "doc-123",
  "module_name": "chart-module",
  "policy_id": "basic-security",
  "requested_permissions": {
    "memory_limit": 16777216,
    "cpu_time_limit": 5000,
    "allow_networking": false,
    "allow_file_system": false,
    "allowed_imports": ["console", "dom"]
  },
  "user_context": {
    "user_id": "user-123",
    "session_id": "session-456",
    "ip_address": "192.168.1.100",
    "roles": ["user"]
  },
  "justification": "Interactive chart rendering",
  "requested_at": "2024-01-15T10:30:00Z"
}
```

**Response:**
```json
{
  "granted": true,
  "inherited_from": "",
  "restrictions": [
    "Memory limited to 16777216 bytes (requested 16777216)"
  ],
  "warnings": [
    {
      "type": "high_memory_usage",
      "description": "High memory usage requested: 16 MB",
      "recommendation": "Consider optimizing memory usage"
    }
  ],
  "trust_chain": [],
  "evaluated_at": "2024-01-15T10:30:01Z",
  "expires_at": "2024-01-15T11:30:01Z"
}
```

### Get Permission Templates

```http
GET /api/permissions/templates
```

**Response:**
```json
[
  {
    "id": "basic-document",
    "name": "Basic Document",
    "description": "Basic permissions for simple document rendering",
    "category": "document",
    "permissions": {
      "memory_limit": 4194304,
      "cpu_time_limit": 2000,
      "allow_networking": false,
      "allow_file_system": false,
      "allowed_imports": ["console"]
    },
    "restrictions": ["No network access", "No file system access"],
    "use_case": "Static documents with minimal interactivity"
  }
]
```

### Get Security Policies

```http
GET /api/permissions/policies
```

### Validate Trust Chain

```http
GET /api/permissions/trust-chain?document_id=doc-123
```

## Security Considerations

### Permission Inheritance

The system supports hierarchical permission inheritance:

1. **Child Policy**: Evaluated first
2. **Parent Policy**: Evaluated if child denies
3. **Grandparent Policy**: Evaluated recursively up the chain
4. **Default Policy**: Final fallback

### Security Warnings

The system generates warnings for potentially risky permissions:

- **High Memory Usage**: > 32MB memory allocation
- **Long CPU Time**: > 10 seconds execution time
- **Network Access**: Any network connectivity
- **File System Access**: Any file system operations

### Restrictions

Automatic restrictions are applied when requested permissions exceed policy limits:

- Memory limits enforced at runtime
- CPU time limits with automatic termination
- Network access blocked for unauthorized domains
- File system access restricted to allowed paths
- Import restrictions prevent unauthorized module loading

### Audit Logging

All permission evaluations are logged for compliance:

- User context and session information
- Requested permissions and justification
- Evaluation results and applied restrictions
- Security warnings and policy violations
- Trust chain validation results

## Configuration

### Policy Manager Configuration

```go
config := &PolicyManagerConfig{
    DefaultPolicyID:         "default",
    EnablePolicyInheritance: true,
    MaxPolicyDepth:          5,
    EnableVersioning:        true,
    AuditLogPath:           "./audit.log",
    EventLogPath:           "./security-events.log",
}
```

### Server Configuration

```bash
# Environment variables
export LIV_PERMISSION_PORT=8080
export LIV_PERMISSION_CONFIG_DIR=./security-config
export LIV_PERMISSION_LOG_LEVEL=info
export LIV_PERMISSION_ENABLE_TLS=false
```

## Integration Examples

### Go Integration

```go
// Create permission manager
permManager := security.NewPermissionManager(
    policyManager, 
    securityManager, 
    cryptoProvider, 
    logger,
)

// Evaluate permissions
request := &security.PermissionRequest{
    DocumentID: "doc-123",
    RequestedPerms: &core.WASMPermissions{
        MemoryLimit: 16 * 1024 * 1024,
        CPUTimeLimit: 5000,
        AllowNetworking: false,
        AllowFileSystem: false,
        AllowedImports: []string{"console", "dom"},
    },
    PolicyID: "basic-security",
    UserContext: userContext,
    Justification: "Interactive content rendering",
}

evaluation, err := permManager.EvaluatePermissionRequest(ctx, request)
if err != nil {
    log.Fatal(err)
}

if evaluation.Granted {
    // Proceed with document processing
} else {
    // Handle permission denial
}
```

### JavaScript Integration

```javascript
// Evaluate permissions via API
async function evaluatePermissions(documentId, permissions) {
    const response = await fetch('/api/permissions/evaluate', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            document_id: documentId,
            requested_permissions: permissions,
            policy_id: 'basic-security',
            user_context: {
                user_id: 'current-user',
                session_id: 'current-session'
            },
            justification: 'Document rendering'
        })
    });
    
    const evaluation = await response.json();
    return evaluation;
}
```

## Troubleshooting

### Common Issues

1. **Permission Denied**
   - Check policy limits against requested permissions
   - Verify user context and roles
   - Review inheritance chain for conflicts

2. **Trust Chain Validation Failed**
   - Ensure document is properly signed
   - Check certificate validity and expiration
   - Verify certificate authority is trusted

3. **High Memory/CPU Warnings**
   - Optimize WASM module resource usage
   - Consider breaking work into smaller chunks
   - Review algorithm efficiency

### Debug Mode

Enable debug logging for detailed information:

```bash
go run cmd/permission-server/main.go -log-level debug
```

### Log Files

Check log files for detailed information:
- `security-events.log`: Security events and violations
- `audit.log`: Permission evaluations and policy changes

## Best Practices

1. **Use Templates**: Start with permission templates for common use cases
2. **Principle of Least Privilege**: Grant minimum necessary permissions
3. **Regular Reviews**: Periodically review and update security policies
4. **Monitor Warnings**: Address security warnings promptly
5. **Audit Compliance**: Maintain audit logs for compliance requirements
6. **Test Inheritance**: Verify permission inheritance works as expected
7. **Validate Trust Chains**: Ensure proper signature verification