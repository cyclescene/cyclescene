// src/routes/+page.server.ts
import { superValidate } from 'sveltekit-superforms';
import { zod4 as zod } from 'sveltekit-superforms/adapters';
import { fail, redirect } from '@sveltejs/kit';
import { rideSubmissionSchema } from '$lib/schemas/ride';
import type { PageServerLoad, Actions } from './$types';
import { API_URL, API_DEV_MODE } from '$env/static/private';


export const load: PageServerLoad = async ({ url, request }) => {
  const token = url.searchParams.get('token');
  const city = url.searchParams.get('city');
  const isDevMode = API_DEV_MODE === 'true';

  // In dev mode, allow access without token but require city
  if (isDevMode) {
    if (!city) {
      throw redirect(302, '/error?message=Missing city parameter');
    }
    console.log('[DEV MODE] Skipping token validation');
  } else {
    // Production mode: require both token and city
    if (!token || !city) {
      throw redirect(302, '/error?message=Missing token or city');
    }

    // Check referrer to ensure request came from PWA
    const referrer = request.headers.get('referer') || '';
    const validReferrers = [
      'https://pdx.cyclescene.cc',
      'https://slc.cyclescene.cc',
      'http://localhost' // for dev only
    ];

    const isValidReferrer = validReferrers.some(valid => referrer.startsWith(valid));

    if (!isValidReferrer) {
      throw redirect(302, '/error?message=Invalid referrer');
    }

    try {
      // Validate the token with your API - use full URL for server-side fetch
      const response = await fetch(`${API_URL}/v1/tokens/validate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ token, city })
      });

      if (!response.ok) {
        throw redirect(302, '/error?message=Token validation failed');
      }

      const validation = await response.json();

      if (!validation.valid) {
        throw redirect(302, '/error?message=Invalid or expired token');
      }
    } catch (err) {
      console.error('Token validation failed:', err);
      throw redirect(302, '/error?message=Token validation failed');
    }
  }

  // Initialize form with city pre-filled
  // Don't validate on initial load (errors: false)
  const form = await superValidate(
    {
      city,
      title: '',
      description: '',
      web_url: '',
      web_name: '',
      venue_name: '',
      address: '',
      organizer_name: '',
      organizer_email: '',
      audience: 'G', // Default to General
      date_type: 'S' as const,
      is_loop_ride: false,
      hide_email: false,
      hide_phone: false,
      hide_contact_name: false,
      occurrences: []
    },
    zod(rideSubmissionSchema),
    { errors: false } // Don't show errors on initial load
  );

  return {
    form,
    token,
    city
  };
};

export const actions = {
  default: async ({ request, url }) => {
    const formData = await request.formData();

    // Validate form data with schema
    const form = await superValidate(formData, zod(rideSubmissionSchema));

    if (!form.valid) {
      return fail(400, { form });
    }

    const isDevMode = API_DEV_MODE === 'true';
    const token = url.searchParams.get('token');

    // In production, token is required
    if (!isDevMode && !token) {
      return fail(400, {
        form,
        error: 'Missing submission token'
      });
    }

    // In dev mode, use a dummy token if none provided
    const submissionToken = isDevMode && !token ? 'dev-mode-token' : token;

    if (isDevMode && !token) {
      console.log('[DEV MODE] Using dummy token for submission');
    }

    const response = await fetch(`${API_URL}/v1/rides/submit`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-BFF-Token': submissionToken!,
      },
      body: JSON.stringify(form.data)
    }).catch(err => {
      return fail(500, {
        form,
        error: err instanceof Error ? err.message : 'An error occurred'
      });
    }) as Response

    if ('status' in response && response.status === 500) {
      return response
    }

    const result = await response.json() as { success: true; event_id: number; edit_token: string }


    if (result.success) {
      const city = form.data.city || 'pdx';
      throw redirect(303, `/success?token=${result.edit_token}&event_id=${result.event_id}&city=${city}`);
    }

  }
} satisfies Actions;
