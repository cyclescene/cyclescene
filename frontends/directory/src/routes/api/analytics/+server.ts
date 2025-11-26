import { json, type RequestHandler } from '@sveltejs/kit';
import { getAnalyticsDb } from '$lib/server/analytics-db';

// Headers to exclude from analytics (invasive or infrastructure-related)
const EXCLUDED_HEADERS = new Set([
  // IP/location headers
  'x-forwarded-for',
  'x-real-ip',
  'cf-connecting-ip',
  'cf-client-ip',
  'x-client-ip',
  'x-original-ip',
  // Cloudflare/Vercel headers
  'cf-ray',
  'cf-request-id',
  'x-vercel-id',
  'x-vercel-deployment-url',
  'x-vercel-forwarded-for',
  // Auth/cookies
  'cookie',
  'authorization',
  'x-auth-token',
  // Other potentially invasive headers
  'x-api-key',
  'x-api-version',
  'x-custom-header'
]);

export const POST: RequestHandler = async ({ request }) => {
  try {
    const { source } = await request.json();

    // Get request headers to store, filtering out invasive ones
    const headers: Record<string, string> = {};
    request.headers.forEach((value, key) => {
      if (!EXCLUDED_HEADERS.has(key.toLowerCase())) {
        headers[key] = value;
      }
    });

    const analyticsDb = getAnalyticsDb();
    const result = await analyticsDb.execute({
      sql: `
        INSERT INTO directory_analytics (source, clicked_cta, request_headers)
        VALUES (?, ?, ?)
      `,
      args: [source || null, 0, JSON.stringify(headers)]
    });

    // Convert BigInt to string for JSON serialization
    const analyticsId = result.lastInsertRowid ? String(result.lastInsertRowid) : null;

    return json({ success: true, analyticsId }, { status: 201 });
  } catch (error) {
    console.error('Analytics tracking error:', error);
    return json({ success: false, error: 'Failed to track analytics' }, { status: 500 });
  }
};

export const PATCH: RequestHandler = async ({ request }) => {
  try {
    const body = await request.json();
    const { id, pwa_clicked } = body;

    if (!id) {
      return json({ success: false, error: 'Missing analytics ID' }, { status: 400 });
    }

    const analyticsDb = getAnalyticsDb();
    await analyticsDb.execute({
      sql: `
        UPDATE directory_analytics
        SET clicked_cta = 1, pwa_clicked = ?, clicked_at = STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')
        WHERE id = ?
      `,
      args: [pwa_clicked || null, id]
    });

    return json({ success: true });
  } catch (error) {
    console.error('Analytics update error:', error);
    return json({ success: false, error: 'Failed to update analytics' }, { status: 500 });
  }
};
