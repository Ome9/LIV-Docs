// API client for backend operations
export class ApiClient {
  private baseUrl: string

  constructor(baseUrl: string = '') {
    this.baseUrl = baseUrl
  }

  private async request(endpoint: string, options: RequestInit = {}): Promise<ApiResponse> {
    const url = `${this.baseUrl}/api${endpoint}`
    const response = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    })

    const data = await response.json()
    
    if (!response.ok) {
      throw new ApiError(data.error || 'API request failed')
    }

    return data
  }

  // PDF Operations
  async mergePdfs(files: string[], outputPath: string) {
    return this.request('/pdf/merge', {
      method: 'POST',
      body: JSON.stringify({ files, outputPath }),
    })
  }

  async splitPdf(inputPath: string, ranges: string, outputDir: string) {
    return this.request('/pdf/split', {
      method: 'POST',
      body: JSON.stringify({ inputPath, ranges, outputDir }),
    })
  }

  async compressPdf(inputPath: string, outputPath: string, quality?: number) {
    return this.request('/pdf/compress', {
      method: 'POST',
      body: JSON.stringify({ inputPath, outputPath, quality }),
    })
  }

  async extractPages(inputPath: string, pages: string, outputDir: string) {
    return this.request('/pdf/extract', {
      method: 'POST',
      body: JSON.stringify({ inputPath, pages, outputDir }),
    })
  }

  async addWatermark(
    inputPath: string,
    outputPath: string,
    options: {
      watermarkText?: string
      watermarkImage?: string
      opacity?: number
      position?: string
    }
  ) {
    return this.request('/pdf/watermark', {
      method: 'POST',
      body: JSON.stringify({ inputPath, outputPath, ...options }),
    })
  }

  async rotatePdf(inputPath: string, outputPath: string, angle: number, pages?: string) {
    return this.request('/pdf/rotate', {
      method: 'POST',
      body: JSON.stringify({ inputPath, outputPath, angle, pages }),
    })
  }

  async encryptPdf(inputPath: string, outputPath: string, password: string, permissions?: string) {
    return this.request('/pdf/encrypt', {
      method: 'POST',
      body: JSON.stringify({ inputPath, outputPath, password, permissions }),
    })
  }

  async getPdfInfo(inputPath: string) {
    return this.request('/pdf/info', {
      method: 'POST',
      body: JSON.stringify({ inputPath }),
    })
  }

  // Conversion Operations
  async convertFile(inputPath: string, outputPath: string, format: string) {
    return this.request('/converter', {
      method: 'POST',
      body: JSON.stringify({ inputPath, outputPath, format }),
    })
  }

  // LIV Document Operations
  async buildLivDocument(inputPath: string, outputPath: string, manifest?: string) {
    return this.request('/builder', {
      method: 'POST',
      body: JSON.stringify({ inputPath, outputPath, manifest }),
    })
  }

  async packLivDocument(inputPath: string, outputPath: string, key?: string) {
    return this.request('/pack', {
      method: 'POST',
      body: JSON.stringify({ inputPath, outputPath, key }),
    })
  }
}

// Create a singleton instance
export const apiClient = new ApiClient()

// Hook for React components
export function useApiClient() {
  return apiClient
}

// Error handling utility
export class ApiError extends Error {
  constructor(message: string, public code?: string) {
    super(message)
    this.name = 'ApiError'
  }
}

// Type definitions for API responses
export interface ApiResponse<T = any> {
  success: boolean
  message?: string
  error?: string
  data?: T
}

export interface PdfInfo {
  pages: number
  title?: string
  author?: string
  creator?: string
  producer?: string
  creationDate?: string
  modificationDate?: string
  encrypted: boolean
  pageSize?: {
    width: number
    height: number
  }
}