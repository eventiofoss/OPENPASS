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

export interface AttendeeExportFile {
	blob: Blob;
	fileName: string;
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

function getExportFileName(
	contentDisposition: string | null,
	fallback: string
): string {
	if (!contentDisposition) {
		return fallback;
	}

	const utf8Match = contentDisposition.match(
		/filename\*=UTF-8''([^;]+)/i
	);
	if (utf8Match?.[1]) {
		return decodeURIComponent(utf8Match[1]);
	}

	const quotedMatch = contentDisposition.match(
		/filename="([^"]+)"/i
	);
	if (quotedMatch?.[1]) {
		return quotedMatch[1];
	}

	const plainMatch = contentDisposition.match(
		/filename=([^;]+)/i
	);
	if (plainMatch?.[1]) {
		return plainMatch[1].trim();
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

/** Downloads the attendee CSV for an organizer-owned event. */
export async function downloadAttendeeExport(
	fetchFn: typeof fetch,
	eventId: string,
	fallbackFileName = 'attendees.csv'
): Promise<AttendeeExportFile> {
	const res = await apiFetch(
		fetchFn,
		buildUrl(`/${eventId}/export`)
	);

	if (!res.ok) {
		throw new Error(
			await handleAuthError(
				res,
				'Failed to export attendee data.'
			)
		);
	}

	return {
		blob: await res.blob(),
		fileName: getExportFileName(
			res.headers.get('content-disposition'),
			fallbackFileName
		)
	};
}
