# LIV Document Format - Deployment Guide

## Overview

This guide covers how to build, package, and deploy the complete LIV Document Format system across different platforms and environments.

## System Architecture

```
LIV Document Format System
├── Core Go Backend
│   ├── CLI Tools (liv-cli, liv-viewer, liv-builder)
│   ├── Security System (permission-server, security-admin)
│   └── Core Libraries (container, manifest, security, performance)
├── WASM Modules
│   ├── Interactive Engine (Rust)
│   └── Editor Engine (Rust)
├── JavaScript SDK
│   ├── Browser SDK
│   ├── Node.js SDK
│   └── TypeScript Definitions
├── Python SDK
│   ├── Core Library
│   ├── CLI Interface
│   └── Async Processing
├── Desktop Application
│   ├── Electron Wrapper
│   ├── Native File Associations
│   └── Cross-Platform Builds
└── Web Components
    ├── WYSIWYG Editor
    ├── Document Viewer
    └── Security Admin UI
```

## Prerequisites

### Development Environment

- **Go 1.19+**
- **Node.js 16+** with npm
- **Python 3.8+** with pip
- **Rust 1.65+** with cargo
- **Git** for version control

### Build Tools

```bash
# Install Go dependencies
go mod download

# Install Node.js dependencies
npm install -g typescript webpack webpack-cli

# Install Rust tools
rustup target add wasm32-unknown-unknown
cargo install wasm-pack

# Install Python tools
pip install build twine setuptools wheel
```

## Build Process

### 1. Complete System Build

```bash
# Clone repository
git clone https://github.com/your-org/liv-document-format.git
cd liv-document-format

# Install all dependencies
make install

# Build all components
make build

# Run comprehensive tests
make test-all

# Generate documentation
make docs
```

### 2. Component-Specific Builds

#### Go Backend and CLI Tools

```bash
# Build Go components
make build-go

# This creates:
# - bin/liv-cli (main CLI tool)
# - bin/liv-viewer (document viewer)
# - bin/liv-builder (document builder)
# - bin/permission-server (security server)
# - bin/security-admin (security administration)
```

#### WASM Modules

```bash
# Build WASM modules
make build-wasm

# This creates:
# - js/wasm/interactive/interactive_engine.wasm
# - js/wasm/interactive/interactive_engine.js
# - js/wasm/editor/editor_engine.wasm
# - js/wasm/editor/editor_engine.js
```

#### JavaScript SDK

```bash
# Build JavaScript SDK
cd js
npm install
npm run build

# This creates:
# - dist/sdk.js (main SDK)
# - dist/sdk.d.ts (TypeScript definitions)
# - dist/editor.js (WYSIWYG editor)
# - dist/viewer.js (document viewer)
```

#### Python SDK

```bash
# Build Python SDK
cd python
pip install -e .
python -m build

# This creates:
# - dist/liv_document_format-*.whl
# - dist/liv-document-format-*.tar.gz
```

#### Desktop Application

```bash
# Build desktop application
cd desktop
npm install
npm run build

# Platform-specific builds
npm run build:windows  # Creates LIV-Document-Viewer-Setup.exe
npm run build:macos    # Creates LIV-Document-Viewer.dmg
npm run build:linux    # Creates liv-document-viewer.AppImage
```

## Packaging

### 1. Create Release Packages

```bash
# Create complete release package
make release

# This creates:
# - releases/liv-document-format-v1.0.0-windows-amd64.zip
# - releases/liv-document-format-v1.0.0-darwin-amd64.zip
# - releases/liv-document-format-v1.0.0-linux-amd64.zip
```

### 2. Platform-Specific Packages

#### Windows Package

```bash
# Build Windows package
make release-windows

# Contents:
# - liv-cli.exe
# - liv-viewer.exe
# - liv-builder.exe
# - LIV-Document-Viewer-Setup.exe
# - js/dist/ (JavaScript SDK)
# - python/dist/ (Python SDK)
# - docs/ (Documentation)
# - examples/ (Example documents)
```

#### macOS Package

```bash
# Build macOS package
make release-macos

# Contents:
# - liv-cli (Universal binary)
# - liv-viewer (Universal binary)
# - liv-builder (Universal binary)
# - LIV-Document-Viewer.dmg
# - js/dist/ (JavaScript SDK)
# - python/dist/ (Python SDK)
# - docs/ (Documentation)
# - examples/ (Example documents)
```

