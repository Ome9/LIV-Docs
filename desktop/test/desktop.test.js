const { expect } = require('chai');
const fs = require('fs');
const path = require('path');

describe('Desktop Application Structure', () => {
  const desktopDir = path.join(__dirname, '..');
  
  it('should have required package.json', () => {
    const packagePath = path.join(desktopDir, 'package.json');
    expect(fs.existsSync(packagePath)).to.be.true;
    
    const packageJson = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
    expect(packageJson.name).to.equal('liv-viewer-desktop');
    expect(packageJson.main).to.equal('src/main.js');
  });
  
  it('should have main process file', () => {
    const mainPath = path.join(desktopDir, 'src', 'main.js');
    expect(fs.existsSync(mainPath)).to.be.true;
  });
  
  it('should have preload script', () => {
    const preloadPath = path.join(desktopDir, 'src', 'preload.js');
    expect(fs.existsSync(preloadPath)).to.be.true;
  });
  
  it('should have preferences dialog', () => {
    const prefsPath = path.join(desktopDir, 'src', 'preferences.html');
    expect(fs.existsSync(prefsPath)).to.be.true;
  });
  
  it('should have error page', () => {
    const errorPath = path.join(desktopDir, 'src', 'error.html');
    expect(fs.existsSync(errorPath)).to.be.true;
  });
  
  it('should have build scripts', () => {
    const buildShPath = path.join(desktopDir, 'build.sh');
    const buildBatPath = path.join(desktopDir, 'build.bat');
    
    expect(fs.existsSync(buildShPath)).to.be.true;
    expect(fs.existsSync(buildBatPath)).to.be.true;
  });
  
  it('should have asset directories', () => {
    const iconsPath = path.join(desktopDir, 'assets', 'icons');
    expect(fs.existsSync(iconsPath)).to.be.true;
  });
});

describe('Package.json Configuration', () => {
  let packageJson;
  
  before(() => {
    const packagePath = path.join(__dirname, '..', 'package.json');
    packageJson = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
  });
  
  it('should have correct electron dependency', () => {
    expect(packageJson.devDependencies).to.have.property('electron');
    expect(packageJson.devDependencies).to.have.property('electron-builder');
  });
  
  it('should have required runtime dependencies', () => {
    expect(packageJson.dependencies).to.have.property('electron-updater');
    expect(packageJson.dependencies).to.have.property('electron-store');
    expect(packageJson.dependencies).to.have.property('mime-types');
  });
  
  it('should have build configuration', () => {
    expect(packageJson.build).to.be.an('object');
    expect(packageJson.build.appId).to.equal('com.livformat.viewer');
    expect(packageJson.build.productName).to.equal('LIV Viewer');
  });
  
  it('should have file associations configured', () => {
    expect(packageJson.build.mac.fileAssociations).to.be.an('array');
    expect(packageJson.build.win.fileAssociations).to.be.an('array');
    expect(packageJson.build.linux.fileAssociations).to.be.an('array');
    
    // Check .liv file association
    const macAssoc = packageJson.build.mac.fileAssociations[0];
    expect(macAssoc.ext).to.equal('liv');
    expect(macAssoc.name).to.equal('LIV Document');
  });
  
  it('should have correct build scripts', () => {
    expect(packageJson.scripts).to.have.property('start');
    expect(packageJson.scripts).to.have.property('build');
    expect(packageJson.scripts).to.have.property('build:win');
    expect(packageJson.scripts).to.have.property('build:mac');
    expect(packageJson.scripts).to.have.property('build:linux');
  });
});

describe('Security Configuration', () => {
  it('should have secure webPreferences in main.js', () => {
    const mainPath = path.join(__dirname, '..', 'src', 'main.js');
    const mainContent = fs.readFileSync(mainPath, 'utf8');
    
    // Check for security settings
    expect(mainContent).to.include('nodeIntegration: false');
    expect(mainContent).to.include('contextIsolation: true');
    expect(mainContent).to.include('enableRemoteModule: false');
    expect(mainContent).to.include('webSecurity: true');
    expect(mainContent).to.include('allowRunningInsecureContent: false');
  });
  
  it('should use context bridge in preload script', () => {
    const preloadPath = path.join(__dirname, '..', 'src', 'preload.js');
    const preloadContent = fs.readFileSync(preloadPath, 'utf8');
    
    expect(preloadContent).to.include('contextBridge');
    expect(preloadContent).to.include('exposeInMainWorld');
    
    // Check that Node.js globals are removed
    expect(preloadContent).to.include('delete window.require');
    expect(preloadContent).to.include('delete window.exports');
    expect(preloadContent).to.include('delete window.module');
  });
});

describe('File Association Support', () => {
  it('should handle file opening in main process', () => {
    const mainPath = path.join(__dirname, '..', 'src', 'main.js');
    const mainContent = fs.readFileSync(mainPath, 'utf8');
    
    // Check for file association handlers
    expect(mainContent).to.include("app.on('open-file'");
    expect(mainContent).to.include('.liv');
  });
  
  it('should register custom protocol', () => {
    const mainPath = path.join(__dirname, '..', 'src', 'main.js');
    const mainContent = fs.readFileSync(mainPath, 'utf8');
    
    expect(mainContent).to.include('protocol.registerFileProtocol');
    expect(mainContent).to.include("'liv'");
  });
});

describe('Integration Features', () => {
  it('should have auto-updater configuration', () => {
    const mainPath = path.join(__dirname, '..', 'src', 'main.js');
    const mainContent = fs.readFileSync(mainPath, 'utf8');
    
    expect(mainContent).to.include('electron-updater');
    expect(mainContent).to.include('autoUpdater');
    expect(mainContent).to.include('checkForUpdatesAndNotify');
  });
  
  it('should have settings persistence', () => {
    const mainPath = path.join(__dirname, '..', 'src', 'main.js');
    const mainContent = fs.readFileSync(mainPath, 'utf8');
    
    expect(mainContent).to.include('electron-store');
    expect(mainContent).to.include('Store');
    expect(mainContent).to.include('windowBounds');
    expect(mainContent).to.include('recentFiles');
  });
  
  it('should have native menu integration', () => {
    const mainPath = path.join(__dirname, '..', 'src', 'main.js');
    const mainContent = fs.readFileSync(mainPath, 'utf8');
    
    expect(mainContent).to.include('Menu.buildFromTemplate');
    expect(mainContent).to.include('createMenu');
    expect(mainContent).to.include('accelerator');
  });
});

describe('Cross-Platform Support', () => {
  it('should handle platform-specific paths', () => {
    const mainPath = path.join(__dirname, '..', 'src', 'main.js');
    const mainContent = fs.readFileSync(mainPath, 'utf8');
    
    expect(mainContent).to.include('process.platform');
    expect(mainContent).to.include('darwin');
    expect(mainContent).to.include('win32');
  });
  
  it('should have platform-specific build configurations', () => {
    const packagePath = path.join(__dirname, '..', 'package.json');
    const packageJson = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
    
    expect(packageJson.build).to.have.property('mac');
    expect(packageJson.build).to.have.property('win');
    expect(packageJson.build).to.have.property('linux');
    
    // Check platform-specific icons
    expect(packageJson.build.mac.icon).to.include('.icns');
    expect(packageJson.build.win.icon).to.include('.ico');
    expect(packageJson.build.linux.icon).to.include('.png');
  });
});