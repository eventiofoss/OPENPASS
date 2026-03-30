import { apiFetch, getApiBaseUrl } from './http';

/** Analytics counters returned by the backend. */
export interface AnalyticsData {
	tickets_sold: number;
	check_in_count: number;
	total_revenue: number;
	capacity: number;
	capacity_pct: number;
}

/** Organizer event summary for the sidebar list. */
export interface OrganizerEvent {
	id: string;
	title: string;
	slug: string;
	status: 'draft' | 'active' | 'full' | 'cancelled';
	start_date: string;
	venue: string;
	capacity: number;
	is_public: boolean;
}

/** Full event detail for the dashboard header. */
export interface OrganizerEventDetail extends OrganizerEvent {
	description: string;
	price: number;
	tickets_sold: number;
	check_in_count: number;
	total_revenue: number;
}

function buildUrl(path: string): string {
	const baseUrl = getApiBaseUrl();

	if (typeof window === 'undefined') {
		return `${baseUrl}/api/events${path}`;
	}

	return `/api/events${path}`;
}

async function handleAuthError(
	res: Response,
	fallback: string
): Promise<string> {
	if (res.status === 401 && typeof window !== 'undefined') {
		window.location.href = '/organizer/login';
		return 'Session expired. Please sign in again.';
	}

	const data = await res.json().catch(() => ({}));

	if (
		typeof data === 'object' &&
		data !== null &&
		'error' in data &&
		typeof data.error === 'string'
	) {
		return data.error;
	}

	return fallback;
}

/** Fetches the analytics dashboard data for an event. */
export async function getEventAnalytics(
	fetchFn: typeof fetch,
	eventId: string
): Promise<AnalyticsData> {
	const res = await apiFetch(fetchFn, buildUrl(`/${eventId}/analytics`));

	if (!res.ok) {
		throw new Error(
			await handleAuthError(
				res,
				'Failed to fetch analytics data.'
			)
		);
	}

	const data = await res.json();
	return data.analytics as AnalyticsData;
}

/** Lists all events owned by the authenticated organizer. */
export async function getOrganizerEvents(
	fetchFn: typeof fetch
): Promise<OrganizerEvent[]> {
	const res = await apiFetch(fetchFn, buildUrl('/'));

	if (!res.ok) {
		throw new Error(
			await handleAuthError(
				res,
				'Failed to fetch events.'
			)
		);
	}

	const data = await res.json();
	return (data.events ?? []) as OrganizerEvent[];
}

/** Fetches a single organizer-owned event detail. */
export async function getEventDetail(
	fetchFn: typeof fetch,
	eventId: string
): Promise<OrganizerEventDetail> {
	const res = await apiFetch(fetchFn, buildUrl(`/${eventId}`));

	if (!res.ok) {
		throw new Error(
			await handleAuthError(
				res,
				'Failed to fetch event details.'
			)
		);
	}

	const data = await res.json();
	return data.event as OrganizerEventDetail;
}

/** Returns the URL for downloading the attendee CSV. */
export function getExportUrl(eventId: string): string {
	return buildUrl(`/${eventId}/export`);
}