#### Linux Package

```bash
# Build Linux package
make release-linux

# Contents:
# - liv-cli
# - liv-viewer
# - liv-builder
# - liv-document-viewer.AppImage
# - js/dist/ (JavaScript SDK)
# - python/dist/ (Python SDK)
# - docs/ (Documentation)
# - examples/ (Example documents)
```

### 3. SDK Packages

#### JavaScript SDK Package

```bash
cd js
npm pack

# Creates: liv-document-format-1.0.0.tgz
# Ready for npm publish
```

#### Python SDK Package

```bash
cd python
python -m build

# Creates:
# - dist/liv_document_format-1.0.0-py3-none-any.whl
# - dist/liv-document-format-1.0.0.tar.gz
# Ready for PyPI upload
```

## Installation Methods

### 1. Binary Installation

#### Windows

```powershell
# Download and extract release package
Invoke-WebRequest -Uri "https://github.com/your-org/liv-document-format/releases/download/v1.0.0/liv-document-format-v1.0.0-windows-amd64.zip" -OutFile "liv.zip"
Expand-Archive -Path "liv.zip" -DestinationPath "C:\Program Files\LIV"

# Add to PATH
$env:PATH += ";C:\Program Files\LIV\bin"

# Install desktop application
Start-Process "C:\Program Files\LIV\LIV-Document-Viewer-Setup.exe"
```

#### macOS

```bash
# Download and extract release package
curl -L "https://github.com/your-org/liv-document-format/releases/download/v1.0.0/liv-document-format-v1.0.0-darwin-amd64.zip" -o liv.zip
unzip liv.zip -d /usr/local/liv

# Add to PATH
echo 'export PATH="/usr/local/liv/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# Install desktop application
open /usr/local/liv/LIV-Document-Viewer.dmg
```

#### Linux

```bash
# Download and extract release package
wget "https://github.com/your-org/liv-document-format/releases/download/v1.0.0/liv-document-format-v1.0.0-linux-amd64.zip"
unzip liv-document-format-v1.0.0-linux-amd64.zip -d /opt/liv

# Add to PATH
echo 'export PATH="/opt/liv/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Make desktop application executable
chmod +x /opt/liv/liv-document-viewer.AppImage

# Create desktop entry
cat > ~/.local/share/applications/liv-document-viewer.desktop << EOF
[Desktop Entry]
Name=LIV Document Viewer
Exec=/opt/liv/liv-document-viewer.AppImage %f
Icon=/opt/liv/icons/liv-icon.png
Type=Application
MimeType=application/x-liv-document;
EOF
```

### 2. Package Manager Installation

#### Homebrew (macOS/Linux)

```bash
# Add tap
brew tap your-org/liv-document-format

# Install
brew install liv-document-format
```

#### Chocolatey (Windows)

```powershell
# Install
choco install liv-document-format
```

#### APT (Ubuntu/Debian)

```bash
# Add repository
curl -fsSL https://packages.liv-format.org/gpg | sudo apt-key add -
echo "deb https://packages.liv-format.org/apt stable main" | sudo tee /etc/apt/sources.list.d/liv.list

# Install
sudo apt update
sudo apt install liv-document-format
```

### 3. SDK Installation

#### JavaScript SDK

```bash
# Install from npm
npm install liv-document-format

# Install from GitHub
npm install https://github.com/your-org/liv-document-format/releases/download/v1.0.0/liv-document-format-1.0.0.tgz
```

#### Python SDK

```bash
# Install from PyPI
pip install liv-document-format

# Install from wheel
pip install https://github.com/your-org/liv-document-format/releases/download/v1.0.0/liv_document_format-1.0.0-py3-none-any.whl
```

## Docker Deployment

### 1. Docker Images

#### CLI Tools Image

```dockerfile
# Dockerfile.cli
FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY . .
RUN make build-go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/ ./bin/
ENV PATH="/root/bin:${PATH}"

ENTRYPOINT ["liv-cli"]
```

```bash
# Build and run
docker build -f Dockerfile.cli -t liv-cli:latest .
docker run -v $(pwd):/workspace liv-cli build --source /workspace --output /workspace/document.liv
```

#### Web Viewer Image

```dockerfile
# Dockerfile.viewer
FROM node:16-alpine AS builder

WORKDIR /app
COPY js/ ./
RUN npm install && npm run build

FROM nginx:alpine
COPY --from=builder /app/dist/ /usr/share/nginx/html/
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80
```

