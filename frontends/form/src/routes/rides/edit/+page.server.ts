import { superValidate } from 'sveltekit-superforms';
import { zod4 as zod } from 'sveltekit-superforms/adapters';
import { fail, redirect } from '@sveltejs/kit';
import { rideSubmissionSchema } from '$lib/schemas/ride';
import type { PageServerLoad, Actions } from './$types';
import { API_URL } from '$env/static/private';

export const load: PageServerLoad = async ({ url }) => {
  const token = url.searchParams.get('token');

  // Validate token is present
  if (!token) {
    throw redirect(302, '/error?message=Missing edit token');
  }

  try {
    // Fetch existing ride data using the edit token
    const response = await fetch(`${API_URL}/v1/rides/edit/${token}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      }
    });

    if (!response.ok) {
      throw redirect(302, '/error?message=Could not load ride. Token may be invalid or expired.');
    }

    const rideData = await response.json() as { event: any; is_published: boolean };

    // Convert ride data to form format
    const form = await superValidate(
      {
        city: rideData.event.city || '',
        title: rideData.event.title || '',
        tinytitle: rideData.event.tinytitle || '',
        description: rideData.event.description || '',
        image_url: rideData.event.image_url || '',
        image_uuid: rideData.event.image_uuid || '',
        image_srcset: rideData.event.image_srcset || '',
        audience: rideData.event.audience || '',
        ride_length: rideData.event.ride_length || '',
        area: rideData.event.area || '',
        date_type: rideData.event.date_type || 'S',
        venue_name: rideData.event.venue_name || '',
        address: rideData.event.address || '',
        location_details: rideData.event.location_details || '',
        ending_location: rideData.event.ending_location || '',
        is_loop_ride: rideData.event.is_loop_ride || false,
        organizer_name: rideData.event.organizer_name || '',
        organizer_email: rideData.event.organizer_email || '',
        organizer_phone: rideData.event.organizer_phone || '',
        web_url: rideData.event.web_url || '',
        web_name: rideData.event.web_name || '',
        newsflash: rideData.event.newsflash || '',
        hide_email: rideData.event.hide_email || false,
        hide_phone: rideData.event.hide_phone || false,
        hide_contact_name: rideData.event.hide_contact_name || false,
        group_code: rideData.event.group_code || '',
        occurrences: rideData.event.occurrences || []
      },
      zod(rideSubmissionSchema),
      { errors: false }
    );

    return {
      rideData: rideData,
      token,
      city: rideData.event.city
    };
  } catch (err) {
    console.error('Failed to load ride:', err);
    throw redirect(302, '/error?message=Failed to load ride data');
  }
};

export const actions = {
  default: async ({ request, url }) => {
    const formData = await request.formData();

    // Validate form data with schema
    const form = await superValidate(formData, zod(rideSubmissionSchema));

    if (!form.valid) {
      return fail(400, { form });
    }

    const token = url.searchParams.get('token');
    if (!token) {
      return fail(400, {
        form,
        error: 'Missing edit token'
      });
    }

    try {
      const response = await fetch(`${API_URL}/v1/rides/edit/${token}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(form.data)
      });

      if (!response.ok) {
        return fail(response.status, {
          form,
          error: 'Failed to update ride'
        });
      }

      const result = await response.json() as { success: boolean; message?: string };

      if (result.success) {
        return {
          form,
          success: true,
          message: 'Ride updated successfully! Your changes have been saved.'
        };
      }

      return fail(500, {
        form,
        error: 'Unexpected response from server'
      });
    } catch (err) {
      return fail(500, {
        form,
        error: err instanceof Error ? err.message : 'An error occurred'
      });
    }
  }
} satisfies Actions;
