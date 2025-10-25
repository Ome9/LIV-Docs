// Error handling classes for LIV document system

export enum LIVErrorType {
  INVALID_FILE = 'invalid_file',
  CORRUPTED = 'corrupted',
  UNSUPPORTED = 'unsupported',
  SECURITY = 'security',
  TIMEOUT = 'timeout',
  RESOURCE_LIMIT = 'resource_limit',
  VALIDATION = 'validation',
  PARSING = 'parsing',
  NETWORK = 'network',
  PERMISSION_DENIED = 'permission_denied'
}

export class LIVError extends Error {
  public readonly type: LIVErrorType;
  public readonly cause?: Error;
  public readonly details?: Record<string, any>;

  constructor(type: LIVErrorType, message: string, cause?: Error, details?: Record<string, any>) {
    super(message);
    this.name = 'LIVError';
    this.type = type;
    this.cause = cause;
    this.details = details;

    // Maintain proper stack trace
    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, LIVError);
    }
  }

  toString(): string {
    let result = `${this.name} [${this.type}]: ${this.message}`;
    
    if (this.cause) {
      result += `\nCaused by: ${this.cause.message}`;
    }
    
    if (this.details) {
      result += `\nDetails: ${JSON.stringify(this.details, null, 2)}`;
    }
    
    return result;
  }

  toJSON(): Record<string, any> {
    return {
      name: this.name,
      type: this.type,
      message: this.message,
      cause: this.cause?.message,
      details: this.details,
      stack: this.stack
    };
  }
}

// Specific error classes for different scenarios

export class InvalidFileError extends LIVError {
  constructor(message: string, filename?: string) {
    super(LIVErrorType.INVALID_FILE, message, undefined, { filename });
    this.name = 'InvalidFileError';
  }
}

export class CorruptedFileError extends LIVError {
  constructor(message: string, cause?: Error) {
    super(LIVErrorType.CORRUPTED, message, cause);
    this.name = 'CorruptedFileError';
  }
}

export class UnsupportedFeatureError extends LIVError {
  constructor(feature: string, version?: string) {
    super(
      LIVErrorType.UNSUPPORTED, 
      `Unsupported feature: ${feature}`,
      undefined,
      { feature, version }
    );
    this.name = 'UnsupportedFeatureError';
  }
}

export class SecurityError extends LIVError {
  constructor(message: string, violationType?: string) {
    super(LIVErrorType.SECURITY, message, undefined, { violationType });
    this.name = 'SecurityError';
  }
}

export class ValidationError extends LIVError {
  public readonly errors: string[];
  public readonly warnings: string[];

  constructor(message: string, errors: string[] = [], warnings: string[] = []) {
    super(LIVErrorType.VALIDATION, message, undefined, { errors, warnings });
    this.name = 'ValidationError';
    this.errors = errors;
    this.warnings = warnings;
  }
}

export class ResourceLimitError extends LIVError {
  constructor(resource: string, limit: number, actual: number) {
    super(
      LIVErrorType.RESOURCE_LIMIT,
      `Resource '${resource}' exceeds limit: ${actual} > ${limit}`,
      undefined,
      { resource, limit, actual }
    );
    this.name = 'ResourceLimitError';
  }
}

export class TimeoutError extends LIVError {
  constructor(operation: string, timeout: number) {
    super(
      LIVErrorType.TIMEOUT,
      `Operation '${operation}' timed out after ${timeout}ms`,
      undefined,
      { operation, timeout }
    );
    this.name = 'TimeoutError';
  }
}

export class ParsingError extends LIVError {
  constructor(message: string, context?: string, cause?: Error) {
    super(LIVErrorType.PARSING, message, cause, { context });
    this.name = 'ParsingError';
  }
}

export class NetworkError extends LIVError {
  constructor(message: string, url?: string, status?: number) {
    super(LIVErrorType.NETWORK, message, undefined, { url, status });
    this.name = 'NetworkError';
  }
}

export class PermissionDeniedError extends LIVError {
  constructor(permission: string, resource?: string) {
    super(
      LIVErrorType.PERMISSION_DENIED,
      `Permission denied: ${permission}`,
      undefined,
      { permission, resource }
    );
    this.name = 'PermissionDeniedError';
  }
}

// Error handler utility class
export class ErrorHandler {
  private static instance: ErrorHandler;
  private errorListeners: Array<(error: LIVError) => void> = [];
  private errorHistory: LIVError[] = [];
  private maxHistorySize: number = 100;

