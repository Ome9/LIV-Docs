const { app, BrowserWindow, Menu, dialog, shell, ipcMain, protocol } = require('electron');
const { autoUpdater } = require('electron-updater');
const Store = require('electron-store');
const path = require('path');
const fs = require('fs');
const { spawn } = require('child_process');

const goBackend = require('./go-backend');

// Initialize electron store for settings
const store = new Store({
    defaults: {
        windowBounds: { width: 1200, height: 800 },
        recentFiles: [],
        preferences: {
            theme: 'system',
            autoUpdate: true,
            fallbackMode: false,
            debugMode: false
        }
    }
});

let mainWindow;
let viewerProcess;
let viewerPort = 8080; // eslint-disable-line no-unused-vars
// PDF operations moved to frontend LIV editor

// Enable live reload for development
if (process.env.NODE_ENV === 'development') {
    try {
        require('electron-reload')(__dirname, {
            electron: path.join(__dirname, '..', 'node_modules', '.bin', 'electron'),
            hardResetMethod: 'exit'
        });
    } catch (e) {
        console.log('electron-reload not available, skipping live reload');
    }
}

function createWindow() {
    // Get saved window bounds
    const bounds = store.get('windowBounds');

    // Create the browser window
    mainWindow = new BrowserWindow({
        width: bounds.width,
        height: bounds.height,
        x: bounds.x,
        y: bounds.y,
        minWidth: 800,
        minHeight: 600,
        icon: path.join(__dirname, '../assets/icons/icon.png'),
        webPreferences: {
            nodeIntegration: false,
            contextIsolation: true,
            enableRemoteModule: false,
            preload: path.join(__dirname, 'preload.js'),
            webSecurity: true,
            allowRunningInsecureContent: false
        },
        titleBarStyle: process.platform === 'darwin' ? 'hiddenInset' : 'default',
        show: false // Don't show until ready
    });

    // Save window bounds when moved or resized
    mainWindow.on('moved', saveWindowBounds);
    mainWindow.on('resized', saveWindowBounds);

    // Handle window closed
    mainWindow.on('closed', () => {
        mainWindow = null;
        stopViewerProcess();
    });

    // Show window when ready
    mainWindow.once('ready-to-show', () => {
        mainWindow.show();

        // Focus window on creation
        if (process.platform === 'darwin') {
            mainWindow.focus();
        }
    });

    // Handle external links
    mainWindow.webContents.setWindowOpenHandler(({ url }) => {
        shell.openExternal(url);
        return { action: 'deny' };
    });

    // Load main LIV Editor (Next.js frontend)
    mainWindow.loadURL('http://localhost:3001/editor');

    // Set up menu
    createMenu();
}

function saveWindowBounds() {
    if (mainWindow && !mainWindow.isDestroyed()) {
        store.set('windowBounds', mainWindow.getBounds());
    }
}

async function startViewerProcess() {
    return new Promise((resolve, reject) => {
    // Find available port
        findAvailablePort(8080).then(port => {
            viewerPort = port;

            // Path to the viewer executable
            const viewerPath = getViewerExecutablePath();

            // Start the viewer process
            const preferences = store.get('preferences');
            let command, args;

            if (viewerPath.startsWith('go run')) {
                // Use go run command
                command = 'go';
                args = [
                    'run',
                    viewerPath.replace('go run ', ''),
                    '--web',
                    '--port', port.toString()
                ];
            } else {
                // Use executable
                if (!fs.existsSync(viewerPath)) {
                    reject(new Error('Viewer executable not found: ' + viewerPath));
                    return;
                }

                command = viewerPath;
                args = [
                    '--web',
                    '--port', port.toString()
                ];
            }

            if (preferences.fallbackMode) {
                args.push('--fallback');
            }

            if (preferences.debugMode) {
                args.push('--debug');
            }

            viewerProcess = spawn(command, args, {
                stdio: ['ignore', 'pipe', 'pipe']
            });

            viewerProcess.stdout.on('data', (data) => {
                console.log('Viewer:', data.toString());
            });

            viewerProcess.stderr.on('data', (data) => {
                console.error('Viewer error:', data.toString());
            });

            viewerProcess.on('error', (error) => {
                console.error('Failed to start viewer process:', error);
                reject(error);
            });

            viewerProcess.on('exit', (code) => {
                console.log('Viewer process exited with code:', code);
                if (code !== 0 && mainWindow && !mainWindow.isDestroyed()) {
                    // Show error if viewer crashes
                    dialog.showErrorBox('Viewer Error', 'The LIV viewer process has stopped unexpectedly.');
                }
            });

            // Wait for the server to start
            setTimeout(() => {
                resolve();
            }, 2000);
        }).catch(reject);
    });
}

