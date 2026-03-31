import type { PageServerLoad } from './$types';
import { apiFetch, getApiBaseUrl } from '$lib/api/http';

interface ApiEvent {
	id: string;
	slug: string;
	title: string;
	description?: string;
	poster_url?: string | null;
	venue?: string;
	start_date?: string;
	capacity?: number;
	total_registered?: number;
	status?: string;
	organizer_name?: string;
}

interface LandingEvent {
	id: string;
	slug: string;
	title: string;
	posterUrl: string | null;
	venue: string;
	dateLabel: string;
	timeLabel: string;
	status: string;
	statusLabel: string;
	capacity: number;
	attendeeCount: number;
	tags: string[];
}

function buildApiUrl(path: string): string {
	const baseUrl = getApiBaseUrl();
	return baseUrl ? `${baseUrl}${path}` : path;
}

function formatDate(iso: string): string {
	const date = new Date(iso);
	if (Number.isNaN(date.getTime())) return 'Date TBA';
	return date.toLocaleDateString('en-IN', { month: 'short', day: 'numeric' });
}

function formatTime(iso: string): string {
	const date = new Date(iso);
	if (Number.isNaN(date.getTime())) return 'Time TBA';
	return date.toLocaleTimeString('en-IN', {
		hour: '2-digit',
		minute: '2-digit',
		hour12: true
	});
}

function toStatusLabel(status: string): string {
	switch (status) {
		case 'active': return 'Live';
		case 'full': return 'Full';
		case 'cancelled': return 'Cancelled';
		default: return 'Draft';
	}
}

function mapEvent(event: ApiEvent): LandingEvent {
	const status = event.status ?? 'draft';
	const organizer = event.organizer_name?.trim() || 'Open Pass';

	return {
		id: event.id,
		slug: event.slug,
		title: event.title,
		posterUrl: event.poster_url ?? null,
		venue: event.venue?.trim() || 'Venue TBA',
		dateLabel: formatDate(event.start_date ?? ''),
		timeLabel: formatTime(event.start_date ?? ''),
		status,
		statusLabel: toStatusLabel(status),
		capacity: event.capacity ?? 0,
		attendeeCount: event.total_registered ?? 0,
		tags: [organizer]
	};
}


export const load: PageServerLoad = async ({ fetch }) => {
	try {
		const res = await apiFetch(
			fetch,
			buildApiUrl('/api/public/events?status=published')
		);

		if (!res.ok) {
			return { happeningNow: [] as LandingEvent[], discoverNearby: [] as LandingEvent[] };
		}

		const payload = await res.json().catch(() => ({}));
		const rawEvents: ApiEvent[] =
			typeof payload === 'object' &&
			payload !== null &&
			Array.isArray((payload as { events?: unknown[] }).events)
				? ((payload as { events: ApiEvent[] }).events ?? [])
				: [];

		const validEvents = rawEvents.filter(
			(e): e is ApiEvent =>
				typeof e === 'object' &&
				e !== null &&
				typeof e.id === 'string' &&
				typeof e.slug === 'string' &&
				typeof e.title === 'string'
		);

		// "Happening Now": next 4 upcoming active events (start_date >= now),
		// sorted by soonest first. Relaxed from the previous ±24hr window
		// so that seed data and future events always appear.
		const now = new Date();
		const upcoming = validEvents
			.filter((e) => {
				if (!e.start_date || e.status !== 'active') return false;
				const d = new Date(e.start_date);
				return !Number.isNaN(d.getTime()) && d.getTime() >= now.getTime();
			})
			.sort((a, b) =>
				new Date(a.start_date!).getTime() - new Date(b.start_date!).getTime()
			);

		const happeningNow = upcoming.slice(0, 4).map(mapEvent);
		const discoverNearby = validEvents.map(mapEvent);

		return { happeningNow, discoverNearby };
	} catch {
		return { happeningNow: [] as LandingEvent[], discoverNearby: [] as LandingEvent[] };
	}
};
