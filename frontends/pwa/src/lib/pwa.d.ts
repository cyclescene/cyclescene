// Declare the global interface for the Service Worker Registration
declare global {
  /**
   * Defines the interface for the PeriodicSyncManager.
   * Note: This is not officially in the standard TS lib yet.
   */
  interface PeriodicSyncManager {
    register(tag: string, options?: { minInterval: number }): Promise<void>;
    unregister(tag: string): Promise<void>;
    getTags(): Promise<string[]>;
  }

  /**
   * Defines the interface for the SyncManager (Background Sync API).
   * Used on Apple devices and browsers that don't support Periodic Sync.
   */
  interface SyncManager {
    register(tag: string): Promise<void>;
    getTags(): Promise<string[]>;
  }

  /**
   * Extends the standard ServiceWorkerRegistration to include sync managers.
   */
  interface ServiceWorkerRegistration {
    readonly periodicSync: PeriodicSyncManager;
    readonly sync: SyncManager;
  }

  // Also add the custom event type for the Service Worker file
  interface PeriodicSyncEvent extends ExtendableEvent {
    readonly tag: string;
    readonly lastChance: boolean;
  }

  /**
   * Background Sync event for the sync listener.
   */
  interface SyncEvent extends ExtendableEvent {
    readonly tag: string;
    readonly lastChance: boolean;
  }
}

// Export a dummy object to ensure the file is treated as a module
export { };