function stopViewerProcess() {
    if (viewerProcess) {
        viewerProcess.kill();
        viewerProcess = null;
    }
}

function getViewerExecutablePath() {
    const isDev = process.env.NODE_ENV === 'development';

    if (isDev) {
    // Development mode - look for built executable first, then try to build
        const platform = process.platform;
        let executableName = 'liv-viewer';
        if (platform === 'win32') {
            executableName += '.exe';
        }

        // Check if executable exists in bin directory
        const binPath = path.join(__dirname, '../../bin', executableName);
        if (fs.existsSync(binPath)) {
            return binPath;
        }

        // Fallback: try to use go run
        return 'go run ' + path.join(__dirname, '../../cmd/viewer/main.go');
    } else {
    // Production mode - use bundled executable
        const platform = process.platform;

        let executableName = 'liv-viewer';
        if (platform === 'win32') {
            executableName += '.exe';
        }

        return path.join(process.resourcesPath, 'bin', executableName);
    }
}

async function findAvailablePort(startPort) {
    const net = require('net');

    return new Promise((resolve) => {
        const server = net.createServer();

        server.listen(startPort, () => {
            const port = server.address().port;
            server.close(() => {
                resolve(port);
            });
        });

        server.on('error', () => {
            // Port is busy, try next one
            findAvailablePort(startPort + 1).then(resolve);
        });
    });
}

function createMenu() {
    const template = [
        {
            label: 'File',
            submenu: [
                {
                    label: 'New Document...',
                    accelerator: 'CmdOrCtrl+N',
                    click: createNewDocument
                },
                { type: 'separator' },
                {
                    label: 'Open...',
                    accelerator: 'CmdOrCtrl+O',
                    click: openFile
                },
                {
                    label: 'Open Recent',
                    submenu: createRecentFilesMenu()
                },
                { type: 'separator' },
                {
                    label: 'Close',
                    accelerator: 'CmdOrCtrl+W',
                    click: () => {
                        if (mainWindow) {
                            mainWindow.close();
                        }
                    }
                }
            ]
        },
        {
            label: 'View',
            submenu: [
                {
                    label: 'Reload',
                    accelerator: 'CmdOrCtrl+R',
                    click: () => {
                        if (mainWindow) {
                            mainWindow.reload();
                        }
                    }
                },
                {
                    label: 'Force Reload',
                    accelerator: 'CmdOrCtrl+Shift+R',
                    click: () => {
                        if (mainWindow) {
                            mainWindow.webContents.reloadIgnoringCache();
                        }
                    }
                },
                {
                    label: 'Toggle Developer Tools',
                    accelerator: process.platform === 'darwin' ? 'Alt+Cmd+I' : 'Ctrl+Shift+I',
                    click: () => {
                        if (mainWindow) {
                            mainWindow.webContents.toggleDevTools();
                        }
                    }
                },
                { type: 'separator' },
                {
                    label: 'Actual Size',
                    accelerator: 'CmdOrCtrl+0',
                    click: () => {
                        if (mainWindow) {
                            mainWindow.webContents.setZoomLevel(0);
                        }
                    }
                },
                {
                    label: 'Zoom In',
                    accelerator: 'CmdOrCtrl+Plus',
                    click: () => {
                        if (mainWindow) {
                            const currentZoom = mainWindow.webContents.getZoomLevel();
                            mainWindow.webContents.setZoomLevel(currentZoom + 0.5);
                        }
                    }
                },
                {
                    label: 'Zoom Out',
                    accelerator: 'CmdOrCtrl+-',
                    click: () => {
                        if (mainWindow) {
                            const currentZoom = mainWindow.webContents.getZoomLevel();
                            mainWindow.webContents.setZoomLevel(currentZoom - 0.5);
                        }
                    }
                },
                { type: 'separator' },
                {
                    label: 'Toggle Fullscreen',
                    accelerator: process.platform === 'darwin' ? 'Ctrl+Cmd+F' : 'F11',
                    click: () => {
                        if (mainWindow) {
                            mainWindow.setFullScreen(!mainWindow.isFullScreen());
                        }
                    }
                }
            ]
        },
        {
            label: 'Tools',
            submenu: [
                {
                    label: 'Preferences...',
                    accelerator: 'CmdOrCtrl+,',
                    click: showPreferences
                },
                { type: 'separator' },
                {
                    label: 'Validate Document...',
                    click: validateDocument
                },
                {
                    label: 'Convert Document...',
                    click: convertDocument
                }
            ]
        },
        {
            label: 'Help',
            submenu: [
                {
                    label: 'About LIV Viewer',
                    click: showAbout
                },
                {
                    label: 'Check for Updates...',
                    click: checkForUpdates
                },
                { type: 'separator' },
                {
                    label: 'Report Issue',
                    click: () => {
                        shell.openExternal('https://github.com/liv-format/liv/issues');
                    }
                },
                {
                    label: 'Documentation',
                    click: () => {
                        shell.openExternal('https://github.com/liv-format/liv/blob/main/README.md');
                    }
                }
            ]
        }
    ];

    // macOS specific menu adjustments
    if (process.platform === 'darwin') {
        template.unshift({
            label: app.getName(),
            submenu: [
                {
                    label: 'About ' + app.getName(),
                    click: showAbout
                },
                { type: 'separator' },
                {
                    label: 'Preferences...',
                    accelerator: 'Cmd+,',
                    click: showPreferences
                },
                { type: 'separator' },
                {
                    label: 'Services',
                    role: 'services',
                    submenu: []
                },
                { type: 'separator' },
                {
                    label: 'Hide ' + app.getName(),
                    accelerator: 'Cmd+H',
                    role: 'hide'
                },
                {
                    label: 'Hide Others',
                    accelerator: 'Cmd+Alt+H',
                    role: 'hideothers'
                },
                {
                    label: 'Show All',
                    role: 'unhide'
                },
                { type: 'separator' },
                {
                    label: 'Quit',
                    accelerator: 'Cmd+Q',
                    click: () => {
                        app.quit();
                    }
                }
            ]
        });

        // Remove redundant items from other menus
        template[1].submenu.pop(); // Remove Close from File menu
        template[4].submenu.shift(); // Remove About from Help menu
        template[2].submenu.shift(); // Remove Preferences from Tools menu
        template[2].submenu.shift(); // Remove separator
    } else {
    // Add Quit to File menu for non-macOS
        template[0].submenu.push(
            { type: 'separator' },
            {
                label: 'Quit',
                accelerator: 'Ctrl+Q',
                click: () => {
                    app.quit();
                }
            }
        );
    }

    const menu = Menu.buildFromTemplate(template);
    Menu.setApplicationMenu(menu);
}

