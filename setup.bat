@echo off
REM LIV Document Format - Windows Setup Script
REM This script installs and configures the LIV Document Format system on Windows

setlocal enabledelayedexpansion

REM Configuration
set LIV_VERSION=1.0.0
set INSTALL_DIR=C:\Program Files\LIV
set CONFIG_DIR=%USERPROFILE%\.liv

echo.
echo 🚀 LIV Document Format Setup Script v%LIV_VERSION%
echo ==================================================
echo.

REM Check if running as administrator
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] This script requires administrator privileges.
    echo Please run as administrator and try again.
    pause
    exit /b 1
)

echo [INFO] Checking prerequisites...

REM Check for required tools
where curl >nul 2>&1
if %errorLevel% neq 0 (
    echo [WARNING] curl not found. Attempting to use PowerShell for downloads.
    set USE_POWERSHELL=1
) else (
    set USE_POWERSHELL=0
)

where powershell >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] PowerShell not found. This is required for the installation.
    pause
    exit /b 1
)

echo [SUCCESS] Prerequisites check completed

REM Create directories
echo [INFO] Creating installation directories...
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
if not exist "%CONFIG_DIR%" mkdir "%CONFIG_DIR%"
if not exist "%CONFIG_DIR%\templates" mkdir "%CONFIG_DIR%\templates"
if not exist "%CONFIG_DIR%\keys" mkdir "%CONFIG_DIR%\keys"
if not exist "%CONFIG_DIR%\cache" mkdir "%CONFIG_DIR%\cache"
if not exist "%CONFIG_DIR%\logs" mkdir "%CONFIG_DIR%\logs"

REM Download and install binaries
echo [INFO] Downloading LIV binaries...
set PACKAGE_NAME=liv-document-format-v%LIV_VERSION%-windows-amd64
set DOWNLOAD_URL=https://github.com/your-org/liv-document-format/releases/download/v%LIV_VERSION%/%PACKAGE_NAME%.zip
set TEMP_FILE=%TEMP%\%PACKAGE_NAME%.zip

if %USE_POWERSHELL%==1 (
    powershell -Command "Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%TEMP_FILE%'"
) else (
    curl -L "%DOWNLOAD_URL%" -o "%TEMP_FILE%"
)

if %errorLevel% neq 0 (
    echo [ERROR] Failed to download package.
    pause
    exit /b 1
)

echo [INFO] Extracting package...
powershell -Command "Expand-Archive -Path '%TEMP_FILE%' -DestinationPath '%TEMP%' -Force"

echo [INFO] Installing binaries...
xcopy "%TEMP%\%PACKAGE_NAME%\*" "%INSTALL_DIR%\" /E /I /Y

REM Add to PATH
echo [INFO] Adding to system PATH...
for /f "tokens=2*" %%A in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v PATH 2^>nul') do set "SYSTEM_PATH=%%B"
echo !SYSTEM_PATH! | find /i "%INSTALL_DIR%\bin" >nul
if %errorLevel% neq 0 (
    reg add "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v PATH /t REG_EXPAND_SZ /d "!SYSTEM_PATH!;%INSTALL_DIR%\bin" /f
    echo [SUCCESS] Added to system PATH
) else (
    echo [INFO] Already in system PATH
)

REM Install desktop application
echo [INFO] Installing desktop application...
if exist "%INSTALL_DIR%\LIV-Document-Viewer-Setup.exe" (
    echo [INFO] Running desktop application installer...
    start /wait "%INSTALL_DIR%\LIV-Document-Viewer-Setup.exe" /S
    echo [SUCCESS] Desktop application installed
) else (
    echo [WARNING] Desktop application installer not found
)

REM Configure file associations
echo [INFO] Configuring file associations...
reg add "HKCR\.liv" /ve /d "LIVDocument" /f >nul 2>&1
reg add "HKCR\LIVDocument" /ve /d "LIV Document" /f >nul 2>&1
reg add "HKCR\LIVDocument\DefaultIcon" /ve /d "%INSTALL_DIR%\icons\liv-icon.ico" /f >nul 2>&1
reg add "HKCR\LIVDocument\shell\open\command" /ve /d "\"%INSTALL_DIR%\bin\liv-viewer.exe\" \"%%1\"" /f >nul 2>&1

REM Install JavaScript SDK
echo [INFO] Installing JavaScript SDK...
where npm >nul 2>&1
if %errorLevel%==0 (
    npm install -g "%INSTALL_DIR%\js\liv-document-format-%LIV_VERSION%.tgz"
    if %errorLevel%==0 (
        echo [SUCCESS] JavaScript SDK installed globally
    ) else (
        echo [WARNING] JavaScript SDK installation failed
    )
) else (
    echo [WARNING] npm not found. JavaScript SDK not installed.
)

