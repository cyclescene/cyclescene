/**
 * Error logging utility for the PWA
 * Logs errors to the backend monitoring database with client context
 * Features:
 * - Client ID persistence across sessions
 * - Rate limiting (max 10 errors per minute)
 * - Async fire-and-forget logging
 * - Auto-detection of OS, URL, user agent
 */

export type ErrorType = 'db_error' | 'api_error' | 'ui_error' | 'unexpected_error';

export interface ErrorLogContext {
  component?: string;
  action?: string;
  additionalData?: Record<string, any>;
}

class ErrorLogger {
  private clientId: string | null = null;
  private errorCount: number = 0;
  private lastErrorResetTime: number = Date.now();
  private readonly RATE_LIMIT_WINDOW = 60 * 1000; // 1 minute
  private readonly RATE_LIMIT_MAX = 10; // max 10 errors per minute

  constructor() {
    this.initializeClientId();
  }

  /**
   * Get or create client ID from localStorage
   * Falls back to in-memory storage if localStorage is unavailable
   */
  private async initializeClientId(): Promise<void> {
    if (this.clientId) return;

    try {
      // Try to get from localStorage
      if (typeof window !== 'undefined' && window.localStorage) {
        const stored = localStorage.getItem('client-id');
        if (stored) {
          this.clientId = stored;
          return;
        }
      }
    } catch (e) {
      console.warn('localStorage not available, using in-memory client ID', e);
    }

    // Generate new client ID if not found
    const newClientId = this.generateClientId();
    this.clientId = newClientId;

    // Try to persist to localStorage
    try {
      if (typeof window !== 'undefined' && window.localStorage) {
        localStorage.setItem('client-id', newClientId);
      }
    } catch (e) {
      console.warn('Failed to persist client ID to localStorage', e);
    }
  }

  /**
   * Generate a UUID v4 client ID
   */
  private generateClientId(): string {
    if (typeof window !== 'undefined' && window.crypto && window.crypto.randomUUID) {
      return window.crypto.randomUUID();
    }
    // Fallback for older browsers
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
      const r = (Math.random() * 16) | 0;
      const v = c === 'x' ? r : (r & 0x3) | 0x8;
      return v.toString(16);
    });
  }

  /**
   * Check and update rate limit
   */
  private checkRateLimit(): boolean {
    const now = Date.now();

    // Reset counter if window has expired
    if (now - this.lastErrorResetTime > this.RATE_LIMIT_WINDOW) {
      this.errorCount = 0;
      this.lastErrorResetTime = now;
    }

    // Check if we've hit the limit
    if (this.errorCount >= this.RATE_LIMIT_MAX) {
      console.warn(
        `Error logging rate limit exceeded (${this.RATE_LIMIT_MAX} errors per ${this.RATE_LIMIT_WINDOW / 1000}s)`
      );
      return false;
    }

    this.errorCount++;
    return true;
  }

  /**
   * Detect operating system from user agent
   */
  private detectOS(): string {
    if (typeof navigator === 'undefined') return 'Unknown';

    const userAgent = navigator.userAgent.toLowerCase();

    if (userAgent.includes('iphone') || userAgent.includes('ipad')) {
      return 'iOS';
    } else if (userAgent.includes('android')) {
      return 'Android';
    } else if (userAgent.includes('win')) {
      return 'Windows';
    } else if (userAgent.includes('mac')) {
      return 'macOS';
    } else if (userAgent.includes('linux')) {
      return 'Linux';
    } else if (userAgent.includes('x11')) {
      return 'Unix';
    }

    return 'Unknown';
  }

  /**
   * Log an error to the backend
   */
  async logError(
    errorType: ErrorType,
    error: Error | string,
    context?: ErrorLogContext
  ): Promise<void> {
    // Check rate limit
    if (!this.checkRateLimit()) {
      return;
    }

    // Ensure client ID is initialized
    if (!this.clientId) {
      await this.initializeClientId();
    }

    if (!this.clientId) {
      console.error('Failed to initialize client ID');
      return;
    }

    // Extract error details
    const errorMsg = error instanceof Error ? error.message : String(error);
    const stackTrace = error instanceof Error ? error.stack : undefined;

    // Get current URL
    const url = typeof window !== 'undefined' ? window.location.href : '';

    // Get user agent
    const userAgent = typeof navigator !== 'undefined' ? navigator.userAgent : '';

    // Build request payload
    const payload = {
      client_id: this.clientId,
      error_type: errorType,
      error_msg: errorMsg,
      stack_trace: stackTrace,
      component: context?.component,
      action: context?.action,
      url,
      user_agent: userAgent,
      os: this.detectOS(),
      timestamp: new Date().toISOString()
    };

    // Log async (fire-and-forget)
    this.sendErrorToBackend(payload).catch((err) => {
      console.error('Failed to send error log to backend:', err);
    });
  }

  /**
   * Send error to backend endpoint
   */
  private async sendErrorToBackend(payload: any): Promise<void> {
    try {
      const response = await fetch('https://api.cyclescene.cc/v1/client-errors', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      if (!response.ok) {
        console.warn(`Error logging failed with status ${response.status}`);
      }
    } catch (error) {
      // Silently fail - don't block the application
      console.error('Failed to log error to backend:', error);
    }
  }
}

// Export singleton instance
export const errorLogger = new ErrorLogger();