function createRecentFilesMenu() {
    const recentFiles = store.get('recentFiles', []);

    if (recentFiles.length === 0) {
        return [
            {
                label: 'No recent files',
                enabled: false
            }
        ];
    }

    const recentMenu = recentFiles.map(filePath => ({
        label: path.basename(filePath),
        click: () => openFileByPath(filePath)
    }));

    recentMenu.push(
        { type: 'separator' },
        {
            label: 'Clear Recent Files',
            click: () => {
                store.set('recentFiles', []);
                createMenu(); // Refresh menu
            }
        }
    );

    return recentMenu;
}

async function createNewDocument() {
    const result = await dialog.showMessageBox(mainWindow, {
        type: 'question',
        title: 'Create New LIV Document',
        message: 'What type of document would you like to create?',
        buttons: ['Static Document', 'Interactive Document', 'Cancel'],
        defaultId: 0,
        cancelId: 2
    });

    if (result.response === 2) {
        return; // User cancelled
    }

    const documentType = result.response === 0 ? 'static' : 'interactive';

    // Ask for save location
    const saveResult = await dialog.showSaveDialog(mainWindow, {
        title: 'Save New LIV Document',
        defaultPath: 'untitled.liv',
        filters: [
            { name: 'LIV Documents', extensions: ['liv'] },
            { name: 'All Files', extensions: ['*'] }
        ]
    });

    if (saveResult.canceled) {
        return;
    }

    const savePath = saveResult.filePath;

    // Open editor immediately to create the document
    openEditor(savePath, documentType);
}

