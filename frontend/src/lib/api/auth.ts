import { apiFetch } from './http';

export interface RegisterParams {
	name: string;
	email: string;
	password?: string; // Optional for attendees creating accounts, but the backend requires password for an organizer schema. So wait, backend requires password.
	role?: 'organizer' | 'volunteer' | 'participant';
}

export async function registerUser(fetchFn: typeof fetch, params: RegisterParams): Promise<void> {
	const res = await apiFetch(fetchFn, '/api/auth/register', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify(params)
	});

	if (!res.ok) {
		const data = await res.json().catch(() => ({}));
		throw new Error(data.error || 'Failed to create account');
	}
}

export async function loginUser(fetchFn: typeof fetch, params: Partial<RegisterParams>): Promise<void> {
	const res = await apiFetch(fetchFn, '/api/auth/login', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify(params)
	});

	if (!res.ok) {
		const data = await res.json().catch(() => ({}));
		throw new Error(data.error || 'Invalid credentials');
	}
}