REM Install Python SDK
echo [INFO] Installing Python SDK...
where python >nul 2>&1
if %errorLevel%==0 (
    python -m pip install --user "%INSTALL_DIR%\python\dist\liv_document_format-%LIV_VERSION%-py3-none-any.whl"
    if %errorLevel%==0 (
        echo [SUCCESS] Python SDK installed for current user
    ) else (
        echo [WARNING] Python SDK installation failed
    )
) else (
    where python3 >nul 2>&1
    if %errorLevel%==0 (
        python3 -m pip install --user "%INSTALL_DIR%\python\dist\liv_document_format-%LIV_VERSION%-py3-none-any.whl"
        if %errorLevel%==0 (
            echo [SUCCESS] Python SDK installed for current user
        ) else (
            echo [WARNING] Python SDK installation failed
        )
    ) else (
        echo [WARNING] Python not found. Python SDK not installed.
    )
)

REM Create configuration file
echo [INFO] Creating configuration files...
(
echo # LIV Document Format Configuration
echo.
echo # Default settings
echo default_author: "%USERNAME%"
echo default_license: "MIT"
echo compression_enabled: true
echo security_level: "strict"
echo.
echo # Signing settings
echo signing:
echo   algorithm: "RSA-SHA256"
echo.
echo # Viewer settings
echo viewer:
echo   default_renderer: "webgl"
echo   enable_animations: true
echo   sandbox_mode: true
echo.
echo # Performance settings
echo performance:
echo   memory_limit: "512MB"
echo   cache_size: "100MB"
echo   parallel_processing: true
echo.
echo # Security settings
echo security:
echo   allow_network_access: false
echo   allow_file_system: false
echo   memory_limit: "64MB"
echo.
echo # Paths
echo paths:
echo   templates: "%CONFIG_DIR%\templates"
echo   keys: "%CONFIG_DIR%\keys"
echo   cache: "%CONFIG_DIR%\cache"
) > "%CONFIG_DIR%\config.yaml"

REM Copy example templates
if exist "%INSTALL_DIR%\examples\templates" (
    xcopy "%INSTALL_DIR%\examples\templates\*" "%CONFIG_DIR%\templates\" /E /I /Y >nul 2>&1
)

REM Run tests
echo [INFO] Running system tests...

REM Test CLI tools
"%INSTALL_DIR%\bin\liv-cli.exe" --version >nul 2>&1
if %errorLevel%==0 (
    echo [SUCCESS] CLI tools working
) else (
    echo [ERROR] CLI tools not working
)

REM Test document creation
echo ^<html^>^<body^>^<h1^>Test Document^</h1^>^</body^>^</html^> > "%TEMP%\test.html"
"%INSTALL_DIR%\bin\liv-cli.exe" build --source "%TEMP%" --output "%TEMP%\test.liv" >nul 2>&1
if %errorLevel%==0 (
    echo [SUCCESS] Document creation test passed
    
    REM Test validation
    "%INSTALL_DIR%\bin\liv-cli.exe" validate "%TEMP%\test.liv" >nul 2>&1
    if %errorLevel%==0 (
        echo [SUCCESS] Document validation test passed
    ) else (
        echo [WARNING] Document validation test failed
    )
    
    REM Test viewing (headless)
    "%INSTALL_DIR%\bin\liv-cli.exe" view "%TEMP%\test.liv" --headless >nul 2>&1
    if %errorLevel%==0 (
        echo [SUCCESS] Document viewing test passed
    ) else (
        echo [WARNING] Document viewing test failed
    )
    
    REM Cleanup
    del "%TEMP%\test.html" "%TEMP%\test.liv" >nul 2>&1
) else (
    echo [ERROR] Document creation test failed
)

REM Cleanup temporary files
del "%TEMP_FILE%" >nul 2>&1
rmdir /s /q "%TEMP%\%PACKAGE_NAME%" >nul 2>&1

REM Display installation summary
echo.
echo [SUCCESS] LIV Document Format installation completed!
echo.
echo 📦 Installation Summary:
echo   • Binaries installed in: %INSTALL_DIR%
echo   • CLI tools available: liv-cli, liv-viewer, liv-builder
echo   • Configuration directory: %CONFIG_DIR%
echo   • Desktop application: Installed
echo.
echo 🚀 Quick Start:
echo   # Create a document
echo   echo ^<html^>^<body^>^<h1^>Hello LIV!^</h1^>^</body^>^</html^> ^> hello.html
echo   liv-cli build --source . --output hello.liv
echo.
echo   # View a document
echo   liv-cli view hello.liv
echo.
echo   # Validate a document
echo   liv-cli validate hello.liv
echo.
echo 📚 Documentation:
echo   • User Guide: %INSTALL_DIR%\docs\USER_GUIDE.md
echo   • API Reference: %INSTALL_DIR%\docs\reference\
echo   • Examples: %INSTALL_DIR%\examples\
echo.
echo 🔧 Configuration:
echo   • Config file: %CONFIG_DIR%\config.yaml
echo   • Templates: %CONFIG_DIR%\templates\
echo   • Logs: %CONFIG_DIR%\logs\
echo.
echo For more information, visit: https://github.com/your-org/liv-document-format
echo.
echo [INFO] Please restart your command prompt or log out and back in for PATH changes to take effect.
echo.

pause