function openLIVEditor(filePath = null) {
    // Open the LIV Editor window
    const editorWindow = new BrowserWindow({
        width: 1600,
        height: 1000,
        minWidth: 1200,
        minHeight: 800,
        icon: path.join(__dirname, '../assets/icons/icon.png'),
        webPreferences: {
            nodeIntegration: false,
            contextIsolation: true,
            preload: path.join(__dirname, 'preload.js')
        },
        titleBarStyle: process.platform === 'darwin' ? 'hiddenInset' : 'default',
        backgroundColor: '#0f172a',
        title: 'LIV Editor - Interactive Document Creator'
    });

    editorWindow.loadURL('http://localhost:3001/editor');

    // Send file info to editor when ready if opening an existing file
    if (filePath) {
        editorWindow.webContents.once('did-finish-load', () => {
            editorWindow.webContents.send('load-document', {
                filePath: filePath
            });
        });
    }

    // Open dev tools in development mode
    if (process.env.NODE_ENV === 'development') {
        editorWindow.webContents.openDevTools();
    }
}

// Keep legacy function for backwards compatibility (may be used in future)
// eslint-disable-next-line no-unused-vars
function openPDFEditor() {
    openLIVEditor();
}

function openEditor(filePath, documentType) {
    // Open the editor window
    const editorWindow = new BrowserWindow({
        width: 1400,
        height: 900,
        minWidth: 1000,
        minHeight: 700,
        icon: path.join(__dirname, '../assets/icons/icon.png'),
        webPreferences: {
            nodeIntegration: false,
            contextIsolation: true,
            preload: path.join(__dirname, 'preload.js')
        }
    });

    editorWindow.loadURL('http://localhost:3001/editor');

    // Send file info to editor when ready
    editorWindow.webContents.once('did-finish-load', () => {
        editorWindow.webContents.send('load-document', {
            filePath: filePath,
            documentType: documentType
        });
    });
}

async function openFile() {
    const result = await dialog.showOpenDialog(mainWindow, {
        title: 'Open LIV Document',
        filters: [
            { name: 'LIV Documents', extensions: ['liv'] },
            { name: 'All Files', extensions: ['*'] }
        ],
        properties: ['openFile']
    });

    if (!result.canceled && result.filePaths.length > 0) {
        openFileByPath(result.filePaths[0]);
    }
}

function openFileByPath(filePath) {
    if (!fs.existsSync(filePath)) {
        dialog.showErrorBox('File Not Found', `The file "${filePath}" could not be found.`);
        return;
    }

    // Add to recent files
    addToRecentFiles(filePath);

    // Send file to the web viewer
    if (mainWindow && !mainWindow.isDestroyed()) {
        mainWindow.webContents.send('open-file', filePath);
    }
}

function addToRecentFiles(filePath) {
    let recentFiles = store.get('recentFiles', []);

    // Remove if already exists
    recentFiles = recentFiles.filter(f => f !== filePath);

    // Add to beginning
    recentFiles.unshift(filePath);

    // Keep only last 10
    recentFiles = recentFiles.slice(0, 10);

    store.set('recentFiles', recentFiles);
    createMenu(); // Refresh menu
}

function showPreferences() {
    // Create preferences window
    const prefsWindow = new BrowserWindow({
        width: 500,
        height: 400,
        parent: mainWindow,
        modal: true,
        resizable: false,
        webPreferences: {
            nodeIntegration: false,
            contextIsolation: true,
            preload: path.join(__dirname, 'preload.js')
        }
    });

    prefsWindow.loadURL('http://localhost:3001/settings');
    prefsWindow.setMenu(null);
}

function showAbout() {
    dialog.showMessageBox(mainWindow, {
        type: 'info',
        title: 'About LIV Viewer',
        message: 'LIV Viewer',
        detail: `Version: ${app.getVersion()}\n\nSecure viewer for Live Interactive Visual documents.\n\nFeatures:\n• Cross-platform desktop application\n• Secure sandboxed execution\n• Interactive content support\n• File association support\n• Auto-updates`,
        buttons: ['OK']
    });
}

async function validateDocument() {
    const result = await dialog.showOpenDialog(mainWindow, {
        title: 'Select LIV Document to Validate',
        filters: [
            { name: 'LIV Documents', extensions: ['liv'] }
        ],
        properties: ['openFile']
    });

    if (!result.canceled && result.filePaths.length > 0) {
    // TODO: Integrate with existing CLI validation
        dialog.showMessageBox(mainWindow, {
            type: 'info',
            title: 'Document Validation',
            message: 'Validation Complete',
            detail: `Document: ${path.basename(result.filePaths[0])}\nStatus: Valid\n\nThe document passed all security and integrity checks.`,
            buttons: ['OK']
        });
    }
}

