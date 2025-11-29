import { superValidate } from 'sveltekit-superforms';
import { zod4 as zod } from 'sveltekit-superforms/adapters';
import { fail, redirect } from '@sveltejs/kit';
import { groupRegistrationSchema } from '$lib/schemas/ride';
import { getGroupByEditToken, updateGroup } from '$lib/api/client';
import type { PageServerLoad, Actions } from './$types';

export const load: PageServerLoad = async ({ url }) => {
  const token = url.searchParams.get('token');

  if (!token) {
    throw redirect(302, '/error?message=Missing edit token');
  }

  try {
    // Fetch existing group data
    const groupData = await getGroupByEditToken(token);

    // Initialize form with existing data
    const form = await superValidate(zod(groupRegistrationSchema), {
      data: {
        code: groupData.code || '',
        name: groupData.name || '',
        description: groupData.description || '',
        city: groupData.city || '',
        image_uuid: groupData.image_uuid || '',
        web_url: groupData.web_url || ''
      }
    });

    return {
      form,
      token,
      city: groupData.city,
      groupCode: groupData.code
    };
  } catch (err) {
    console.error('Failed to fetch group:', err);
    throw redirect(302, '/error?message=Failed to load group information');
  }
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
        error: 'Missing edit token'
      });
    }

    // Only allow updating name, description, web_url, and image_uuid
    const updatePayload: Record<string, any> = {
      name: form.data.name,
      description: form.data.description,
      web_url: form.data.web_url
    };

    // Include image_uuid if a new marker was uploaded
    if (form.data.image_uuid) {
      updatePayload.image_uuid = form.data.image_uuid;
    }

    try {
      await updateGroup(token, updatePayload);
      throw redirect(303, `/group/success?token=${token}&code=${form.data.code}&edited=true`);
    } catch (err) {
      return fail(500, {
        form,
        error: err instanceof Error ? err.message : 'Failed to update group'
      });
    }
  }
};
