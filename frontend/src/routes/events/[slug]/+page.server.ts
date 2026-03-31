import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { getPublicEvent } from '$lib/api/events';

export const load: PageServerLoad = async ({ params, fetch }) => {
	try {
		const event = await getPublicEvent(fetch, params.slug);
		return { event };
	} catch (err: any) {
		throw error(404, err.message || 'Event not found');
	}
};