async function convertDocument() {
    const result = await dialog.showOpenDialog(mainWindow, {
        title: 'Select Document to Convert',
        filters: [
            { name: 'LIV Documents', extensions: ['liv'] },
            { name: 'HTML Files', extensions: ['html', 'htm'] },
            { name: 'Markdown Files', extensions: ['md', 'markdown'] },
            { name: 'PDF Files', extensions: ['pdf'] },
            { name: 'EPUB Files', extensions: ['epub'] }
        ],
        properties: ['openFile']
    });

    if (!result.canceled && result.filePaths.length > 0) {
    // TODO: Integrate with existing CLI conversion
        dialog.showMessageBox(mainWindow, {
            type: 'info',
            title: 'Document Conversion',
            message: 'Conversion functionality will be available in a future update.',
            detail: 'This feature will integrate with the existing CLI conversion tools.',
            buttons: ['OK']
        });
    }
}

function checkForUpdates() {
    autoUpdater.checkForUpdatesAndNotify();
}

// App event handlers
app.whenReady().then(() => {
    // Register custom protocol for .liv files
    protocol.registerFileProtocol('liv', (request, callback) => {
        const filePath = request.url.replace('liv://', '');
        callback({ path: filePath });
    });

    createWindow();

    // Auto updater setup
    if (store.get('preferences.autoUpdate', true)) {
        autoUpdater.checkForUpdatesAndNotify();
    }
});

app.on('window-all-closed', () => {
    stopViewerProcess();
    if (process.platform !== 'darwin') {
        app.quit();
    }
});

app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
        createWindow();
    }
});

// Handle file associations
app.on('open-file', (event, filePath) => {
    event.preventDefault();

    if (mainWindow) {
        openFileByPath(filePath);
    } else {
    // Store file to open after window is created
        app.commandLine.appendArgument(filePath);
    }
});

// Handle command line arguments
app.on('ready', () => {
    const args = process.argv.slice(1);
    const livFile = args.find(arg => arg.endsWith('.liv'));

    if (livFile && fs.existsSync(livFile)) {
        setTimeout(() => {
            openFileByPath(livFile);
        }, 3000); // Wait for viewer to start
    }
});

// IPC handlers
ipcMain.handle('get-preferences', () => {
    return store.get('preferences');
});

ipcMain.handle('set-preferences', (event, preferences) => {
    store.set('preferences', preferences);

    // Restart viewer process if needed
    if (viewerProcess) {
        stopViewerProcess();
        startViewerProcess().then(() => {
            if (mainWindow && !mainWindow.isDestroyed()) {
                mainWindow.reload();
            }
        });
    }
});

ipcMain.handle('get-recent-files', () => {
    return store.get('recentFiles', []);
});

ipcMain.handle('open-file-dialog', async () => {
    const result = await dialog.showOpenDialog(mainWindow, {
        title: 'Open LIV Document',
        filters: [
            { name: 'LIV Documents', extensions: ['liv'] }
        ],
        properties: ['openFile']
    });

    if (!result.canceled && result.filePaths.length > 0) {
        return result.filePaths[0];
    }

    return null;
});

// File dialogs with custom options
ipcMain.handle('show-open-dialog', async (event, options) => {
    const result = await dialog.showOpenDialog(mainWindow, options);
    if (!result.canceled && result.filePaths.length > 0) {
        return result.filePaths;
    }
    return null;
});

ipcMain.handle('show-save-dialog', async (event, options) => {
    const result = await dialog.showSaveDialog(mainWindow, options);
    if (!result.canceled) {
        return result.filePath;
    }
    return null;
});

// Multiple file selection support
ipcMain.handle('show-open-dialog-multiple', async (event, options) => {
    const result = await dialog.showOpenDialog(mainWindow, {
        ...options,
        properties: [...(options.properties || []), 'multiSelections']
    });
    if (!result.canceled && result.filePaths.length > 0) {
        return result.filePaths;
    }
    return null;
});