  private constructor() {}

  static getInstance(): ErrorHandler {
    if (!ErrorHandler.instance) {
      ErrorHandler.instance = new ErrorHandler();
    }
    return ErrorHandler.instance;
  }

  addErrorListener(listener: (error: LIVError) => void): void {
    this.errorListeners.push(listener);
  }

  removeErrorListener(listener: (error: LIVError) => void): void {
    const index = this.errorListeners.indexOf(listener);
    if (index > -1) {
      this.errorListeners.splice(index, 1);
    }
  }

  handleError(error: Error | LIVError): void {
    let livError: LIVError;

    if (error instanceof LIVError) {
      livError = error;
    } else {
      // Convert generic error to LIVError
      livError = new LIVError(LIVErrorType.VALIDATION, error.message, error);
    }

    // Add to history
    this.errorHistory.push(livError);
    if (this.errorHistory.length > this.maxHistorySize) {
      this.errorHistory.shift();
    }

    // Notify listeners
    this.errorListeners.forEach(listener => {
      try {
        listener(livError);
      } catch (listenerError) {
        console.error('Error in error listener:', listenerError);
      }
    });

    // Log error
    console.error('LIV Error:', livError.toString());
  }

  getErrorHistory(): LIVError[] {
    return [...this.errorHistory];
  }

  clearErrorHistory(): void {
    this.errorHistory = [];
  }

  createRecoveryStrategy(error: LIVError): RecoveryStrategy | null {
    switch (error.type) {
      case LIVErrorType.INVALID_FILE:
        return new RecoveryStrategy(
          'Try loading a different file or check file format',
          () => Promise.resolve(false)
        );

      case LIVErrorType.CORRUPTED:
        return new RecoveryStrategy(
          'File may be corrupted. Try re-downloading or using a backup',
          () => Promise.resolve(false)
        );

      case LIVErrorType.SECURITY:
        return new RecoveryStrategy(
          'Security validation failed. Enable fallback mode or update security settings',
          async () => {
            // Could implement fallback loading here
            return false;
          }
        );

      case LIVErrorType.RESOURCE_LIMIT:
        return new RecoveryStrategy(
          'Resource exceeds limits. Try increasing limits or using a smaller file',
          () => Promise.resolve(false)
        );

      case LIVErrorType.TIMEOUT:
        return new RecoveryStrategy(
          'Operation timed out. Try increasing timeout or check network connection',
          () => Promise.resolve(false)
        );

      default:
        return null;
    }
  }
}

export class RecoveryStrategy {
  constructor(
    public readonly description: string,
    public readonly execute: () => Promise<boolean>
  ) {}
}

// Utility functions for error handling
export function isLIVError(error: any): error is LIVError {
  return error instanceof LIVError;
}

export function createErrorFromValidation(validation: { isValid: boolean; errors: string[]; warnings: string[] }): ValidationError | null {
  if (validation.isValid) {
    return null;
  }

  return new ValidationError(
    `Validation failed with ${validation.errors.length} errors`,
    validation.errors,
    validation.warnings
  );
}

export function wrapAsyncOperation<T>(
  operation: () => Promise<T>,
  errorType: LIVErrorType,
  context?: string
): Promise<T> {
  return operation().catch(error => {
    if (error instanceof LIVError) {
      throw error;
    }
    throw new LIVError(errorType, error.message, error, { context });
  });
}

export function withTimeout<T>(
  promise: Promise<T>,
  timeoutMs: number,
  operation: string
): Promise<T> {
  return Promise.race([
    promise,
    new Promise<never>((_, reject) => {
      setTimeout(() => {
        reject(new TimeoutError(operation, timeoutMs));
      }, timeoutMs);
    })
  ]);
}

// Global error handler setup
export function setupGlobalErrorHandling(): void {
  const errorHandler = ErrorHandler.getInstance();

  // Handle unhandled promise rejections
  window.addEventListener('unhandledrejection', (event) => {
    if (isLIVError(event.reason)) {
      errorHandler.handleError(event.reason);
      event.preventDefault();
    }
  });

  // Handle general errors
  window.addEventListener('error', (event) => {
    const error = new LIVError(
      LIVErrorType.VALIDATION,
      event.message,
      event.error,
      {
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno
      }
    );
    errorHandler.handleError(error);
  });
}