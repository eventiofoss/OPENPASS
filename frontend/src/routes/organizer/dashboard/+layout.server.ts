import type { LayoutServerLoad } from './$types';
import { getOrganizerEvents } from '$lib/api/analytics';
import { redirect } from '@sveltejs/kit';

export const load: LayoutServerLoad = async ({
	fetch,
	cookies,
}) => {
	try {
		const events = await getOrganizerEvents(fetch);
		return { events };
	} catch {
		throw redirect(302, '/organizer/login');
	}
};