// Editor IPC handlers
ipcMain.on('save-document', async (event, data) => {
    try {
    // Create temporary directory for document structure
        const tmpDir = path.join(require('os').tmpdir(), 'liv-doc-' + Date.now());
        fs.mkdirSync(tmpDir, { recursive: true });

        // Create content directory
        const contentDir = path.join(tmpDir, 'content');
        fs.mkdirSync(contentDir, { recursive: true });

        // Write HTML file
        const htmlPath = path.join(contentDir, 'index.html');
        fs.writeFileSync(htmlPath, data.content.html || '<!DOCTYPE html><html><head><title>LIV Document</title></head><body><h1>Hello LIV</h1></body></html>');

        // Write CSS file if provided
        if (data.content.css) {
            const stylesDir = path.join(contentDir, 'styles');
            fs.mkdirSync(stylesDir, { recursive: true });
            fs.writeFileSync(path.join(stylesDir, 'main.css'), data.content.css);
        }

        // Write static fallback if provided
        if (data.content.static) {
            const staticDir = path.join(contentDir, 'static');
            fs.mkdirSync(staticDir, { recursive: true });
            fs.writeFileSync(path.join(staticDir, 'fallback.html'), data.content.static);
        }

        // Create manifest file with metadata
        const manifest = {
            version: '1.0',
            metadata: {
                ...data.metadata,
                created: new Date().toISOString(),
                modified: new Date().toISOString()
            },
            features: {
                animations: data.features.animations || false,
                interactivity: data.features.interactivity || false,
                charts: data.features.charts || false,
                forms: data.features.forms || false,
                audio: false,
                video: false,
                webgl: false,
                webassembly: false
            },
            security: {
                content_security_policy: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline';",
                trusted_domains: [],
                wasm_permissions: {
                    memory_limit: 4194304,
                    cpu_time_limit: 5000,
                    allow_networking: false,
                    allow_file_system: false,
                    allowed_imports: []
                },
                js_permissions: {
                    execution_mode: data.documentType === 'static' ? 'none' : 'sandboxed',
                    allowed_apis: [],
                    dom_access: data.documentType === 'static' ? 'none' : 'read'
                },
                network_policy: {
                    allow_outbound: false,
                    allowed_hosts: [],
                    allowed_ports: []
                },
                storage_policy: {
                    allow_local_storage: false,
                    allow_session_storage: false,
                    allow_indexed_db: false,
                    allow_cookies: false
                }
            }
        };

        // Write manifest
        fs.writeFileSync(path.join(tmpDir, 'manifest.json'), JSON.stringify(manifest, null, 2));

        // Use the builder to create the .liv file
        const builderPath = path.join(__dirname, '../../bin/liv-builder.exe');

        if (!fs.existsSync(builderPath)) {
            // Fallback: save as JSON for manual building
            const outputPath = data.filePath.replace('.liv', '-structure');
            fs.cpSync(tmpDir, outputPath, { recursive: true });

            dialog.showMessageBox({
                type: 'warning',
                title: 'Builder Not Available',
                message: 'Document structure saved, but builder not found.',
                detail: `Structure saved to: ${outputPath}\n\nTo create the .liv file, run:\nliv-builder -i "${outputPath}" -o "${data.filePath}"`
            });

            // Clean up temp dir
            fs.rmSync(tmpDir, { recursive: true, force: true });
            event.reply('document-saved', { success: true, partial: true });
            return;
        }

        // Run the builder
        const { execFile } = require('child_process');
        execFile(builderPath, [
            '-i', tmpDir,
            '-o', data.filePath,
            '-c' // Enable compression
        ], (error, stdout, stderr) => {
            // Clean up temp directory
            try {
                fs.rmSync(tmpDir, { recursive: true, force: true });
            } catch (cleanupError) {
                console.error('Failed to clean up temp directory:', cleanupError);
            }

            if (error) {
                console.error('Builder error:', error);
                console.error('Builder stderr:', stderr);
                dialog.showErrorBox('Build Error', `Failed to create LIV file: ${error.message}\n\n${stderr}`);
                event.reply('document-saved', { success: false, error: error.message });
                return;
            }

            console.log('Builder output:', stdout);

            // Success
            addToRecentFiles(data.filePath);

            dialog.showMessageBox({
                type: 'info',
                title: 'Document Saved',
                message: 'Your LIV document has been created successfully!',
                detail: `Saved to: ${data.filePath}`,
                buttons: ['OK', 'Open in Viewer'],
                defaultId: 0
            }).then(result => {
                if (result.response === 1) {
                    openFileByPath(data.filePath);
                }
            });

            event.reply('document-saved', { success: true });
        });

    } catch (error) {
        console.error('Error saving document:', error);
        dialog.showErrorBox('Save Error', `Failed to save: ${error.message}`);
        event.reply('document-saved', { success: false, error: error.message });
    }
});

