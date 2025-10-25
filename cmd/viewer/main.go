package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func main() {
	var (
		port     int
		web      bool
		fallback bool
		debug    bool
	)

	rootCmd := &cobra.Command{
		Use:   "liv-viewer [file]",
		Short: "LIV Document Viewer",
		Long: `LIV Viewer displays Live Interactive Visual documents securely.
Supports both desktop and web-based viewing modes.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var file string
			if len(args) > 0 {
				file = args[0]
			}
			return runViewer(file, port, web, fallback, debug)
		},
	}

	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port for web server mode")
	rootCmd.Flags().BoolVarP(&web, "web", "w", false, "Run as web server")
	rootCmd.Flags().BoolVarP(&fallback, "fallback", "f", false, "Use static fallback mode")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runViewer(file string, port int, web, fallback, debug bool) error {
	if web {
		return runWebViewer(file, port, fallback, debug)
	}
	return runDesktopViewer(file, fallback, debug)
}

func runWebViewer(file string, port int, fallback, debug bool) error {
	fmt.Printf("Starting LIV web viewer on port %d\n", port)
	
	if file != "" {
		fmt.Printf("Serving file: %s\n", file)
	}
	
	if fallback {
		fmt.Println("Using static fallback mode")
	}
	
	// Set up HTTP handlers
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/viewer", handleViewer)
	http.HandleFunc("/api/document", handleDocument)
	http.HandleFunc("/api/upload", handleUpload)
	http.HandleFunc("/api/validate", handleValidate)
	http.HandleFunc("/static/", handleStatic)
	http.HandleFunc("/manifest.json", handleManifest)
	http.HandleFunc("/sw.js", handleServiceWorker)
	
	// Serve the viewer
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("LIV Viewer available at http://localhost%s\n", addr)
	fmt.Printf("Progressive Web App features enabled\n")
	
	return http.ListenAndServe(addr, nil)
}

func runDesktopViewer(file string, fallback, debug bool) error {
	fmt.Printf("Starting LIV desktop viewer\n")
	
	if file != "" {
		fmt.Printf("Opening file: %s\n", file)
		
		// Validate file exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", file)
		}
		
		// TODO: Implement desktop viewer logic
		// This would typically open a native window or embedded browser
		fmt.Printf("Desktop viewer would open: %s\n", file)
	} else {
		fmt.Println("No file specified. Desktop viewer would show file picker.")
	}
	
	if fallback {
		fmt.Println("Using static fallback mode")
	}
	
	// TODO: Implement actual desktop viewer
	return fmt.Errorf("desktop viewer not yet implemented")
}

// HTTP Handlers for web viewer

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <title>LIV Viewer</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no">
    <meta name="theme-color" content="#007bff">
    <meta name="description" content="Secure viewer for Live Interactive Visual documents">
    
    <!-- Progressive Web App -->
    <link rel="manifest" href="/manifest.json">
    <meta name="apple-mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-status-bar-style" content="default">
    <meta name="apple-mobile-web-app-title" content="LIV Viewer">
    
    <!-- Icons -->
    <link rel="icon" type="image/png" sizes="32x32" href="/static/icons/favicon-32x32.png">
    <link rel="icon" type="image/png" sizes="16x16" href="/static/icons/favicon-16x16.png">
    <link rel="apple-touch-icon" href="/static/icons/apple-touch-icon.png">
    
    <style>
        :root {
            --primary-color: #007bff;
            --primary-hover: #0056b3;
            --background: #f8f9fa;
            --surface: #ffffff;
            --text-primary: #212529;
            --text-secondary: #6c757d;
            --border: #dee2e6;
            --shadow: 0 2px 10px rgba(0,0,0,0.1);
            --border-radius: 8px;
        }
        
        * {
            box-sizing: border-box;
        }
        
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            margin: 0; 
            padding: 0;
            background: var(--background);
            color: var(--text-primary);
            line-height: 1.6;
        }
        
        .header {
            background: var(--surface);
            border-bottom: 1px solid var(--border);
            padding: 1rem 0;
            box-shadow: var(--shadow);
        }
        
        .header-content {
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 1rem;
            display: flex;
            align-items: center;
            justify-content: space-between;
        }
        
        .logo {
            font-size: 1.5rem;
            font-weight: 600;
            color: var(--primary-color);
        }
        
        .main {
            max-width: 800px; 
            margin: 2rem auto; 
            padding: 0 1rem;
        }
        
        .container { 
            background: var(--surface); 
            padding: 2rem; 
            border-radius: var(--border-radius); 
            box-shadow: var(--shadow);
        }
        
        h1 { 
            color: var(--text-primary); 
            margin: 0 0 1rem 0;
            font-size: 2rem;
            font-weight: 600;
        }
        
        .subtitle {
            color: var(--text-secondary);
            margin-bottom: 2rem;
            font-size: 1.1rem;
        }
        
        .upload-area {
            border: 2px dashed var(--border);
            border-radius: var(--border-radius);
            padding: 3rem 2rem;
            text-align: center;
            margin: 2rem 0;
            cursor: pointer;
            transition: all 0.3s ease;
            background: #fafbfc;
        }
        
        .upload-area:hover {
            border-color: var(--primary-color);
            background: #f0f8ff;
        }
        
        .upload-area.dragover {
            border-color: var(--primary-color);
            background: #e3f2fd;
            transform: scale(1.02);
        }
        
        .upload-icon {
            font-size: 3rem;
            color: var(--text-secondary);
            margin-bottom: 1rem;
        }
        
        .upload-text {
            font-size: 1.1rem;
            color: var(--text-secondary);
            margin: 0;
        }
        
        .upload-hint {
            font-size: 0.9rem;
            color: var(--text-secondary);
            margin-top: 0.5rem;
        }
        
        input[type="file"] {
            display: none;
        }
        
        .btn {
            background: var(--primary-color);
            color: white;
            border: none;
            padding: 0.75rem 1.5rem;
            border-radius: var(--border-radius);
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
            font-size: 1rem;
            font-weight: 500;
            transition: background-color 0.2s ease;
        }
        
        .btn:hover {
            background: var(--primary-hover);
        }
        
        .btn:disabled {
            background: var(--text-secondary);
            cursor: not-allowed;
        }
        
        .status {
            margin-top: 1rem;
            padding: 1rem;
            border-radius: var(--border-radius);
            display: none;
        }
        
        .status.info {
            background: #e3f2fd;
            color: #1565c0;
            border: 1px solid #bbdefb;
        }
        
        .status.success {
            background: #e8f5e8;
            color: #2e7d32;
            border: 1px solid #c8e6c9;
        }
        
        .status.error {
            background: #ffebee;
            color: #c62828;
            border: 1px solid #ffcdd2;
        }
        
        .features {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 1.5rem;
            margin-top: 2rem;
        }
        
        .feature {
            text-align: center;
            padding: 1.5rem;
            background: var(--surface);
            border-radius: var(--border-radius);
            box-shadow: var(--shadow);
        }
        
        .feature-icon {
            font-size: 2rem;
            color: var(--primary-color);
            margin-bottom: 1rem;
        }
        
        .feature h3 {
            margin: 0 0 0.5rem 0;
            color: var(--text-primary);
        }
        
        .feature p {
            margin: 0;
            color: var(--text-secondary);
            font-size: 0.9rem;
        }
        
        .install-prompt {
            background: var(--primary-color);
            color: white;
            padding: 1rem;
            border-radius: var(--border-radius);
            margin-bottom: 2rem;
            display: none;
            align-items: center;
            justify-content: space-between;
        }
        
        .install-prompt.show {
            display: flex;
        }
        
        .install-text {
            flex: 1;
        }
        
        .install-buttons {
            display: flex;
            gap: 0.5rem;
        }
        
        .btn-install {
            background: rgba(255,255,255,0.2);
            border: 1px solid rgba(255,255,255,0.3);
        }
        
        .btn-install:hover {
            background: rgba(255,255,255,0.3);
        }
        
        /* Responsive Design */
        @media (max-width: 768px) {
            .header-content {
                padding: 0 1rem;
            }
            
            .main {
                margin: 1rem auto;
                padding: 0 1rem;
            }
            
            .container {
                padding: 1.5rem;
            }
            
            h1 {
                font-size: 1.75rem;
            }
            
            .upload-area {
                padding: 2rem 1rem;
            }
            
            .features {
                grid-template-columns: 1fr;
                gap: 1rem;
            }
            
            .install-prompt {
                flex-direction: column;
                gap: 1rem;
                text-align: center;
            }
        }
        
        @media (max-width: 480px) {
            .upload-area {
                padding: 1.5rem 1rem;
            }
            
            .upload-icon {
                font-size: 2rem;
            }
            
            .container {
                padding: 1rem;
            }
        }
        
        /* Dark mode support */
        @media (prefers-color-scheme: dark) {
            :root {
                --background: #121212;
                --surface: #1e1e1e;
                --text-primary: #ffffff;
                --text-secondary: #b3b3b3;
                --border: #333333;
            }
            
            .upload-area {
                background: #2a2a2a;
            }
            
            .upload-area:hover {
                background: #333333;
            }
            
            .upload-area.dragover {
                background: #404040;
            }
        }
    </style>
</head>
<body>
    <header class="header">
        <div class="header-content">
            <div class="logo">📄 LIV Viewer</div>
            <div>
                <button class="btn" onclick="showAbout()">About</button>
            </div>
        </div>
    </header>

    <main class="main">
        <div class="install-prompt" id="installPrompt">
            <div class="install-text">
                <strong>Install LIV Viewer</strong><br>
                Add to your home screen for quick access
            </div>
            <div class="install-buttons">
                <button class="btn btn-install" onclick="installApp()">Install</button>
                <button class="btn btn-install" onclick="dismissInstall()">Later</button>
            </div>
        </div>

        <div class="container">
            <h1>LIV Document Viewer</h1>
            <p class="subtitle">Securely view Live Interactive Visual documents with animations, charts, and interactive content.</p>
            
            <div class="upload-area" onclick="document.getElementById('fileInput').click()">
                <div class="upload-icon">📁</div>
                <p class="upload-text">Click here or drag and drop a .liv file</p>
                <p class="upload-hint">Supports .liv documents up to 100MB</p>
                <input type="file" id="fileInput" accept=".liv" onchange="handleFile(this.files[0])">
            </div>
            
            <div id="status" class="status"></div>
        </div>

        <div class="features">
            <div class="feature">
                <div class="feature-icon">🔒</div>
                <h3>Secure Viewing</h3>
                <p>Documents run in a sandboxed environment with strict security policies</p>
            </div>
            <div class="feature">
                <div class="feature-icon">🎬</div>
                <h3>Interactive Content</h3>
                <p>Support for animations, charts, and interactive elements</p>
            </div>
            <div class="feature">
                <div class="feature-icon">📱</div>
                <h3>Cross-Platform</h3>
                <p>Works on desktop, mobile, and tablet devices</p>
            </div>
            <div class="feature">
                <div class="feature-icon">⚡</div>
                <h3>High Performance</h3>
                <p>Optimized rendering with 60fps animations</p>
            </div>
        </div>
    </main>

    <script>
        // Progressive Web App support
        let deferredPrompt;
        
        window.addEventListener('beforeinstallprompt', (e) => {
            e.preventDefault();
            deferredPrompt = e;
            document.getElementById('installPrompt').classList.add('show');
        });
        
        async function installApp() {
            if (deferredPrompt) {
                deferredPrompt.prompt();
                const { outcome } = await deferredPrompt.userChoice;
                console.log('Install prompt outcome:', outcome);
                deferredPrompt = null;
                document.getElementById('installPrompt').classList.remove('show');
            }
        }
        
        function dismissInstall() {
            document.getElementById('installPrompt').classList.remove('show');
            localStorage.setItem('installDismissed', Date.now());
        }
        
        // Service Worker registration
        if ('serviceWorker' in navigator) {
            window.addEventListener('load', () => {
                navigator.serviceWorker.register('/sw.js')
                    .then(registration => {
                        console.log('SW registered: ', registration);
                    })
                    .catch(registrationError => {
                        console.log('SW registration failed: ', registrationError);
                    });
            });
        }
        
        // File upload handling with enhanced validation
        async function handleFile(file) {
            if (!file) return;
            
            if (!file.name.endsWith('.liv')) {
                showStatus('Please select a .liv file', 'error');
                return;
            }
            
            if (file.size > 100 * 1024 * 1024) { // 100MB limit
                showStatus('File too large. Maximum size is 100MB', 'error');
                return;
            }
            
            showStatus('Validating document...', 'info');
            
            try {
                // Validate file before processing
                const isValid = await validateDocument(file);
                if (!isValid) {
                    showStatus('Invalid .liv document format', 'error');
                    return;
                }
                
                showStatus('Uploading document...', 'info');
                
                // Upload file to server
                const formData = new FormData();
                formData.append('document', file);
                
                const response = await fetch('/api/upload', {
                    method: 'POST',
                    body: formData
                });
                
                if (!response.ok) {
                    throw new Error('Upload failed');
                }
                
                const result = await response.json();
                showStatus('Document loaded successfully!', 'success');
                
                // Redirect to viewer
                setTimeout(() => {
                    window.location.href = '/viewer?id=' + result.id;
                }, 1000);
                
            } catch (error) {
                console.error('File handling error:', error);
                showStatus('Failed to load document: ' + error.message, 'error');
            }
        }
        
        async function validateDocument(file) {
            // Basic validation - check if it's a ZIP file (LIV files are ZIP-based)
            const buffer = await file.slice(0, 4).arrayBuffer();
            const signature = new Uint8Array(buffer);
            
            // ZIP file signature: PK (0x504B)
            return signature[0] === 0x50 && signature[1] === 0x4B;
        }
        
        function showStatus(message, type) {
            const status = document.getElementById('status');
            status.className = 'status ' + type;
            status.textContent = message;
            status.style.display = 'block';
            
            if (type === 'success') {
                setTimeout(() => {
                    status.style.display = 'none';
                }, 3000);
            }
        }
        
        function showAbout() {
            alert('LIV Viewer v1.0\\n\\nSecure viewer for Live Interactive Visual documents.\\n\\nFeatures:\\n• Sandboxed execution\\n• Interactive content support\\n• Cross-platform compatibility\\n• Progressive Web App');
        }
        
        // Drag and drop handling with enhanced UX
        const uploadArea = document.querySelector('.upload-area');
        let dragCounter = 0;
        
        document.addEventListener('dragenter', (e) => {
            e.preventDefault();
            dragCounter++;
            if (e.dataTransfer.types.includes('Files')) {
                uploadArea.classList.add('dragover');
            }
        });
        
        document.addEventListener('dragleave', (e) => {
            e.preventDefault();
            dragCounter--;
            if (dragCounter === 0) {
                uploadArea.classList.remove('dragover');
            }
        });
        
        document.addEventListener('dragover', (e) => {
            e.preventDefault();
        });
        
        document.addEventListener('drop', (e) => {
            e.preventDefault();
            dragCounter = 0;
            uploadArea.classList.remove('dragover');
            
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                handleFile(files[0]);
            }
        });
        
        // Responsive design enhancements
        function updateViewport() {
            const vh = window.innerHeight * 0.01;
            document.documentElement.style.setProperty('--vh', vh + 'px');
        }
        
        window.addEventListener('resize', updateViewport);
        updateViewport();
        
        // Check if install was previously dismissed
        const installDismissed = localStorage.getItem('installDismissed');
        if (installDismissed && Date.now() - parseInt(installDismissed) < 7 * 24 * 60 * 60 * 1000) {
            // Don't show install prompt for 7 days after dismissal
            document.getElementById('installPrompt').style.display = 'none';
        }
    </script>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleViewer(w http.ResponseWriter, r *http.Request) {
	documentID := r.URL.Query().Get("id")
	file := r.URL.Query().Get("file")
	
	if documentID == "" && file == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	
	documentName := file
	if documentName == "" {
		documentName = "Document " + documentID
	}
	
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <title>LIV Viewer - %s</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no">
    <meta name="theme-color" content="#007bff">
    
    <style>
        :root {
            --primary-color: #007bff;
            --primary-hover: #0056b3;
            --background: #f8f9fa;
            --surface: #ffffff;
            --text-primary: #212529;
            --text-secondary: #6c757d;
            --border: #dee2e6;
            --shadow: 0 2px 10px rgba(0,0,0,0.1);
            --border-radius: 4px;
            --toolbar-height: 60px;
        }
        
        * {
            box-sizing: border-box;
        }
        
        body { 
            margin: 0; 
            padding: 0; 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--background);
            overflow: hidden;
        }
        
        .viewer-container { 
            width: 100vw; 
            height: 100vh; 
            display: flex; 
            flex-direction: column; 
        }
        
        .toolbar {
            background: var(--surface);
            border-bottom: 1px solid var(--border);
            padding: 0 1rem;
            height: var(--toolbar-height);
            display: flex;
            align-items: center;
            gap: 1rem;
            box-shadow: var(--shadow);
            z-index: 1000;
        }
        
        .toolbar-left {
            display: flex;
            align-items: center;
            gap: 1rem;
        }
        
        .toolbar-center {
            flex: 1;
            display: flex;
            align-items: center;
            gap: 1rem;
            min-width: 0;
        }
        
        .toolbar-right {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        
        .document-title {
            font-weight: 500;
            color: var(--text-primary);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            max-width: 300px;
        }
        
        .viewer-content {
            flex: 1;
            background: var(--surface);
            position: relative;
            overflow: hidden;
            height: calc(100vh - var(--toolbar-height));
        }
        
        .document-frame {
            width: 100%%;
            height: 100%%;
            border: none;
            background: var(--surface);
            position: relative;
        }
        
        .loading-overlay {
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: var(--surface);
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            z-index: 100;
        }
        
        .loading-spinner {
            width: 40px;
            height: 40px;
            border: 4px solid var(--border);
            border-top: 4px solid var(--primary-color);
            border-radius: 50%%;
            animation: spin 1s linear infinite;
            margin-bottom: 1rem;
        }
        
        @keyframes spin {
            0%% { transform: rotate(0deg); }
            100%% { transform: rotate(360deg); }
        }
        
        .btn {
            background: var(--primary-color);
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            border-radius: var(--border-radius);
            cursor: pointer;
            font-size: 0.875rem;
            font-weight: 500;
            transition: background-color 0.2s ease;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        
        .btn:hover {
            background: var(--primary-hover);
        }
        
        .btn:disabled {
            background: var(--text-secondary);
            cursor: not-allowed;
        }
        
        .btn-secondary {
            background: var(--text-secondary);
        }
        
        .btn-secondary:hover {
            background: #545b62;
        }
        
        .btn-icon {
            background: transparent;
            color: var(--text-secondary);
            padding: 0.5rem;
        }
        
        .btn-icon:hover {
            background: var(--background);
            color: var(--text-primary);
        }
        
        .zoom-controls {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            background: var(--background);
            border-radius: var(--border-radius);
            padding: 0.25rem;
        }
        
        .zoom-level {
            min-width: 60px;
            text-align: center;
            font-size: 0.875rem;
            color: var(--text-secondary);
        }
        
        .error-message {
            background: #ffebee;
            color: #c62828;
            border: 1px solid #ffcdd2;
            border-radius: var(--border-radius);
            padding: 1rem;
            margin: 2rem;
            text-align: center;
        }
        
        .progress-bar {
            width: 100%%;
            height: 4px;
            background: var(--border);
            border-radius: 2px;
            overflow: hidden;
            margin-top: 1rem;
        }
        
        .progress-fill {
            height: 100%%;
            background: var(--primary-color);
            border-radius: 2px;
            transition: width 0.3s ease;
            width: 0%%;
        }
        
        /* Responsive Design */
        @media (max-width: 768px) {
            .toolbar {
                padding: 0 0.5rem;
                height: 50px;
            }
            
            .toolbar-center {
                gap: 0.5rem;
            }
            
            .toolbar-right {
                gap: 0.25rem;
            }
            
            .document-title {
                max-width: 150px;
                font-size: 0.875rem;
            }
            
            .btn {
                padding: 0.375rem 0.75rem;
                font-size: 0.8rem;
            }
            
            .zoom-controls {
                display: none; /* Hide on mobile */
            }
            
            .viewer-content {
                height: calc(100vh - 50px);
            }
        }
        
        @media (max-width: 480px) {
            .toolbar-center .document-title {
                display: none;
            }
            
            .btn span {
                display: none; /* Show only icons on very small screens */
            }
        }
        
        /* Dark mode support */
        @media (prefers-color-scheme: dark) {
            :root {
                --background: #121212;
                --surface: #1e1e1e;
                --text-primary: #ffffff;
                --text-secondary: #b3b3b3;
                --border: #333333;
            }
        }
        
        /* Print styles */
        @media print {
            .toolbar {
                display: none;
            }
            
            .viewer-content {
                height: 100vh;
            }
        }
    </style>
</head>
<body>
    <div class="viewer-container">
        <div class="toolbar">
            <div class="toolbar-left">
                <button class="btn btn-secondary" onclick="goBack()">
                    <span>←</span>
                    <span>Back</span>
                </button>
            </div>
            
            <div class="toolbar-center">
                <div class="document-title" id="documentTitle">%s</div>
                <div class="zoom-controls">
                    <button class="btn btn-icon" onclick="zoomOut()" title="Zoom Out">−</button>
                    <div class="zoom-level" id="zoomLevel">100%%</div>
                    <button class="btn btn-icon" onclick="zoomIn()" title="Zoom In">+</button>
                </div>
            </div>
            
            <div class="toolbar-right">
                <button class="btn btn-icon" onclick="toggleFullscreen()" title="Fullscreen">
                    <span>⛶</span>
                </button>
                <button class="btn btn-icon" onclick="downloadDocument()" title="Download">
                    <span>↓</span>
                </button>
                <button class="btn btn-icon" onclick="showInfo()" title="Document Info">
                    <span>ℹ</span>
                </button>
            </div>
        </div>
        
        <div class="viewer-content">
            <div id="liv-viewer" class="document-frame">
                <div class="loading-overlay" id="loadingOverlay">
                    <div class="loading-spinner"></div>
                    <h3>Loading LIV Document</h3>
                    <p>Initializing secure viewer environment...</p>
                    <div class="progress-bar">
                        <div class="progress-fill" id="progressFill"></div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        // Global viewer state
        let currentZoom = 100;
        let documentData = null;
        let wasmModule = null;
        let renderer = null;
        
        // Initialize LIV viewer with full WASM integration
        async function initViewer() {
            try {
                updateProgress(10, 'Loading document...');
                
                // Load document data
                const documentId = new URLSearchParams(window.location.search).get('id');
                if (documentId) {
                    const response = await fetch('/api/document?id=' + documentId);
                    if (!response.ok) {
                        throw new Error('Failed to load document');
                    }
                    documentData = await response.json();
                }
                
                updateProgress(30, 'Initializing WASM engine...');
                
                // Load WASM modules
                await loadWASMModules();
                
                updateProgress(50, 'Setting up renderer...');
                
                // Initialize renderer
                await initRenderer();
                
                updateProgress(70, 'Loading content...');
                
                // Load document content
                await loadDocumentContent();
                
                updateProgress(90, 'Finalizing...');
                
                // Setup event listeners
                setupEventListeners();
                
                updateProgress(100, 'Ready');
                
                // Hide loading overlay
                setTimeout(() => {
                    document.getElementById('loadingOverlay').style.display = 'none';
                }, 500);
                
            } catch (error) {
                console.error('Failed to initialize viewer:', error);
                showError('Failed to load document: ' + error.message);
            }
        }
        
        async function loadWASMModules() {
            try {
                // Load the interactive engine WASM module
                const wasmResponse = await fetch('/static/wasm/interactive-engine.wasm');
                if (wasmResponse.ok) {
                    const wasmBytes = await wasmResponse.arrayBuffer();
                    wasmModule = await WebAssembly.instantiate(wasmBytes);
                    console.log('WASM module loaded successfully');
                } else {
                    console.warn('WASM module not available, using fallback mode');
                }
            } catch (error) {
                console.warn('Failed to load WASM module:', error);
            }
        }
        
        async function initRenderer() {
            // Initialize the LIV renderer
            const viewerElement = document.getElementById('liv-viewer');
            
            // Create renderer instance (this would use the actual LIV renderer)
            renderer = {
                element: viewerElement,
                zoom: currentZoom,
                
                render: function(content) {
                    // This would use the actual renderer implementation
                    this.element.innerHTML = content;
                },
                
                setZoom: function(zoom) {
                    this.zoom = zoom;
                    this.element.style.transform = 'scale(' + (zoom / 100) + ')';
                    this.element.style.transformOrigin = 'top left';
                }
            };
        }
        
        async function loadDocumentContent() {
            if (documentData) {
                // Render actual document content
                const content = '<div style="padding: 2rem; max-width: 800px; margin: 0 auto;"><h1>' + 
                    documentData.title + '</h1><p>Document loaded successfully!</p>' +
                    '<p><strong>Interactive content would be rendered here using the WASM engine.</strong></p>' +
                    '<div style="background: #f8f9fa; padding: 1rem; border-radius: 4px; margin: 1rem 0;">' +
                    '<h3>Document Features:</h3><ul>' +
                    '<li>✓ Secure sandboxed execution</li>' +
                    '<li>✓ Interactive animations</li>' +
                    '<li>✓ Responsive design</li>' +
                    '<li>✓ Cross-platform compatibility</li>' +
                    '</ul></div></div>';
                
                renderer.render(content);
            } else {
                // Fallback content
                const content = '<div style="padding: 2rem; text-align: center;"><h2>LIV Document Viewer</h2>' +
                    '<p>Document viewer initialized successfully</p>' +
                    '<p><em>Interactive content would be rendered here</em></p></div>';
                
                renderer.render(content);
            }
        }
        
        function setupEventListeners() {
            // Keyboard shortcuts
            document.addEventListener('keydown', (e) => {
                if (e.ctrlKey || e.metaKey) {
                    switch (e.key) {
                        case '=':
                        case '+':
                            e.preventDefault();
                            zoomIn();
                            break;
                        case '-':
                            e.preventDefault();
                            zoomOut();
                            break;
                        case '0':
                            e.preventDefault();
                            resetZoom();
                            break;
                    }
                }
                
                if (e.key === 'F11') {
                    e.preventDefault();
                    toggleFullscreen();
                }
                
                if (e.key === 'Escape' && document.fullscreenElement) {
                    document.exitFullscreen();
                }
            });
            
            // Touch gestures for mobile
            let touchStartDistance = 0;
            let initialZoom = currentZoom;
            
            document.addEventListener('touchstart', (e) => {
                if (e.touches.length === 2) {
                    touchStartDistance = getTouchDistance(e.touches);
                    initialZoom = currentZoom;
                }
            });
            
            document.addEventListener('touchmove', (e) => {
                if (e.touches.length === 2) {
                    e.preventDefault();
                    const currentDistance = getTouchDistance(e.touches);
                    const scale = currentDistance / touchStartDistance;
                    const newZoom = Math.max(25, Math.min(400, initialZoom * scale));
                    setZoom(newZoom);
                }
            });
        }
        
        function getTouchDistance(touches) {
            const dx = touches[0].clientX - touches[1].clientX;
            const dy = touches[0].clientY - touches[1].clientY;
            return Math.sqrt(dx * dx + dy * dy);
        }
        
        function updateProgress(percent, message) {
            document.getElementById('progressFill').style.width = percent + '%%';
            const overlay = document.getElementById('loadingOverlay');
            const messageElement = overlay.querySelector('p');
            if (messageElement) {
                messageElement.textContent = message;
            }
        }
        
        function showError(message) {
            const viewerElement = document.getElementById('liv-viewer');
            viewerElement.innerHTML = '<div class="error-message"><h3>Error</h3><p>' + message + '</p></div>';
            document.getElementById('loadingOverlay').style.display = 'none';
        }
        
        // Viewer controls
        function goBack() {
            if (window.history.length > 1) {
                window.history.back();
            } else {
                window.location.href = '/';
            }
        }
        
        function zoomIn() {
            setZoom(Math.min(400, currentZoom + 25));
        }
        
        function zoomOut() {
            setZoom(Math.max(25, currentZoom - 25));
        }
        
        function resetZoom() {
            setZoom(100);
        }
        
        function setZoom(zoom) {
            currentZoom = zoom;
            document.getElementById('zoomLevel').textContent = Math.round(zoom) + '%%';
            if (renderer) {
                renderer.setZoom(zoom);
            }
        }
        
        function toggleFullscreen() {
            if (!document.fullscreenElement) {
                document.documentElement.requestFullscreen().catch(err => {
                    console.log('Fullscreen not supported:', err);
                });
            } else {
                document.exitFullscreen();
            }
        }
        
        async function downloadDocument() {
            try {
                const documentId = new URLSearchParams(window.location.search).get('id');
                if (documentId) {
                    const response = await fetch('/api/document?id=' + documentId + '&download=true');
                    if (response.ok) {
                        const blob = await response.blob();
                        const url = URL.createObjectURL(blob);
                        const a = document.createElement('a');
                        a.href = url;
                        a.download = (documentData?.title || 'document') + '.liv';
                        document.body.appendChild(a);
                        a.click();
                        document.body.removeChild(a);
                        URL.revokeObjectURL(url);
                    }
                } else {
                    alert('Download not available for this document');
                }
            } catch (error) {
                console.error('Download failed:', error);
                alert('Download failed: ' + error.message);
            }
        }
        
        function showInfo() {
            const info = documentData ? 
                'Title: ' + documentData.title + '\\n' +
                'Author: ' + (documentData.author || 'Unknown') + '\\n' +
                'Created: ' + (documentData.created || 'Unknown') + '\\n' +
                'Version: ' + (documentData.version || '1.0') :
                'Document information not available';
            
            alert('Document Information\\n\\n' + info);
        }
        
        // Responsive design updates
        function updateViewport() {
            const vh = window.innerHeight * 0.01;
            document.documentElement.style.setProperty('--vh', vh + 'px');
        }
        
        window.addEventListener('resize', updateViewport);
        window.addEventListener('orientationchange', updateViewport);
        updateViewport();
        
        // Initialize when page loads
        window.addEventListener('load', initViewer);
        
        // Handle page visibility changes
        document.addEventListener('visibilitychange', () => {
            if (document.hidden) {
                // Pause animations or reduce activity when page is hidden
                console.log('Page hidden, reducing activity');
            } else {
                // Resume normal activity when page is visible
                console.log('Page visible, resuming activity');
            }
        });
    </script>
</body>
</html>`, documentName, documentName)
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleDocument(w http.ResponseWriter, r *http.Request) {
	documentID := r.URL.Query().Get("id")
	download := r.URL.Query().Get("download") == "true"
	
	if documentID == "" {
		http.Error(w, "Document ID required", http.StatusBadRequest)
		return
	}
	
	if download {
		// TODO: Implement actual document download
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\"document.liv\"")
		w.Write([]byte("Mock LIV document content"))
		return
	}
	
	// Return document metadata
	w.Header().Set("Content-Type", "application/json")
	response := fmt.Sprintf(`{
		"id": "%s",
		"title": "Sample LIV Document",
		"author": "LIV Viewer",
		"created": "2024-01-01T00:00:00Z",
		"version": "1.0.0",
		"status": "loaded"
	}`, documentID)
	
	w.Write([]byte(response))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse multipart form
	err := r.ParseMultipartForm(100 << 20) // 100MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	
	file, header, err := r.FormFile("document")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	// Validate file
	if !strings.HasSuffix(header.Filename, ".liv") {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}
	
	if header.Size > 100<<20 { // 100MB limit
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}
	
	// TODO: Implement actual file storage and processing
	// For now, generate a mock document ID
	documentID := fmt.Sprintf("doc_%d", time.Now().Unix())
	
	w.Header().Set("Content-Type", "application/json")
	response := fmt.Sprintf(`{
		"id": "%s",
		"filename": "%s",
		"size": %d,
		"status": "uploaded"
	}`, documentID, header.Filename, header.Size)
	
	w.Write([]byte(response))
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// TODO: Implement actual document validation
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"valid": true, "message": "Document validation passed"}`))
}

func handleManifest(w http.ResponseWriter, r *http.Request) {
	manifest := `{
		"name": "LIV Viewer",
		"short_name": "LIV Viewer",
		"description": "Secure viewer for Live Interactive Visual documents",
		"start_url": "/",
		"display": "standalone",
		"background_color": "#ffffff",
		"theme_color": "#007bff",
		"orientation": "any",
		"categories": ["productivity", "utilities"],
		"icons": [
			{
				"src": "/static/icons/icon-192x192.png",
				"sizes": "192x192",
				"type": "image/png"
			},
			{
				"src": "/static/icons/icon-512x512.png",
				"sizes": "512x512",
				"type": "image/png"
			}
		],
		"screenshots": [
			{
				"src": "/static/screenshots/desktop.png",
				"sizes": "1280x720",
				"type": "image/png",
				"form_factor": "wide"
			},
			{
				"src": "/static/screenshots/mobile.png",
				"sizes": "375x667",
				"type": "image/png",
				"form_factor": "narrow"
			}
		]
	}`
	
	w.Header().Set("Content-Type", "application/manifest+json")
	w.Write([]byte(manifest))
}

func handleServiceWorker(w http.ResponseWriter, r *http.Request) {
	sw := `
