import { env } from '$env/dynamic/private';
import { env as publicEnv } from '$env/dynamic/public';

/**
 * Helper to determine the correct API base URL.
 * During SSR (server-side), it prefers INTERNAL_API_BASE_URL (for Docker networking).
 * On the client, it uses PUBLIC_API_URL or a local relative path.
 */
export function getApiBaseUrl(): string {
	if (typeof window === 'undefined') {
		// Server-side
		return env.INTERNAL_API_BASE_URL || 'http://localhost:8080';
	}
	// Client-side
	return publicEnv.PUBLIC_API_URL || ''; // Empty string means it will be relative to current origin, handled by a reverse proxy
}

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

export async function getPublicEvent(
	fetchFn: typeof fetch,
	slug: string
): Promise<PublicEvent> {
	const baseUrl = getApiBaseUrl();
	// For client-side requests, assuming Vite's proxy or Caddy forwards `/api` directly to backend
	const url = typeof window === 'undefined' ? `${baseUrl}/api/public/events/${slug}` : `/api/public/events/${slug}`;

	const res = await fetchFn(url);

	if (!res.ok) {
		if (res.status === 404) {
			throw new Error('Event not found');
		}
		throw new Error(`Failed to fetch event: ${res.statusText}`);
	}

	const data = await res.json();
	return data.event as PublicEvent;
}
