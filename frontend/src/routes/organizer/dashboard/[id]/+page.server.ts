import type { PageServerLoad } from './$types';
import {
	getEventAnalytics,
	getEventDetail,
} from '$lib/api/analytics';
import { error } from '@sveltejs/kit';

export const load: PageServerLoad = async ({
	params,
	fetch,
}) => {
	const eventId = params.id;

	try {
		const [eventDetail, analytics] = await Promise.all([
			getEventDetail(fetch, eventId),
			getEventAnalytics(fetch, eventId),
		]);

		return { eventDetail, analytics };
	} catch (err) {
		if (
			err instanceof Error &&
			err.message.includes('not found')
		) {
			throw error(404, 'Event not found');
		}

		throw error(500, 'Failed to load dashboard data');
	}
};
