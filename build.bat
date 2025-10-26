@echo off
REM Build script for LIV Format project (Windows)

echo Building LIV Format Binaries...
echo ================================

REM Build main CLI
echo Building liv CLI...
go build -o bin\liv.exe .\cmd\cli\

REM Build PDF operations tool
echo Building liv-pdf...
go build -o bin\liv-pdf.exe .\cmd\liv-pdf\

REM Build builder
echo Building liv-builder...
go build -o bin\liv-builder.exe .\cmd\builder\

REM Build viewer
echo Building liv-viewer...
go build -o bin\liv-viewer.exe .\cmd\viewer\

REM Build integrity checker
echo Building liv-integrity...
go build -o bin\liv-integrity.exe .\cmd\liv-integrity\

REM Build manifest validator
echo Building liv-manifest-validator...
go build -o bin\liv-manifest-validator.exe .\cmd\manifest-validator\

REM Build security admin
echo Building liv-security-admin...
go build -o bin\liv-security-admin.exe .\cmd\security-admin\

REM Build permission server
echo Building liv-permission-server...
go build -o bin\liv-permission-server.exe .\cmd\permission-server\

REM Build liv-pack
echo Building liv-pack...
go build -o bin\liv-pack.exe .\cmd\liv-pack\

REM Build liv-converter
echo Building liv-converter...
go build -o bin\liv-converter.exe .\cmd\liv-converter\

echo.
echo All binaries built successfully!
echo   Output: .\bin\
echo.
echo Available commands:
echo   - liv: Main CLI for LIV documents
echo   - liv-pdf: PDF operations (extract, merge, split, etc.)
echo   - liv-converter: PDF to LIV converter (RECOMMENDED FOR PDF→LIV)
echo   - liv-builder: Document builder
echo   - liv-viewer: Document viewer
echo   - liv-integrity: Integrity checker
echo   - liv-manifest-validator: Manifest validator
echo   - liv-security-admin: Security administration
echo   - liv-permission-server: Permission server
echo   - liv-pack: Document packer
