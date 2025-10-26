const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

class GoBackend {
    constructor() {
    // Path to Go binaries
        this.binPath = path.join(__dirname, '..', '..', 'bin');
        this.platform = process.platform;

        // Binary names
        this.binaries = {
            cli: this.getBinaryName('liv'),
            pdf: this.getBinaryName('liv-pdf'),
            converter: this.getBinaryName('liv-converter'),
            builder: this.getBinaryName('liv-builder'),
            viewer: this.getBinaryName('liv-viewer'),
            integrity: this.getBinaryName('liv-integrity'),
            validator: this.getBinaryName('liv-manifest-validator'),
            security: this.getBinaryName('liv-security-admin'),
            permission: this.getBinaryName('liv-permission-server'),
            pack: this.getBinaryName('liv-pack')
        };
    }

    getBinaryName(name) {
        return this.platform === 'win32' ? `${name}.exe` : name;
    }

    getBinaryPath(binaryType) {
        const binaryName = this.binaries[binaryType];
        if (!binaryName) {
            throw new Error(`Unknown binary type: ${binaryType}`);
        }
        return path.join(this.binPath, binaryName);
    }

    async execute(binaryType, args = [], options = {}) {
        return new Promise((resolve, reject) => {
            const binaryPath = this.getBinaryPath(binaryType);

            // Check if binary exists
            if (!fs.existsSync(binaryPath)) {
                reject(new Error(`Binary not found: ${binaryPath}. Run build.bat/build.sh first.`));
                return;
            }

            const child = spawn(binaryPath, args, {
                cwd: options.cwd || process.cwd(),
                env: { ...process.env, ...options.env }
            });

            let stdout = '';
            let stderr = '';

            child.stdout.on('data', (data) => {
                stdout += data.toString();
            });

            child.stderr.on('data', (data) => {
                stderr += data.toString();
            });

            child.on('error', (error) => {
                reject(new Error(`Failed to execute ${binaryType}: ${error.message}`));
            });

            child.on('close', (code) => {
                if (code === 0) {
                    resolve({
                        success: true,
                        stdout: stdout.trim(),
                        stderr: stderr.trim(),
                        exitCode: code
                    });
                } else {
                    reject(new Error(`${binaryType} exited with code ${code}\nStderr: ${stderr}`));
                }
            });
        });
    }

    // PDF Operations
    async pdfExtractText(inputPath) {
        const result = await this.execute('pdf', ['extract-text', inputPath]);
        return result.stdout;
    }

    async pdfMerge(inputPaths, outputPath) {
        const args = ['merge', ...inputPaths, '--output', outputPath];
        return await this.execute('pdf', args);
    }

    async pdfSplit(inputPath, ranges, outputDir) {
        const args = ['split', inputPath, ranges, '--output-dir', outputDir];
        return await this.execute('pdf', args);
    }

    async pdfExtractPages(inputPath, pages, outputPath) {
        const args = ['extract-pages', inputPath, pages, '--output', outputPath];
        return await this.execute('pdf', args);
    }

    async pdfRotate(inputPath, pages, angle, outputPath) {
        const args = ['rotate', inputPath, pages, '--angle', angle.toString(), '--output', outputPath];
        return await this.execute('pdf', args);
    }

    async pdfWatermark(inputPath, text, outputPath) {
        const args = ['watermark', inputPath, '--text', text, '--output', outputPath];
        return await this.execute('pdf', args);
    }

    async pdfCompress(inputPath, outputPath) {
        const args = ['compress', inputPath, '--output', outputPath];
        return await this.execute('pdf', args);
    }

    async pdfEncrypt(inputPath, password, outputPath) {
        const args = ['encrypt', inputPath, '--password', password, '--output', outputPath];
        return await this.execute('pdf', args);
    }

    async pdfInfo(inputPath) {
        const result = await this.execute('pdf', ['info', inputPath, '--json']);
        return JSON.parse(result.stdout);
    }

    async pdfSetInfo(inputPath, info, outputPath) {
        const args = ['set-info', inputPath, '--output', outputPath];
        if (info.title) args.push('--title', info.title);
        if (info.author) args.push('--author', info.author);
        if (info.subject) args.push('--subject', info.subject);
        if (info.keywords) args.push('--keywords', info.keywords);
        return await this.execute('pdf', args);
    }

    async pdfToLIV(inputPath, outputPath, options = {}) {
        const args = ['convert', inputPath];
        if (outputPath) args.push('--output', outputPath);
        if (options.title) args.push('--title', options.title);
        if (options.author) args.push('--author', options.author);
        if (options.compress !== false) args.push('--compress');
        if (options.quality) args.push('--quality', options.quality.toString());
        return await this.execute('converter', args);
    }

    // LIV Document Operations
    async buildDocument(inputDir, outputFile, options = {}) {
        const args = ['build', '--input', inputDir, '--output', outputFile];
        if (options.manifest) args.push('--manifest', options.manifest);
        if (options.sign) args.push('--sign', '--key', options.keyFile);
        if (!options.compress) args.push('--compress=false');
        return await this.execute('cli', args);
    }

    async validateDocument(filePath) {
        const result = await this.execute('cli', ['validate', filePath]);
        return result.stdout;
    }

    async signDocument(inputFile, outputFile, keyFile) {
        return await this.execute('cli', ['sign', '--input', inputFile, '--output', outputFile, '--key', keyFile]);
    }

    async viewDocument(filePath, port = 8080) {
        return await this.execute('cli', ['view', filePath, '--port', port.toString()]);
    }

    async convertDocument(inputFile, outputFile, format) {
        return await this.execute('cli', ['convert', '--input', inputFile, '--output', outputFile, '--format', format]);
    }

    // Integrity Operations
    async checkIntegrity(filePath) {
        return await this.execute('integrity', [filePath]);
    }

    // Manifest Operations
    async validateManifest(manifestPath) {
        return await this.execute('validator', [manifestPath]);
    }
}

module.exports = new GoBackend();
