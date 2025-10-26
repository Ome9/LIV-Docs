'use client'

import React, { useState } from 'react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { 
  FileText, 
  Split, 
  Merge, 
  Minimize2, 
  RotateCw, 
  Lock, 
  Info,
  Download,
  FileImage,
  Scissors
} from 'lucide-react'
import { apiClient } from '@/lib/api-client'
import { fileDialog } from '@/lib/file-dialog'

export function PdfToolsPanel() {
  const [loading, setLoading] = useState<string | null>(null)
  const [results, setResults] = useState<any>(null)

  const handleApiCall = async (operation: string, apiCall: () => Promise<any>) => {
    setLoading(operation)
    setResults(null)
    try {
      const result = await apiCall()
      setResults(result)
    } catch (error) {
      setResults({ success: false, error: error instanceof Error ? error.message : 'Unknown error' })
    } finally {
      setLoading(null)
    }
  }

  const mergePdfs = async () => {
    if (!fileDialog.isDesktopApp) {
      setResults({ success: false, error: 'This feature requires the desktop app' })
      return
    }

    const files = await fileDialog.openFiles({
      title: 'Select PDF files to merge',
      filters: [{ name: 'PDF Files', extensions: ['pdf'] }]
    })

    if (!files || files.length < 2) {
      setResults({ success: false, error: 'Please select at least 2 PDF files' })
      return
    }

    const outputPath = await fileDialog.saveFile({
      title: 'Save merged PDF',
      defaultPath: 'merged.pdf',
      filters: [{ name: 'PDF Files', extensions: ['pdf'] }]
    })

    if (!outputPath) return

    await handleApiCall('merge', () => apiClient.mergePdfs(files, outputPath))
  }

  const splitPdf = async () => {
    if (!fileDialog.isDesktopApp) {
      setResults({ success: false, error: 'This feature requires the desktop app' })
      return
    }

    const inputPath = await fileDialog.openFile({
      title: 'Select PDF to split',
      filters: [{ name: 'PDF Files', extensions: ['pdf'] }]
    })

    if (!inputPath) return

    const outputDir = await fileDialog.selectDirectory({
      title: 'Select output directory'
    })

    if (!outputDir) return

    // For demo, using page ranges "1-5,10-15"
    const ranges = prompt('Enter page ranges (e.g., "1-5,10-15"):')
    if (!ranges) return

    await handleApiCall('split', () => apiClient.splitPdf(inputPath, ranges, outputDir))
  }

  const compressPdf = async () => {
    if (!fileDialog.isDesktopApp) {
      setResults({ success: false, error: 'This feature requires the desktop app' })
      return
    }

    const inputPath = await fileDialog.openFile({
      title: 'Select PDF to compress',
      filters: [{ name: 'PDF Files', extensions: ['pdf'] }]
    })

    if (!inputPath) return

    const outputPath = await fileDialog.saveFile({
      title: 'Save compressed PDF',
      defaultPath: 'compressed.pdf',
      filters: [{ name: 'PDF Files', extensions: ['pdf'] }]
    })

    if (!outputPath) return

    await handleApiCall('compress', () => apiClient.compressPdf(inputPath, outputPath, 75))
  }

  const extractPages = async () => {
    if (!fileDialog.isDesktopApp) {
      setResults({ success: false, error: 'This feature requires the desktop app' })
      return
    }

    const inputPath = await fileDialog.openFile({
      title: 'Select PDF to extract pages from',
      filters: [{ name: 'PDF Files', extensions: ['pdf'] }]
    })

    if (!inputPath) return

    const outputDir = await fileDialog.selectDirectory({
      title: 'Select output directory'
    })

    if (!outputDir) return

    const pages = prompt('Enter pages to extract (e.g., "1,3,5-7"):')
    if (!pages) return

    await handleApiCall('extract', () => apiClient.extractPages(inputPath, pages, outputDir))
  }

  const rotatePdf = async () => {
    if (!fileDialog.isDesktopApp) {
      setResults({ success: false, error: 'This feature requires the desktop app' })
      return
    }

    const inputPath = await fileDialog.openFile({
      title: 'Select PDF to rotate',
      filters: [{ name: 'PDF Files', extensions: ['pdf'] }]
    })

    if (!inputPath) return

    const outputPath = await fileDialog.saveFile({
      title: 'Save rotated PDF',
      defaultPath: 'rotated.pdf',
      filters: [{ name: 'PDF Files', extensions: ['pdf'] }]
    })

    if (!outputPath) return

    const angle = prompt('Enter rotation angle (90, 180, 270):')
    if (!angle) return

    await handleApiCall('rotate', () => apiClient.rotatePdf(inputPath, outputPath, parseInt(angle)))
  }

  const addWatermark = async () => {
    if (!fileDialog.isDesktopApp) {
      setResults({ success: false, error: 'This feature requires the desktop app' })
      return
    }

    const inputPath = await fileDialog.openFile({
      title: 'Select PDF to watermark',
      filters: [{ name: 'PDF Files', extensions: ['pdf'] }]
    })

    if (!inputPath) return

    const outputPath = await fileDialog.saveFile({
      title: 'Save watermarked PDF',
      defaultPath: 'watermarked.pdf',
      filters: [{ name: 'PDF Files', extensions: ['pdf'] }]
    })

    if (!outputPath) return

    const watermarkText = prompt('Enter watermark text:')
    if (!watermarkText) return

    await handleApiCall('watermark', () => 
      apiClient.addWatermark(inputPath, outputPath, { 
        watermarkText, 
        opacity: 0.5,
        position: 'center'
      })
    )
  }

  const encryptPdf = async () => {
    if (!fileDialog.isDesktopApp) {
      setResults({ success: false, error: 'This feature requires the desktop app' })
      return
    }

    const inputPath = await fileDialog.openFile({
      title: 'Select PDF to encrypt',
      filters: [{ name: 'PDF Files', extensions: ['pdf'] }]
    })

    if (!inputPath) return

    const outputPath = await fileDialog.saveFile({
      title: 'Save encrypted PDF',
      defaultPath: 'encrypted.pdf',
      filters: [{ name: 'PDF Files', extensions: ['pdf'] }]
    })

    if (!outputPath) return

    const password = prompt('Enter password for encryption:')
    if (!password) return

    await handleApiCall('encrypt', () => apiClient.encryptPdf(inputPath, outputPath, password))
  }

  const getPdfInfo = async () => {
    if (!fileDialog.isDesktopApp) {
      setResults({ success: false, error: 'This feature requires the desktop app' })
      return
    }

    const inputPath = await fileDialog.openFile({
      title: 'Select PDF to analyze',
      filters: [{ name: 'PDF Files', extensions: ['pdf'] }]
    })

    if (!inputPath) return

    await handleApiCall('info', () => apiClient.getPdfInfo(inputPath))
  }

  const convertFile = async () => {
    if (!fileDialog.isDesktopApp) {
      setResults({ success: false, error: 'This feature requires the desktop app' })
      return
    }

    const inputPath = await fileDialog.openFile({
      title: 'Select file to convert',
      filters: [
        { name: 'PDF Files', extensions: ['pdf'] },
        { name: 'HTML Files', extensions: ['html', 'htm'] },
        { name: 'Markdown Files', extensions: ['md'] },
        { name: 'All Files', extensions: ['*'] }
      ]
    })

    if (!inputPath) return

    const outputPath = await fileDialog.saveFile({
      title: 'Save converted file',
      defaultPath: 'converted.liv',
      filters: [{ name: 'LIV Files', extensions: ['liv'] }]
    })

    if (!outputPath) return

    const format = prompt('Enter target format (liv, pdf, html):')
    if (!format) return

    await handleApiCall('convert', () => apiClient.convertFile(inputPath, outputPath, format))
  }

  return (
    <div className="w-full h-full">
      <div className="p-3 border-b border-border">
        <h3 className="font-semibold text-sm flex items-center gap-2">
          <FileText className="h-4 w-4" />
          Tools
        </h3>
      </div>
      <div className="flex-1 overflow-hidden">
        <Tabs defaultValue="pdf-operations" className="flex-1 flex flex-col">
          <TabsList className="grid w-full grid-cols-3 shrink-0">
            <TabsTrigger value="pdf-operations">Operations</TabsTrigger>
            <TabsTrigger value="conversion">Convert</TabsTrigger>
            <TabsTrigger value="analysis">Analyze</TabsTrigger>
          </TabsList>

            <TabsContent value="pdf-operations" className="space-y-2 p-4">
              <div className="space-y-2">
                <Button 
                  onClick={mergePdfs} 
                  disabled={loading === 'merge'}
                  variant="outline"
                  className="w-full justify-start gap-3 h-12 text-left"
                >
                  <Merge className="h-5 w-5 text-blue-600" />
                  <div className="flex flex-col items-start">
                    <span className="font-medium">{loading === 'merge' ? 'Merging PDFs...' : 'Merge PDFs'}</span>
                    <span className="text-xs text-muted-foreground">Combine multiple PDF files</span>
                  </div>
                </Button>

                <Button 
                  onClick={splitPdf} 
                  disabled={loading === 'split'}
                  variant="outline"
                  className="w-full justify-start gap-3 h-12 text-left"
                >
                  <Split className="h-5 w-5 text-green-600" />
                  <div className="flex flex-col items-start">
                    <span className="font-medium">{loading === 'split' ? 'Splitting PDF...' : 'Split PDF'}</span>
                    <span className="text-xs text-muted-foreground">Split PDF by page ranges</span>
                  </div>
                </Button>

                <Button 
                  onClick={compressPdf} 
                  disabled={loading === 'compress'}
                  variant="outline"
                  className="w-full justify-start gap-3 h-12 text-left"
                >
                  <Minimize2 className="h-5 w-5 text-purple-600" />
                  <div className="flex flex-col items-start">
                    <span className="font-medium">{loading === 'compress' ? 'Compressing PDF...' : 'Compress PDF'}</span>
                    <span className="text-xs text-muted-foreground">Reduce file size</span>
                  </div>
                </Button>

                <Button 
                  onClick={extractPages} 
                  disabled={loading === 'extract'}
                  variant="outline"
                  className="w-full justify-start gap-3 h-12 text-left"
                >
                  <Scissors className="h-5 w-5 text-orange-600" />
                  <div className="flex flex-col items-start">
                    <span className="font-medium">{loading === 'extract' ? 'Extracting Pages...' : 'Extract Pages'}</span>
                    <span className="text-xs text-muted-foreground">Extract specific pages</span>
                  </div>
                </Button>

                <Button 
                  onClick={rotatePdf} 
                  disabled={loading === 'rotate'}
                  variant="outline"
                  className="w-full justify-start gap-3 h-12 text-left"
                >
                  <RotateCw className="h-5 w-5 text-teal-600" />
                  <div className="flex flex-col items-start">
                    <span className="font-medium">{loading === 'rotate' ? 'Rotating PDF...' : 'Rotate PDF'}</span>
                    <span className="text-xs text-muted-foreground">Rotate pages by angle</span>
                  </div>
                </Button>

                <Button 
                  onClick={addWatermark} 
                  disabled={loading === 'watermark'}
                  variant="outline"
                  className="w-full justify-start gap-3 h-12 text-left"
                >
                  <FileImage className="h-5 w-5 text-indigo-600" />
                  <div className="flex flex-col items-start">
                    <span className="font-medium">{loading === 'watermark' ? 'Adding Watermark...' : 'Add Watermark'}</span>
                    <span className="text-xs text-muted-foreground">Add text or image overlay</span>
                  </div>
                </Button>

                <Button 
                  onClick={encryptPdf} 
                  disabled={loading === 'encrypt'}
                  variant="outline"
                  className="w-full justify-start gap-3 h-12 text-left"
                >
                  <Lock className="h-5 w-5 text-red-600" />
                  <div className="flex flex-col items-start">
                    <span className="font-medium">{loading === 'encrypt' ? 'Encrypting PDF...' : 'Encrypt PDF'}</span>
                    <span className="text-xs text-muted-foreground">Password protect document</span>
                  </div>
                </Button>
              </div>
            </TabsContent>

            <TabsContent value="conversion" className="space-y-2 p-4">
              <div className="space-y-2">
                <Button 
                  onClick={convertFile} 
                  disabled={loading === 'convert'}
                  variant="outline"
                  className="w-full justify-start gap-3 h-12 text-left"
                >
                  <Download className="h-5 w-5 text-blue-600" />
                  <div className="flex flex-col items-start">
                    <span className="font-medium">{loading === 'convert' ? 'Converting Document...' : 'Convert Document'}</span>
                    <span className="text-xs text-muted-foreground">Convert files to LIV format</span>
                  </div>
                </Button>
              </div>
            </TabsContent>

            <TabsContent value="analysis" className="space-y-2 p-4">
              <div className="space-y-2">
                <Button 
                  onClick={getPdfInfo} 
                  disabled={loading === 'info'}
                  variant="outline"
                  className="w-full justify-start gap-3 h-12 text-left"
                >
                  <Info className="h-5 w-5 text-cyan-600" />
                  <div className="flex flex-col items-start">
                    <span className="font-medium">{loading === 'info' ? 'Analyzing PDF...' : 'Get PDF Info'}</span>
                    <span className="text-xs text-muted-foreground">View document metadata</span>
                  </div>
                </Button>
              </div>
            </TabsContent>
          {/* Results Display */}
          {results && (
            <div className="p-3 mx-4 mb-4 border rounded-lg bg-card">
              <h4 className={`font-medium mb-2 text-sm ${results.success ? 'text-green-600' : 'text-red-600'}`}>
                {results.success ? 'Success' : 'Error'}
              </h4>
              <div>
                {results.success ? (
                  <div className="space-y-2">
                    <p className="text-green-700 text-sm">{results.message}</p>
                    {results.info && (
                      <pre className="bg-muted p-2 rounded text-xs overflow-auto max-h-24">
                        {typeof results.info === 'object' 
                          ? JSON.stringify(results.info, null, 2)
                          : results.info
                        }
                      </pre>
                    )}
                  </div>
                ) : (
                  <p className="text-red-700 text-sm">{results.error}</p>
                )}
              </div>
            </div>
          )}

          {!fileDialog.isDesktopApp && (
            <div className="p-3 mx-4 mb-4 bg-amber-50 border border-amber-200 rounded-lg">
              <p className="text-amber-800 text-xs">
                <strong>Note:</strong> PDF processing requires the desktop app.
              </p>
            </div>
          )}
        </Tabs>
      </div>
    </div>
  )
}