```bash
# Build and run
docker build -f Dockerfile.viewer -t liv-viewer:latest .
docker run -p 8080:80 liv-viewer:latest
```

### 2. Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  liv-cli:
    build:
      context: .
      dockerfile: Dockerfile.cli
    volumes:
      - ./documents:/workspace
    command: ["build", "--source", "/workspace/source", "--output", "/workspace/output/document.liv"]

  liv-viewer:
    build:
      context: .
      dockerfile: Dockerfile.viewer
    ports:
      - "8080:80"
    depends_on:
      - liv-cli

  security-server:
    build:
      context: .
      dockerfile: Dockerfile.security
    ports:
      - "8443:8443"
    environment:
      - SECURITY_LEVEL=strict
      - LOG_LEVEL=info

  performance-monitor:
    build:
      context: .
      dockerfile: Dockerfile.monitor
    ports:
      - "9090:9090"
    volumes:
      - ./logs:/var/log/liv
```

```bash
# Deploy with Docker Compose
docker-compose up -d

# Scale services
docker-compose up -d --scale liv-cli=3
```

## Cloud Deployment

### 1. AWS Deployment

#### Lambda Functions

```yaml
# serverless.yml
service: liv-document-processor

provider:
  name: aws
  runtime: go1.x
  region: us-east-1

functions:
  processDocument:
    handler: bin/lambda-processor
    events:
      - s3:
          bucket: liv-documents
          event: s3:ObjectCreated:*
    environment:
      SECURITY_LEVEL: strict

  validateDocument:
    handler: bin/lambda-validator
    events:
      - http:
          path: /validate
          method: post

resources:
  Resources:
    DocumentBucket:
      Type: AWS::S3::Bucket
      Properties:
        BucketName: liv-documents
```

#### ECS Deployment

```yaml
# ecs-task-definition.json
{
  "family": "liv-document-processor",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "containerDefinitions": [
    {
      "name": "liv-processor",
      "image": "your-account.dkr.ecr.us-east-1.amazonaws.com/liv-processor:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "SECURITY_LEVEL",
          "value": "strict"
        }
      ]
    }
  ]
}
```

### 2. Google Cloud Deployment

#### Cloud Run

```yaml
# cloudbuild.yaml
steps:
  - name: 'gcr.io/cloud-builders/docker'
    args: ['build', '-t', 'gcr.io/$PROJECT_ID/liv-processor', '.']
  - name: 'gcr.io/cloud-builders/docker'
    args: ['push', 'gcr.io/$PROJECT_ID/liv-processor']
  - name: 'gcr.io/cloud-builders/gcloud'
    args:
      - 'run'
      - 'deploy'
      - 'liv-processor'
      - '--image'
      - 'gcr.io/$PROJECT_ID/liv-processor'
      - '--region'
      - 'us-central1'
      - '--platform'
      - 'managed'
```

#### Cloud Functions

```go
// cloud-function/main.go
package main

import (
    "context"
    "github.com/GoogleCloudPlatform/functions-framework-go/functions"
    "github.com/your-org/liv-document-format/pkg/container"
)

func init() {
    functions.HTTP("ProcessDocument", processDocument)
}

func processDocument(w http.ResponseWriter, r *http.Request) {
    // Document processing logic
}
```

### 3. Azure Deployment

#### Container Instances

```yaml
# azure-container-instances.yml
apiVersion: 2019-12-01
location: eastus
name: liv-processor
properties:
  containers:
  - name: liv-processor
    properties:
      image: your-registry.azurecr.io/liv-processor:latest
      resources:
        requests:
          cpu: 1
          memoryInGb: 1.5
      ports:
      - port: 8080
  osType: Linux
  restartPolicy: Always
```

## Monitoring and Logging

### 1. Performance Monitoring

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'liv-processor'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
```

### 2. Log Aggregation

```yaml
# fluentd.conf
<source>
  @type tail
  path /var/log/liv/*.log
  pos_file /var/log/fluentd/liv.log.pos
  tag liv.*
  format json
</source>

<match liv.**>
  @type elasticsearch
  host elasticsearch
  port 9200
  index_name liv-logs
</match>
```

### 3. Health Checks