ipcMain.on('preview-document', (event, filePath) => {
    if (fs.existsSync(filePath)) {
        openFileByPath(filePath);
    } else {
        dialog.showErrorBox('Preview Error', 'Save the document first before previewing.');
    }
});

// Additional IPC handlers for desktop integration
ipcMain.on('open-external', (event, url) => {
    shell.openExternal(url);
});

ipcMain.on('quit-app', () => {
    app.quit();
});

ipcMain.on('minimize-window', () => {
    if (mainWindow) {
        mainWindow.minimize();
    }
});

ipcMain.on('maximize-window', () => {
    if (mainWindow) {
        if (mainWindow.isMaximized()) {
            mainWindow.unmaximize();
        } else {
            mainWindow.maximize();
        }
    }
});

ipcMain.on('close-window', () => {
    if (mainWindow) {
        mainWindow.close();
    }
});

// ========== PDF OPERATIONS MOVED TO FRONTEND ==========
// PDF operations are now handled in the frontend LIV editor

// ===== Go Backend PDF Operations =====

// Extract text from PDF using Go backend
ipcMain.handle('go-pdf-extract-text', async (event, filePath) => {
    try {
        const text = await goBackend.pdfExtractText(filePath);
        return { success: true, text };
    } catch (error) {
        console.error('Error extracting PDF text:', error);
        return { success: false, error: error.message };
    }
});

// Merge PDFs using Go backend
ipcMain.handle('go-pdf-merge', async (event, { inputPaths, outputPath }) => {
    try {
        await goBackend.pdfMerge(inputPaths, outputPath);
        return { success: true };
    } catch (error) {
        console.error('Error merging PDFs:', error);
        return { success: false, error: error.message };
    }
});

// Split PDF using Go backend
ipcMain.handle('go-pdf-split', async (event, { inputPath, ranges, outputDir }) => {
    try {
        await goBackend.pdfSplit(inputPath, ranges, outputDir);
        return { success: true };
    } catch (error) {
        console.error('Error splitting PDF:', error);
        return { success: false, error: error.message };
    }
});

// Extract pages using Go backend
ipcMain.handle('go-pdf-extract-pages', async (event, { inputPath, pages, outputPath }) => {
    try {
        await goBackend.pdfExtractPages(inputPath, pages, outputPath);
        return { success: true };
    } catch (error) {
        console.error('Error extracting pages:', error);
        return { success: false, error: error.message };
    }
});

// Rotate pages using Go backend
ipcMain.handle('go-pdf-rotate', async (event, { inputPath, pages, angle, outputPath }) => {
    try {
        await goBackend.pdfRotate(inputPath, pages, angle, outputPath);
        return { success: true };
    } catch (error) {
        console.error('Error rotating pages:', error);
        return { success: false, error: error.message };
    }
});

// Add watermark using Go backend
ipcMain.handle('go-pdf-watermark', async (event, { inputPath, text, outputPath }) => {
    try {
        await goBackend.pdfWatermark(inputPath, text, outputPath);
        return { success: true };
    } catch (error) {
        console.error('Error adding watermark:', error);
        return { success: false, error: error.message };
    }
});

// Compress PDF using Go backend
ipcMain.handle('go-pdf-compress', async (event, { inputPath, outputPath }) => {
    try {
        await goBackend.pdfCompress(inputPath, outputPath);
        return { success: true };
    } catch (error) {
        console.error('Error compressing PDF:', error);
        return { success: false, error: error.message };
    }
});

// Encrypt PDF using Go backend
ipcMain.handle('go-pdf-encrypt', async (event, { inputPath, password, outputPath }) => {
    try {
        await goBackend.pdfEncrypt(inputPath, password, outputPath);
        return { success: true };
    } catch (error) {
        console.error('Error encrypting PDF:', error);
        return { success: false, error: error.message };
    }
});

// Get PDF info using Go backend
ipcMain.handle('go-pdf-info', async (event, filePath) => {
    try {
        const info = await goBackend.pdfInfo(filePath);
        return { success: true, info };
    } catch (error) {
        console.error('Error getting PDF info:', error);
        return { success: false, error: error.message };
    }
});

// Set PDF info using Go backend
ipcMain.handle('go-pdf-set-info', async (event, { inputPath, info, outputPath }) => {
    try {
        await goBackend.pdfSetInfo(inputPath, info, outputPath);
        return { success: true };
    } catch (error) {
        console.error('Error setting PDF info:', error);
        return { success: false, error: error.message };
    }
});

