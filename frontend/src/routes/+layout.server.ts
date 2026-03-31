import type { LayoutServerLoad } from './$types';
import { getApiBaseUrl } from '$lib/api/http';

export interface SessionUser {
	id: string;
	name: string;
	email: string;
	role: 'organizer' | 'admin' | 'participant';
}

export interface GuestUser {
	email: string;
	name: string;
	registered_event_ids: string[];
}

export const load: LayoutServerLoad = async ({ fetch, cookies }) => {
	const jwt = cookies.get('eventio_jwt');
	const guestSession = cookies.get('guest_session');
	const baseUrl = getApiBaseUrl();

	let user: SessionUser | null = null;
	let guest: GuestUser | null = null;

	if (jwt) {
		try {
			const res = await fetch(`${baseUrl}/api/auth/me`, {
				headers: { Cookie: `eventio_jwt=${jwt}` }
			});

			if (res.ok) {
				const data = await res.json();
				const raw = data.user ?? data.organizer ?? null;
				if (raw && raw.id) {
					user = {
						id: raw.id,
						name: raw.name,
						email: raw.email,
						role: raw.role
					};
				}
			}
		} catch {}
	}

	if (guestSession && !user) {
		try {
			const res = await fetch(`${baseUrl}/api/guest/session`, {
				headers: { Cookie: `guest_session=${guestSession}` }
			});

			if (res.ok) {
				const data = await res.json();
				if (data.guest && data.guest.email) {
					guest = {
						email: data.guest.email,
						name: data.guest.name,
						registered_event_ids: data.guest.registered_event_ids || []
					};
				}
			}
		} catch {}
	}
	return { user, guest };
};
