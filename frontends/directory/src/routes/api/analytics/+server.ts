import { json, type RequestHandler } from '@sveltejs/kit';
import { turso } from '$lib/server/db';

export const POST: RequestHandler = async ({ request }) => {
  try {
    const { source } = await request.json();

    // Get request headers to store
    const headers: Record<string, string> = {};
    request.headers.forEach((value, key) => {
      headers[key] = value;
    });

    const result = await turso.execute({
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

    await turso.execute({
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
