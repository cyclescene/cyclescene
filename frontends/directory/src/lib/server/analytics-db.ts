import { createClient } from '@libsql/client';

let analyticsClient: ReturnType<typeof createClient> | null = null;

// Lazily initialize analytics database client
// This is a separate Turso instance dedicated to marketing analytics
// Environment variables: ANALYTICS_DB_URL and ANALYTICS_DB_TOKEN
export function getAnalyticsDb() {
  if (!analyticsClient) {
    const url = process.env.ANALYTICS_DB_URL;
    const token = process.env.ANALYTICS_DB_TOKEN;

    if (!url || !token) {
      throw new Error(
        'Analytics database credentials not configured. ' +
        'Set ANALYTICS_DB_URL and ANALYTICS_DB_TOKEN environment variables.'
      );
    }

    analyticsClient = createClient({
      url,
      authToken: token
    });
  }

  return analyticsClient;
}