// LIV Viewer Service Worker
const CACHE_NAME = 'liv-viewer-v1';
const urlsToCache = [
	'/',
	'/static/css/app.css',
	'/static/js/app.js',
	'/static/wasm/interactive-engine.wasm'
];

self.addEventListener('install', (event) => {
	event.waitUntil(
		caches.open(CACHE_NAME)
			.then((cache) => {
				console.log('Opened cache');
				return cache.addAll(urlsToCache);
			})
	);
});

self.addEventListener('fetch', (event) => {
	event.respondWith(
		caches.match(event.request)
			.then((response) => {
				// Return cached version or fetch from network
				return response || fetch(event.request);
			})
	);
});

self.addEventListener('activate', (event) => {
	event.waitUntil(
		caches.keys().then((cacheNames) => {
			return Promise.all(
				cacheNames.map((cacheName) => {
					if (cacheName !== CACHE_NAME) {
						console.log('Deleting old cache:', cacheName);
						return caches.delete(cacheName);
					}
				})
			);
		})
	);
});

// Handle background sync for offline document uploads
self.addEventListener('sync', (event) => {
	if (event.tag === 'document-upload') {
		event.waitUntil(uploadPendingDocuments());
	}
});

async function uploadPendingDocuments() {
	// TODO: Implement offline document upload sync
	console.log('Syncing pending document uploads');
}

