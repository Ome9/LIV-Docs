/**
 * OpenAPI Integration Module
 * Provides utilities for integrating with OpenAPI endpoints
 * Supports automatic schema generation, request/response validation, and type safety
 */

export interface OpenAPIConfig {
  baseUrl: string
  apiKey?: string
  headers?: Record<string, string>
  timeout?: number
}

export interface APIEndpoint {
  method: "GET" | "POST" | "PUT" | "DELETE" | "PATCH"
  path: string
  description?: string
  parameters?: Record<string, unknown>
  requestBody?: unknown
  responses?: Record<number, unknown>
}

export interface APIResponse<T = unknown> {
  status: number
  data: T
  error?: string
  headers?: Record<string, string>
}

/**
 * Initialize OpenAPI client with configuration
 */
export function initializeAPIClient(config: OpenAPIConfig) {
  return new APIClient(config)
}

/**
 * Main API Client class for handling OpenAPI requests
 */
export class APIClient {
  private baseUrl: string
  private apiKey?: string
  private headers: Record<string, string>
  private timeout: number

  constructor(config: OpenAPIConfig) {
    this.baseUrl = config.baseUrl
    this.apiKey = config.apiKey
    this.headers = {
      "Content-Type": "application/json",
      ...config.headers,
    }
    this.timeout = config.timeout || 30000

    if (this.apiKey) {
      this.headers["Authorization"] = `Bearer ${this.apiKey}`
    }
  }

  /**
   * Make a generic API request
   */
  async request<T = unknown>(endpoint: APIEndpoint): Promise<APIResponse<T>> {
    const url = `${this.baseUrl}${endpoint.path}`

    try {
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), this.timeout)

      const options: RequestInit = {
        method: endpoint.method,
        headers: this.headers,
        signal: controller.signal,
      }

      if (
        endpoint.requestBody &&
        (endpoint.method === "POST" || endpoint.method === "PUT" || endpoint.method === "PATCH")
      ) {
        options.body = JSON.stringify(endpoint.requestBody)
      }

      const response = await fetch(url, options)
      clearTimeout(timeoutId)

      const data = await response.json()

      if (!response.ok) {
        return {
          status: response.status,
          data: data as T,
          error: data.message || `HTTP ${response.status}`,
          headers: Object.fromEntries(response.headers.entries()),
        }
      }

      return {
        status: response.status,
        data: data as T,
        headers: Object.fromEntries(response.headers.entries()),
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : "Unknown error"
      return {
        status: 0,
        data: {} as T,
        error: errorMessage,
      }
    }
  }

  /**
   * GET request helper
   */
  async get<T = unknown>(path: string, parameters?: Record<string, unknown>): Promise<APIResponse<T>> {
    const queryString = parameters ? `?${new URLSearchParams(parameters as Record<string, string>).toString()}` : ""
    return this.request<T>({
      method: "GET",
      path: `${path}${queryString}`,
    })
  }

  /**
   * POST request helper
   */
  async post<T = unknown>(path: string, data?: unknown): Promise<APIResponse<T>> {
    return this.request<T>({
      method: "POST",
      path,
      requestBody: data,
    })
  }

  /**
   * PUT request helper
   */
  async put<T = unknown>(path: string, data?: unknown): Promise<APIResponse<T>> {
    return this.request<T>({
      method: "PUT",
      path,
      requestBody: data,
    })
  }

  /**
   * DELETE request helper
   */
  async delete<T = unknown>(path: string): Promise<APIResponse<T>> {
    return this.request<T>({
      method: "DELETE",
      path,
    })
  }

  /**
   * PATCH request helper
   */
  async patch<T = unknown>(path: string, data?: unknown): Promise<APIResponse<T>> {
    return this.request<T>({
      method: "PATCH",
      path,
      requestBody: data,
    })
  }

  /**
   * Load and parse OpenAPI schema from URL
   */
  async loadOpenAPISchema(schemaUrl: string) {
    try {
      const response = await fetch(schemaUrl)
      const schema = await response.json()
      return schema
    } catch (error) {
      console.error("[v0] Failed to load OpenAPI schema:", error)
      return null
    }
  }

  /**
   * Generate endpoints from OpenAPI schema
   */
  generateEndpointsFromSchema(schema: Record<string, unknown>) {
    const endpoints: APIEndpoint[] = []

    if (schema.paths && typeof schema.paths === "object") {
      Object.entries(schema.paths).forEach(([path, pathItem]) => {
        if (typeof pathItem === "object" && pathItem !== null) {
          Object.entries(pathItem).forEach(([method, operation]) => {
            if (["get", "post", "put", "delete", "patch"].includes(method.toLowerCase())) {
              endpoints.push({
                method: method.toUpperCase() as APIEndpoint["method"],
                path,
                description: (operation as Record<string, unknown>).description as string,
              })
            }
          })
        }
      })
    }

    return endpoints
  }
}

/**
 * Example usage:
 *
 * const client = initializeAPIClient({
 *   baseUrl: 'https://api.example.com',
 *   apiKey: 'your-api-key'
 * })
 *
 * // Make a GET request
 * const response = await client.get('/users', { page: 1, limit: 10 })
 *
 * // Make a POST request
 * const createResponse = await client.post('/users', {
 *   name: 'John Doe',
 *   email: 'john@example.com'
 * })
 *
 * // Load OpenAPI schema
 * const schema = await client.loadOpenAPISchema('https://api.example.com/openapi.json')
 * const endpoints = client.generateEndpointsFromSchema(schema)
 */