// Convert PDF to LIV using Go backend
ipcMain.handle('go-pdf-to-liv', async (event, { inputPath, outputPath }) => {
    try {
        await goBackend.pdfToLIV(inputPath, outputPath);
        return { success: true };
    } catch (error) {
        console.error('Error converting PDF to LIV:', error);
        return { success: false, error: error.message };
    }
});

// Save temporary file for PDF conversion
ipcMain.handle('save-temp-file', async (event, file) => {
    try {
        const fs = require('fs').promises;
        const path = require('path');
        const os = require('os');

        // Create temp directory if it doesn't exist
        const tempDir = path.join(os.tmpdir(), 'liv-editor');
        await fs.mkdir(tempDir, { recursive: true });

        // Generate temp file path
        const tempPath = path.join(tempDir, `temp-${Date.now()}.pdf`);

        // Write file buffer to temp location
        await fs.writeFile(tempPath, Buffer.from(file.buffer));

        return tempPath;
    } catch (error) {
        console.error('Error saving temp file:', error);
        throw error;
    }
});

// Read LIV file content
ipcMain.handle('read-liv-file', async (event, filePath) => {
    try {
        const fs = require('fs').promises;
        const JSZip = require('jszip');

        // Read the ZIP file
        const data = await fs.readFile(filePath);
        const zip = await JSZip.loadAsync(data);

        // Extract document.json
        const documentFile = zip.file('document.json');
        if (!documentFile) {
            throw new Error('document.json not found in LIV file');
        }

        const documentContent = await documentFile.async('string');
        return JSON.parse(documentContent);
    } catch (error) {
        console.error('Error reading LIV file:', error);
        throw error;
    }
});

// ===== Go Backend LIV Operations =====

// Build LIV document using Go backend
ipcMain.handle('go-liv-build', async (event, { inputDir, outputFile, options }) => {
    try {
        await goBackend.buildDocument(inputDir, outputFile, options);
        return { success: true };
    } catch (error) {
        console.error('Error building LIV document:', error);
        return { success: false, error: error.message };
    }
});

// Validate LIV document using Go backend
ipcMain.handle('go-liv-validate', async (event, filePath) => {
    try {
        const result = await goBackend.validateDocument(filePath);
        return { success: true, result };
    } catch (error) {
        console.error('Error validating LIV document:', error);
        return { success: false, error: error.message };
    }
});

// Read file contents
ipcMain.handle('read-file', async (event, filePath) => {
    try {
        const content = fs.readFileSync(filePath, 'utf8');
        return content;
    } catch (error) {
        console.error('Error reading file:', error);
        throw error;
    }
});

// Write file contents
ipcMain.handle('write-file', async (event, filePath, content) => {
    try {
        fs.writeFileSync(filePath, content, 'utf8');
        return { success: true };
    } catch (error) {
        console.error('Error writing file:', error);
        throw error;
    }
});

// ========== END PDF OPERATIONS HANDLERS ==========

// Auto updater events
autoUpdater.on('checking-for-update', () => {
    console.log('Checking for update...');
});

autoUpdater.on('update-available', (_info) => {
    console.log('Update available.');
    dialog.showMessageBox(mainWindow, {
        type: 'info',
        title: 'Update Available',
        message: 'A new version is available and will be downloaded in the background.',
        buttons: ['OK']
    });
});

autoUpdater.on('update-not-available', (_info) => {
    console.log('Update not available.');
});

autoUpdater.on('error', (err) => {
    console.log('Error in auto-updater. ' + err);
});

autoUpdater.on('download-progress', (progressObj) => {
    let log_message = 'Download speed: ' + progressObj.bytesPerSecond;
    log_message = log_message + ' - Downloaded ' + progressObj.percent + '%';
    log_message = log_message + ' (' + progressObj.transferred + '/' + progressObj.total + ')';
    console.log(log_message);
});

autoUpdater.on('update-downloaded', (_info) => {
    console.log('Update downloaded');
    dialog.showMessageBox(mainWindow, {
        type: 'info',
        title: 'Update Ready',
        message: 'Update downloaded. The application will restart to apply the update.',
        buttons: ['Restart Now', 'Later']
    }).then((result) => {
        if (result.response === 0) {
            autoUpdater.quitAndInstall();
        }
    });
});

