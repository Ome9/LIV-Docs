/**
 * PDF Operations Module
 * Comprehensive PDF manipulation using pdf-lib, pdfjs-dist, and custom implementations
 * Features from embed-pdf-viewer and Stirling-PDF integrated
 */

const { PDFDocument, rgb, degrees, StandardFonts, PageSizes } = require('pdf-lib');
const fontkit = require('@pdf-lib/fontkit');
const pdfjsLib = require('pdfjs-dist/legacy/build/pdf');
const { jsPDF } = require('jspdf');
const QRCode = require('qrcode');
const JsBarcode = require('jsbarcode');
const Color = require('color');

// Configure PDF.js worker
pdfjsLib.GlobalWorkerOptions.workerSrc = require('pdfjs-dist/legacy/build/pdf.worker.entry');

class PDFOperations {
  constructor() {
    this.loadedPDF = null;
    this.pdfDoc = null;
    this.currentPage = 1;
    this.totalPages = 0;
    this.zoom = 1.0;
    this.rotation = 0;
  }

  /**
   * Load PDF from file path or buffer
   */
  async loadPDF(source) {
    try {
      if (typeof source === 'string') {
        // Load from file path
        const fs = require('fs');
        const arrayBuffer = fs.readFileSync(source).buffer;
        this.pdfDoc = await PDFDocument.load(arrayBuffer);
        this.loadedPDF = await pdfjsLib.getDocument({ data: arrayBuffer }).promise;
      } else {
        // Load from ArrayBuffer
        this.pdfDoc = await PDFDocument.load(source);
        this.loadedPDF = await pdfjsLib.getDocument({ data: source }).promise;
      }
      
      this.totalPages = this.pdfDoc.getPageCount();
      this.currentPage = 1;
      
      return {
        success: true,
        pages: this.totalPages,
        info: await this.getDocumentInfo()
      };
    } catch (error) {
      console.error('Failed to load PDF:', error);
      return { success: false, error: error.message };
    }
  }

  /**
   * Get document metadata
   */
  async getDocumentInfo() {
    if (!this.pdfDoc) return null;
    
    const title = this.pdfDoc.getTitle();
    const author = this.pdfDoc.getAuthor();
    const subject = this.pdfDoc.getSubject();
    const creator = this.pdfDoc.getCreator();
    const producer = this.pdfDoc.getProducer();
    const creationDate = this.pdfDoc.getCreationDate();
    const modificationDate = this.pdfDoc.getModificationDate();
    
    return {
      title,
      author,
      subject,
      creator,
      producer,
      creationDate,
      modificationDate,
      pages: this.totalPages
    };
  }

  /**
   * Update document metadata
   */
  async setDocumentInfo(info) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      if (info.title) this.pdfDoc.setTitle(info.title);
      if (info.author) this.pdfDoc.setAuthor(info.author);
      if (info.subject) this.pdfDoc.setSubject(info.subject);
      if (info.creator) this.pdfDoc.setCreator(info.creator);
      if (info.producer) this.pdfDoc.setProducer(info.producer);
      