```go
// health/check.go
func HealthCheck() error {
    // Check CLI tools
    if err := checkCLITools(); err != nil {
        return err
    }
    
    // Check WASM modules
    if err := checkWASMModules(); err != nil {
        return err
    }
    
    // Check security system
    if err := checkSecuritySystem(); err != nil {
        return err
    }
    
    return nil
}
```

## Security Considerations

### 1. Production Security

```yaml
# security-config.yml
security:
  level: "strict"
  signing:
    required: true
    algorithm: "RSA-SHA256"
  sandbox:
    memory_limit: "64MB"
    network_access: false
    file_system_access: false
  audit:
    enabled: true
    log_level: "info"
```

### 2. Certificate Management

```bash
# Generate production certificates
openssl genrsa -out private.pem 4096
openssl rsa -in private.pem -pubout -out public.pem

# Store securely
aws secretsmanager create-secret --name "liv-signing-key" --secret-string file://private.pem
```

### 3. Network Security

```yaml
# network-policy.yml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: liv-processor-policy
spec:
  podSelector:
    matchLabels:
      app: liv-processor
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: liv-frontend
    ports:
    - protocol: TCP
      port: 8080
```

## Maintenance and Updates

### 1. Update Process

```bash
# Automated update script
#!/bin/bash
set -e

# Download latest release
wget "https://github.com/your-org/liv-document-format/releases/latest/download/liv-document-format-linux-amd64.zip"

# Backup current installation
cp -r /opt/liv /opt/liv.backup.$(date +%Y%m%d)

# Install update
unzip -o liv-document-format-linux-amd64.zip -d /opt/liv

# Restart services
systemctl restart liv-processor
systemctl restart liv-viewer

# Verify installation
/opt/liv/bin/liv-cli --version
```

### 2. Rollback Procedure

```bash
# Rollback script
#!/bin/bash
BACKUP_DIR="/opt/liv.backup.$(date +%Y%m%d)"

if [ -d "$BACKUP_DIR" ]; then
    systemctl stop liv-processor
    systemctl stop liv-viewer
    
    rm -rf /opt/liv
    mv "$BACKUP_DIR" /opt/liv
    
    systemctl start liv-processor
    systemctl start liv-viewer
    
    echo "Rollback completed successfully"
else
    echo "Backup directory not found: $BACKUP_DIR"
    exit 1
fi
```

### 3. Database Migrations

```sql
-- migrations/001_initial_schema.sql
CREATE TABLE documents (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    content_hash VARCHAR(64) NOT NULL,
    signature_hash VARCHAR(64),
    security_policy JSONB NOT NULL
);

CREATE INDEX idx_documents_author ON documents(author);
CREATE INDEX idx_documents_created ON documents(created_at);
```

## Testing in Production

### 1. Smoke Tests

```bash
# smoke-test.sh
#!/bin/bash
set -e

echo "Running smoke tests..."

# Test CLI tools
liv-cli --version
liv-viewer --version
liv-builder --version

# Test document creation
echo '<html><body><h1>Test</h1></body></html>' > test.html
liv-cli build --source . --output test.liv
liv-cli validate test.liv

# Test viewer
liv-cli view test.liv --headless

# Cleanup
rm test.html test.liv

echo "Smoke tests passed!"
```

### 2. Load Testing

```javascript
// load-test.js
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 100 },
    { duration: '5m', target: 100 },
    { duration: '2m', target: 200 },
    { duration: '5m', target: 200 },
    { duration: '2m', target: 0 },
  ],
};

export default function () {
  let response = http.post('http://localhost:8080/process', {
    document: 'base64-encoded-document-data'
  });
  
  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
}
```

## Support and Documentation

### 1. Documentation Deployment

```bash
# Deploy documentation
cd docs
npm install -g @gitbook/cli
gitbook build
aws s3 sync _book/ s3://docs.liv-format.org --delete
```

### 2. Support Infrastructure

```yaml
# support-stack.yml
version: '3.8'

services:
  documentation:
    image: nginx:alpine
    volumes:
      - ./docs/_book:/usr/share/nginx/html
    ports:
      - "80:80"

  support-api:
    build: ./support-api
    ports:
      - "3000:3000"
    environment:
      - DATABASE_URL=postgresql://user:pass@db:5432/support

  db:
    image: postgres:13
    environment:
      - POSTGRES_DB=support
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=pass
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

This deployment guide provides comprehensive instructions for building, packaging, and deploying the LIV Document Format system across different environments and platforms. The system is designed to be scalable, secure, and maintainable in production environments.