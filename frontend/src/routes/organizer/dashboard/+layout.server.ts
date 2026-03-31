import type { LayoutServerLoad } from './$types';
import { getOrganizerEvents } from '$lib/api/analytics';
import { organizerEventsDependency } from '$lib/utils/organizer-dashboard';
import { redirect } from '@sveltejs/kit';

export const load: LayoutServerLoad = async ({
	fetch,
	cookies,
	depends,
}) => {
	depends(organizerEventsDependency);

	try {
		const events = await getOrganizerEvents(fetch);
		return { events };
	} catch {
		throw redirect(302, '/organizer/login');
	}
};
