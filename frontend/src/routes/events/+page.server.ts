import type { PageServerLoad } from './$types';
import { apiFetch, getApiBaseUrl } from '$lib/api/http';

type DirectoryFilter = 'all' | 'happening';
type EventStatus = 'draft' | 'active' | 'full' | 'cancelled';

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
	status?: EventStatus;
	organizer_name?: string;
}

interface DirectoryEvent {
	id: string;
	slug: string;
	title: string;
	posterUrl: string | null;
	venue: string;
	dateLabel: string;
	timeLabel: string;
	status: EventStatus;
	statusLabel: string;
	capacity: number;
	attendeeCount: number;
	tags: string[];
}

function buildApiUrl(path: string): string {
	const baseUrl = getApiBaseUrl();
	if (!baseUrl) {
		return path;
	}

	return `${baseUrl}${path}`;
}

function normalizeFilter(value: string | null): DirectoryFilter {
	return value === 'happening' ? 'happening' : 'all';
}

function formatDate(iso: string): string {
	const date = new Date(iso);
	if (Number.isNaN(date.getTime())) {
		return 'Date TBA';
	}

	return date.toLocaleDateString('en-IN', {
		month: 'short',
		day: 'numeric'
	});
}

function formatTime(iso: string): string {
	const date = new Date(iso);
	if (Number.isNaN(date.getTime())) {
		return 'Time TBA';
	}

	return date.toLocaleTimeString('en-IN', {
		hour: '2-digit',
		minute: '2-digit',
		hour12: true
	});
}

function toStatusLabel(status: EventStatus): string {
	switch (status) {
	case 'active':
		return 'Live';
	case 'full':
		return 'Full';
	case 'cancelled':
		return 'Cancelled';
	default:
		return 'Draft';
	}
}

function mapEvent(event: ApiEvent): DirectoryEvent {
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

function parseErrorMessage(payload: unknown): string {
	if (
		typeof payload === 'object' &&
		payload !== null &&
		'error' in payload &&
		typeof payload.error === 'string'
	) {
		return payload.error;
	}

	return 'Could not load events right now.';
}

export const load: PageServerLoad = async ({ fetch, url, depends }) => {
	const filter = normalizeFilter(url.searchParams.get('filter'));
	const search = url.searchParams.get('search')?.trim() ?? '';
	depends(`events:public:${filter}:${search}`);

	const apiUrl = search
		? buildApiUrl(`/api/public/events?status=published&search=${encodeURIComponent(search)}`)
		: buildApiUrl('/api/public/events?status=published');

	const res = await apiFetch(fetch, apiUrl);

	if (!res.ok) {
		const payload = await res.json().catch(() => ({}));
		return {
			filter,
			events: [] as DirectoryEvent[],
			error: parseErrorMessage(payload)
		};
	}

	const payload = await res.json().catch(() => ({}));
	const rawEvents =
		typeof payload === 'object' &&
		payload !== null &&
		Array.isArray((payload as { events?: unknown[] }).events)
			? ((payload as { events: ApiEvent[] }).events ?? [])
			: [];

	const allEvents = rawEvents
		.filter((event): event is ApiEvent => {
			return (
				typeof event === 'object' &&
				event !== null &&
				typeof event.id === 'string' &&
				typeof event.slug === 'string' &&
				typeof event.title === 'string'
			);
		})
		.map(mapEvent);

	const events =
		filter === 'happening'
			? allEvents.filter((event) => event.status === 'active')
			: allEvents;

	return {
		filter,
		events,
		error: ''
	};
};
