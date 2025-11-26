import { createClient } from '@libsql/client';

let tursoClient: ReturnType<typeof createClient> | null = null;

// Lazily initialize Turso client to avoid errors during build
// when environment variables aren't available
export function getTurso() {
  if (!tursoClient) {
    const url = process.env.TURSO_DATABASE_URL;
    const token = process.env.TURSO_AUTH_TOKEN;

    if (!url || !token) {
      throw new Error(
        'Turso database credentials not configured. ' +
        'Set TURSO_DATABASE_URL and TURSO_AUTH_TOKEN environment variables.'
      );
    }

    tursoClient = createClient({
      url,
      authToken: token
    });
  }

  return tursoClient;
}
