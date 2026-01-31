import { superValidate } from 'sveltekit-superforms';
import { zod4 as zod } from 'sveltekit-superforms/adapters';
import { fail, redirect } from '@sveltejs/kit';
import { groupRegistrationSchema } from '$lib/schemas/ride';
import { validateSubmissionToken, registerGroup } from '$lib/api/client';
import type { PageServerLoad, Actions } from './$types';
import { API_DEV_MODE } from '$env/static/private';

export const load: PageServerLoad = async ({ url, request }) => {

  const token = url.searchParams.get('token');
  const city = url.searchParams.get('city');
  const isDevMode = API_DEV_MODE === 'true';

  // In dev mode, allow access without token but require city
  if (isDevMode) {
    if (!city) {
      throw redirect(302, '/error?message=Missing city parameter');
    }
    console.log('[DEV MODE] Skipping token validation for group registration');
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
      // Validate the token with your API
      const validation = await validateSubmissionToken(token, city);

      if (!validation.valid) {
        throw redirect(302, '/error?message=Invalid or expired token');
      }
    } catch (err) {
      console.error('Token validation failed:', err);
      throw redirect(302, '/error?message=Token validation failed');
    }
  }

  // Initialize form with city pre-filled
  const form = await superValidate(zod(groupRegistrationSchema), {
    defaults: {
      city,
      code: '',
      name: '',
      description: '',
      email: '',
      web_url: ''
    }
  });

  return {
    form,
    token,
    city
  };
};

export const actions: Actions = {
  default: async ({ request, url }) => {
    const formData = await request.formData();
    const form = await superValidate(formData, zod(groupRegistrationSchema));

    if (!form.valid) {
      return fail(400, { form });
    }

    const isDevMode = API_DEV_MODE === 'true';
    const token = url.searchParams.get('token');
    const city = url.searchParams.get('city');

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
      console.log('[DEV MODE] Using dummy token for group registration');
    }

    const response = await registerGroup(form.data, submissionToken!)
      .catch((err) => {
        return fail(500, {
          form,
          error: err instanceof Error ? err.message : 'An error occurred'
        });
      })

    if ('status' in response && response.status === 500) {
      return response
    }

    throw redirect(303, `/group/success?token=${response.edit_token}&code=${response.code}&city=${city}`);

  }
};
