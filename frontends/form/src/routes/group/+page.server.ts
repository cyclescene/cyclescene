import { superValidate } from 'sveltekit-superforms';
import { zod4 as zod } from 'sveltekit-superforms/adapters';
import { fail, redirect } from '@sveltejs/kit';
import { groupRegistrationSchema } from '$lib/schemas/ride';
import { validateSubmissionToken, registerGroup } from '$lib/api/client';
import type { PageServerLoad, Actions } from './$types';

export const load: PageServerLoad = async ({ url, request }) => {

  const token = url.searchParams.get('token');
  const city = url.searchParams.get('city');

  // Validate token and origin
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

  // Initialize form with city pre-filled
  const form = await superValidate(zod(groupRegistrationSchema), {
    defaults: {
      city,
      code: '',
      name: '',
      description: '',
      icon_url: '',
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

    const token = url.searchParams.get('token');
    if (!token) {
      return fail(400, {
        form,
        error: 'Missing submission token'
      });
    }


    const response = await registerGroup(form.data, token)
      .catch((err) => {
        return fail(500, {
          form,
          error: err instanceof Error ? err.message : 'An error occurred'
        });
      })

    if ('status' in response && response.status === 500) {
      return response
    }

    throw redirect(303, `/group/success?token=${response.edit_token}&code=${response.code}`);

  }
};
