import { apiFetch, getApiBaseUrl } from './http';

export interface PublicEvent {
	slug: string;
	title: string;
	description: string;
	poster_url: string | null;
	venue: string;
	start_date: string;
	capacity: number;
	total_registered: number;
	price: number;
	status: 'draft' | 'active' | 'full' | 'cancelled';
	organizer_name: string;
	organizer_email: string;
}

export type FormFieldType =
	| 'text'
	| 'email'
	| 'select'
	| 'checkbox'
	| 'number';

export interface CreateOrganizerEventInput {
	title: string;
	description: string;
	start_date: string;
	venue: string;
	capacity: number;
	price: number;
	is_public: boolean;
}

export interface OrganizerEventRecord {
	id: string;
	slug: string;
}

export interface OrganizerFormFieldInput {
	name: string;
	type: FormFieldType;
	label: string;
	required: boolean;
	options?: string[];
}

function getOrganizerEventsUrl(path = ''): string {
	const baseUrl = getApiBaseUrl();

	if (typeof window === 'undefined') {
		return `${baseUrl}/api/events${path}`;
	}

	return `/api/events${path}`;
}

async function parseApiError(
	res: Response,
	fallbackMessage: string
): Promise<string> {
	if (res.status === 401 && typeof window !== 'undefined') {
		window.location.href = '/organizer/login';
		return 'Your organizer session expired. Please sign in again.';
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

	return fallbackMessage;
}

export async function getPublicEvent(
	fetchFn: typeof fetch,
	slug: string
): Promise<PublicEvent> {
	const baseUrl = getApiBaseUrl();
	// For client-side requests, assuming Vite's proxy or Caddy forwards `/api` directly to backend
	const url = typeof window === 'undefined' ? `${baseUrl}/api/public/events/${slug}` : `/api/public/events/${slug}`;

	const res = await apiFetch(fetchFn, url);

	if (!res.ok) {
		if (res.status === 404) {
			throw new Error('Event not found');
		}
		throw new Error(`Failed to fetch event: ${res.statusText}`);
	}

	const data = await res.json();
	return data.event as PublicEvent;
}

/** Creates a new organizer-owned event draft. */
export async function createOrganizerEvent(
	fetchFn: typeof fetch,
	input: CreateOrganizerEventInput
): Promise<OrganizerEventRecord> {
	const res = await apiFetch(fetchFn, getOrganizerEventsUrl(), {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify(input)
	});

	if (!res.ok) {
		throw new Error(
			await parseApiError(
				res,
				'Failed to create the event draft.'
			)
		);
	}

	const data = await res.json();
	return data.event as OrganizerEventRecord;
}

/** Replaces the attendee registration form schema for an event. */
export async function setOrganizerEventFormFields(
	fetchFn: typeof fetch,
	eventId: string,
	fields: OrganizerFormFieldInput[]
): Promise<OrganizerFormFieldInput[]> {
	const res = await apiFetch(
		fetchFn,
		getOrganizerEventsUrl(`/${eventId}/forms`),
		{
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ fields })
		}
	);

	if (!res.ok) {
		throw new Error(
			await parseApiError(
				res,
				'Failed to save the attendee form.'
			)
		);
	}

	const data = await res.json();
	return data.fields as OrganizerFormFieldInput[];
}