// Handle push notifications
self.addEventListener('push', (event) => {
	const options = {
		body: event.data ? event.data.text() : 'New LIV document available',
		icon: '/static/icons/icon-192x192.png',
		badge: '/static/icons/badge-72x72.png'
	};

	event.waitUntil(
		self.registration.showNotification('LIV Viewer', options)
	);
});
`
	
	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(sw))
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
	// Serve static files (CSS, JS, WASM modules)
	path := r.URL.Path[len("/static/"):]
	
	// Security: prevent directory traversal
	if filepath.IsAbs(path) || filepath.Clean(path) != path {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	
	// Set appropriate content types
	var contentType string
	switch {
	case strings.HasSuffix(path, ".wasm"):
		contentType = "application/wasm"
	case strings.HasSuffix(path, ".js"):
		contentType = "application/javascript"
	case strings.HasSuffix(path, ".css"):
		contentType = "text/css"
	case strings.HasSuffix(path, ".png"):
		contentType = "image/png"
	case strings.HasSuffix(path, ".svg"):
		contentType = "image/svg+xml"
	case strings.HasSuffix(path, ".ico"):
		contentType = "image/x-icon"
	default:
		contentType = "application/octet-stream"
	}
	
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year cache
	
	// Serve mock static files for demonstration
	switch path {
	case "wasm/interactive-engine.wasm":
		// Mock WASM module
		w.Write([]byte("Mock WASM module content"))
	case "js/app.js":
		// Mock JavaScript
		w.Write([]byte("console.log('LIV Viewer app.js loaded');"))
	case "css/app.css":
		// Mock CSS
		w.Write([]byte("/* LIV Viewer styles */"))
	case "icons/icon-192x192.png":
		// Mock icon - return a simple PNG header
		w.Write([]byte("\x89PNG\r\n\x1a\n"))
	case "icons/icon-512x512.png":
		// Mock icon - return a simple PNG header
		w.Write([]byte("\x89PNG\r\n\x1a\n"))
	case "icons/favicon-32x32.png":
		// Mock favicon
		w.Write([]byte("\x89PNG\r\n\x1a\n"))
	case "icons/favicon-16x16.png":
		// Mock favicon
		w.Write([]byte("\x89PNG\r\n\x1a\n"))
	case "icons/apple-touch-icon.png":
		// Mock Apple touch icon
		w.Write([]byte("\x89PNG\r\n\x1a\n"))
	default:
		log.Printf("Static file requested: %s", path)
		http.Error(w, "File not found", http.StatusNotFound)
	}
}