      return { success: true };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Merge multiple PDFs into one
   */
  async mergePDFs(pdfPaths) {
    try {
      const mergedPdf = await PDFDocument.create();
      const fs = require('fs');
      
      for (const path of pdfPaths) {
        const pdfBytes = fs.readFileSync(path);
        const pdf = await PDFDocument.load(pdfBytes);
        const copiedPages = await mergedPdf.copyPages(pdf, pdf.getPageIndices());
        copiedPages.forEach((page) => mergedPdf.addPage(page));
      }
      
      const mergedPdfBytes = await mergedPdf.save();
      return { success: true, data: mergedPdfBytes };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Split PDF into multiple files
   */
  async splitPDF(ranges) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      const splitPdfs = [];
      
      for (const range of ranges) {
        const newPdf = await PDFDocument.create();
        const { start, end } = range;
        
        for (let i = start - 1; i < end; i++) {
          const [copiedPage] = await newPdf.copyPages(this.pdfDoc, [i]);
          newPdf.addPage(copiedPage);
        }
        
        const pdfBytes = await newPdf.save();
        splitPdfs.push(pdfBytes);
      }
      
      return { success: true, pdfs: splitPdfs };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Extract specific pages
   */
  async extractPages(pageNumbers) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      const newPdf = await PDFDocument.create();
      const indices = pageNumbers.map(n => n - 1);
      const copiedPages = await newPdf.copyPages(this.pdfDoc, indices);
      copiedPages.forEach((page) => newPdf.addPage(page));
      
      const pdfBytes = await newPdf.save();
      return { success: true, data: pdfBytes };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Delete specific pages
   */
  async deletePages(pageNumbers) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      // Sort in descending order to avoid index issues
      const sortedPages = pageNumbers.sort((a, b) => b - a);
      
      for (const pageNum of sortedPages) {
        if (pageNum > 0 && pageNum <= this.totalPages) {
          this.pdfDoc.removePage(pageNum - 1);
        }
      }
      
      this.totalPages = this.pdfDoc.getPageCount();
      const pdfBytes = await this.pdfDoc.save();
      
      return { success: true, data: pdfBytes, pages: this.totalPages };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Rotate pages
   */
  async rotatePages(pageNumbers, rotation) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      const pages = this.pdfDoc.getPages();
      
      for (const pageNum of pageNumbers) {
        if (pageNum > 0 && pageNum <= pages.length) {
          const page = pages[pageNum - 1];
          const currentRotation = page.getRotation().angle;
          page.setRotation(degrees(currentRotation + rotation));
        }
      }
      
      const pdfBytes = await this.pdfDoc.save();
      return { success: true, data: pdfBytes };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Add watermark to pages
   */
  async addWatermark(options) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      const { text, pages, fontSize = 48, opacity = 0.3, rotation = 45, color = '#888888' } = options;
      const pages_array = this.pdfDoc.getPages();
      const font = await this.pdfDoc.embedFont(StandardFonts.HelveticaBold);
      
      const colorObj = Color(color).rgb().array().map(v => v / 255);
      
      for (const pageNum of pages) {
        if (pageNum > 0 && pageNum <= pages_array.length) {
          const page = pages_array[pageNum - 1];
          const { width, height } = page.getSize();
          const textWidth = font.widthOfTextAtSize(text, fontSize);
          
          page.drawText(text, {
            x: width / 2 - textWidth / 2,
            y: height / 2,
            size: fontSize,
            font: font,
            color: rgb(colorObj[0], colorObj[1], colorObj[2]),
            opacity: opacity,
            rotate: degrees(rotation)
          });
        }
      }
      
      const pdfBytes = await this.pdfDoc.save();
      return { success: true, data: pdfBytes };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Add text to PDF
   */
  async addText(options) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      const { 
        text, 
        page, 
        x, 
        y, 
        fontSize = 12, 
        color = '#000000',
        fontName = StandardFonts.Helvetica,
        bold = false,
        italic = false
      } = options;
      
      const pages = this.pdfDoc.getPages();
      const targetPage = pages[page - 1];
      
      // Select font based on style
      let fontType = fontName;
      if (bold && italic) {
        fontType = StandardFonts.HelveticaBoldOblique;
      } else if (bold) {
        fontType = StandardFonts.HelveticaBold;
      } else if (italic) {
        fontType = StandardFonts.HelveticaOblique;
      }
      
      const font = await this.pdfDoc.embedFont(fontType);
      const colorObj = Color(color).rgb().array().map(v => v / 255);
      
      targetPage.drawText(text, {
        x,
        y,
        size: fontSize,
        font,
        color: rgb(colorObj[0], colorObj[1], colorObj[2])
      });
      
      const pdfBytes = await this.pdfDoc.save();
      return { success: true, data: pdfBytes };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Add image to PDF
   */
  async addImage(options) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      const { imagePath, page, x, y, width, height } = options;
      const fs = require('fs');
      const imageBytes = fs.readFileSync(imagePath);
      
      let image;
      const ext = imagePath.toLowerCase().split('.').pop();
      
      if (ext === 'png') {
        image = await this.pdfDoc.embedPng(imageBytes);
      } else if (ext === 'jpg' || ext === 'jpeg') {
        image = await this.pdfDoc.embedJpg(imageBytes);
      } else {
        return { success: false, error: 'Unsupported image format' };
      }
      
      const pages = this.pdfDoc.getPages();
      const targetPage = pages[page - 1];
      
      targetPage.drawImage(image, { x, y, width, height });
      
      const pdfBytes = await this.pdfDoc.save();
      return { success: true, data: pdfBytes };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Add rectangle/shape to PDF
   */
  async addRectangle(options) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      const { 
        page, 
        x, 
        y, 
        width, 
        height, 
        color = '#000000', 
        borderColor = '#000000',
        borderWidth = 1,
        opacity = 1,
        filled = false
      } = options;
      
      const pages = this.pdfDoc.getPages();
      const targetPage = pages[page - 1];
      
      const fillColorObj = Color(color).rgb().array().map(v => v / 255);
      const borderColorObj = Color(borderColor).rgb().array().map(v => v / 255);
      
      if (filled) {
        targetPage.drawRectangle({
          x, y, width, height,
          color: rgb(fillColorObj[0], fillColorObj[1], fillColorObj[2]),
          opacity,
          borderColor: rgb(borderColorObj[0], borderColorObj[1], borderColorObj[2]),
          borderWidth
        });
      } else {
        targetPage.drawRectangle({
          x, y, width, height,
          borderColor: rgb(borderColorObj[0], borderColorObj[1], borderColorObj[2]),
          borderWidth,
          opacity
        });
      }
      
      const pdfBytes = await this.pdfDoc.save();
      return { success: true, data: pdfBytes };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Add QR Code to PDF
   */
  async addQRCode(options) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      const { data, page, x, y, size = 100 } = options;
      
      // Generate QR code as data URL
      const qrDataUrl = await QRCode.toDataURL(data, { width: size, margin: 1 });
      const base64 = qrDataUrl.split(',')[1];
      const imageBytes = Buffer.from(base64, 'base64');
      
      const image = await this.pdfDoc.embedPng(imageBytes);
      const pages = this.pdfDoc.getPages();
      const targetPage = pages[page - 1];
      
      targetPage.drawImage(image, { x, y, width: size, height: size });
      
      const pdfBytes = await this.pdfDoc.save();
      return { success: true, data: pdfBytes };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Add Barcode to PDF
   */
  async addBarcode(options) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      const { data, page, x, y, width = 200, height = 100, format = 'CODE128' } = options;
      
      // Create canvas for barcode
      const { createCanvas } = require('canvas');
      const canvas = createCanvas(width, height);
      
      JsBarcode(canvas, data, {
        format,
        width: 2,
        height: height * 0.8,
        displayValue: true
      });
      
      const imageBytes = canvas.toBuffer('image/png');
      const image = await this.pdfDoc.embedPng(imageBytes);
      
      const pages = this.pdfDoc.getPages();
      const targetPage = pages[page - 1];
      
      targetPage.drawImage(image, { x, y, width, height });
      
      const pdfBytes = await this.pdfDoc.save();
      return { success: true, data: pdfBytes };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Compress PDF (reduce file size)
   */
  async compressPDF(quality = 0.7) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      // Save with compression options
      const pdfBytes = await this.pdfDoc.save({
        useObjectStreams: true,
        addDefaultPage: false,
        objectsPerTick: 50
      });
      
      return { success: true, data: pdfBytes, originalSize: this.pdfDoc.getBytes().length, newSize: pdfBytes.length };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Convert images to PDF
   */
  async imagesToPDF(imagePaths) {
    try {
      const pdfDoc = await PDFDocument.create();
      const fs = require('fs');
      
      for (const imagePath of imagePaths) {
        const imageBytes = fs.readFileSync(imagePath);
        const ext = imagePath.toLowerCase().split('.').pop();
        
        let image;
        if (ext === 'png') {
          image = await pdfDoc.embedPng(imageBytes);
        } else if (ext === 'jpg' || ext === 'jpeg') {
          image = await pdfDoc.embedJpg(imageBytes);
        } else {
          continue;
        }
        
        const page = pdfDoc.addPage([image.width, image.height]);
        page.drawImage(image, {
          x: 0,
          y: 0,
          width: image.width,
          height: image.height
        });
      }
      
      const pdfBytes = await pdfDoc.save();
      return { success: true, data: pdfBytes };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Add blank page
   */
  async addBlankPage(options = {}) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      const { position = this.totalPages, size = PageSizes.A4 } = options;
      const page = this.pdfDoc.insertPage(position, size);
      
      this.totalPages = this.pdfDoc.getPageCount();
      const pdfBytes = await this.pdfDoc.save();
      
      return { success: true, data: pdfBytes, pages: this.totalPages };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Reorder pages
   */
  async reorderPages(newOrder) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      const newPdf = await PDFDocument.create();
      
      for (const pageNum of newOrder) {
        const [copiedPage] = await newPdf.copyPages(this.pdfDoc, [pageNum - 1]);
        newPdf.addPage(copiedPage);
      }
      
      this.pdfDoc = newPdf;
      this.totalPages = this.pdfDoc.getPageCount();
      
      const pdfBytes = await this.pdfDoc.save();
      return { success: true, data: pdfBytes };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Save PDF to file
   */
  async savePDF(outputPath) {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      const pdfBytes = await this.pdfDoc.save();
      const fs = require('fs');
      fs.writeFileSync(outputPath, pdfBytes);
      
      return { success: true, path: outputPath };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Get PDF as bytes
   */
  async getPDFBytes() {
    if (!this.pdfDoc) return { success: false, error: 'No PDF loaded' };
    
    try {
      const pdfBytes = await this.pdfDoc.save();
      return { success: true, data: pdfBytes };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  /**
   * Create new blank PDF
   */
  async createNewPDF(pageSize = PageSizes.A4) {
    try {
      this.pdfDoc = await PDFDocument.create();
      this.pdfDoc.addPage(pageSize);
      this.totalPages = 1;
      this.currentPage = 1;
      
      return { success: true, pages: this.totalPages };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }
}

module.exports = PDFOperations;
