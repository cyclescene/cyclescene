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
    const response = await fetch(`${API_URL}/v1/rides/edit/${encodeURIComponent(token)}`, {
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
  updateRide: async ({ request, url }) => {
    const formData = await request.formData();

    // Parse occurrences if they were sent as JSON string
    const occurrencesJson = formData.get('occurrences');
    if (occurrencesJson && typeof occurrencesJson === 'string') {
      try {
        const occurrences = JSON.parse(occurrencesJson);
        formData.delete('occurrences');
        // Re-add as individual FormData entries for superValidate
        occurrences.forEach((occ: any, idx: number) => {
          Object.entries(occ).forEach(([key, value]) => {
            if (value !== undefined && value !== null) {
              formData.append(`occurrences[${idx}].${key}`, String(value));
            }
          });
        });
      } catch (e) {
        return fail(400, { error: 'Invalid occurrences data' });
      }
    }

    // Validate form data with schema
    const form = await superValidate(formData, zod(rideSubmissionSchema));

    if (!form.valid) {
      console.error('Validation errors:', form.errors);
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
      const response = await fetch(`${API_URL}/v1/rides/edit/${encodeURIComponent(token)}`, {
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
  },

  updateEventDetails: async ({ request, url }) => {
    const formData = await request.formData();
    const token = url.searchParams.get('token');
    const description = formData.get('description');
    const audience = formData.get('audience');
    const rideLength = formData.get('ride_length');

    if (!token || !description) {
      return fail(400, { error: 'Missing required fields' });
    }

    try {
      const response = await fetch(
        `${API_URL}/v1/rides/edit/${encodeURIComponent(token)}/details`,
        {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            description: description,
            audience: audience || '',
            ride_length: rideLength || '',
          }),
        }
      );

      if (!response.ok) {
        const text = await response.text();
        console.error('Failed to update event details:', text);
        return fail(response.status, { error: 'Failed to update event details' });
      }

      return { success: true };
    } catch (err) {
      return fail(500, {
        error: err instanceof Error ? err.message : 'Failed to save changes'
      });
    }
  },

  updateOccurrence: async ({ request, url }) => {
    const formData = await request.formData();
    const token = url.searchParams.get('token');
    const occurrenceId = formData.get('occurrence_id');
    const startTime = formData.get('start_time');
    const eventDurationMinutes = formData.get('event_duration_minutes');
    const eventTimeDetails = formData.get('event_time_details');
    const newsflash = formData.get('newsflash');
    const isCancelled = formData.get('is_cancelled') === 'true';

    if (!token || !occurrenceId) {
      return fail(400, { error: 'Missing required fields' });
    }

    try {
      const response = await fetch(
        `${API_URL}/v1/rides/edit/${encodeURIComponent(token)}/occurrences/${occurrenceId}`,
        {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            start_time: startTime,
            event_duration_minutes: parseInt(eventDurationMinutes as string, 10),
            event_time_details: eventTimeDetails || '',
            newsflash: newsflash || '',
            is_cancelled: isCancelled,
          }),
        }
      );

      if (!response.ok) {
        return fail(response.status, { error: 'Failed to update occurrence' });
      }

      return { success: true };
    } catch (err) {
      return fail(500, {
        error: err instanceof Error ? err.message : 'Failed to save changes'
      });
    }
  }
} satisfies Actions;
