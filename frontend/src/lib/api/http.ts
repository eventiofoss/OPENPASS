import { env as publicEnv } from '$env/dynamic/public';

const csrfCookieName = 'eventio_csrf';
const csrfBootstrapPath = '/api/health';

let csrfBootstrapPromise: Promise<void> | null = null;

function isMutatingMethod(method: string): boolean {
	switch (method.toUpperCase()) {
	case 'POST':
	case 'PUT':
	case 'PATCH':
	case 'DELETE':
		return true;
	default:
		return false;
	}
}

function getCookieValue(name: string): string {
	if (typeof document === 'undefined' || document.cookie === '') {
		return '';
	}

	const encodedName = `${encodeURIComponent(name)}=`;
	const cookies = document.cookie.split('; ');

	for (const cookie of cookies) {
		if (cookie.startsWith(encodedName)) {
			return decodeURIComponent(cookie.slice(encodedName.length));
		}
	}

	return '';
}

function buildApiUrl(path: string): string {
	const baseUrl = getApiBaseUrl();
	if (!baseUrl) {
		return path;
	}

	return `${baseUrl}${path}`;
}

async function ensureCsrfCookie(fetchFn: typeof fetch): Promise<void> {
	if (typeof window === 'undefined') {
		return;
	}

	if (getCookieValue(csrfCookieName) !== '') {
		return;
	}

	if (!csrfBootstrapPromise) {
		csrfBootstrapPromise = (async () => {
			const res = await fetchFn(buildApiUrl(csrfBootstrapPath), {
				method: 'GET',
				credentials: 'include',
				headers: {
					'Cache-Control': 'no-cache'
				}
			});

			if (!res.ok) {
				throw new Error(
					`Failed to initialize CSRF session: ${res.status}`
				);
			}
		})().finally(() => {
			csrfBootstrapPromise = null;
		});
	}

	await csrfBootstrapPromise;
}

/**
 * During SSR, prefer INTERNAL_API_BASE_URL for container networking.
 * In the browser, use PUBLIC_API_URL or same-origin relative requests.
 */
export function getApiBaseUrl(): string {
	return publicEnv.PUBLIC_API_URL || '';
}

/**
 * Shared API fetch wrapper for consistent cookie and CSRF handling.
 */
export async function apiFetch(
	fetchFn: typeof fetch,
	input: RequestInfo | URL,
	init: RequestInit = {}
): Promise<Response> {
	const method = (init.method || 'GET').toUpperCase();
	const headers = new Headers(init.headers);

	if (isMutatingMethod(method) && !headers.has('X-CSRF-Token') && typeof window !== 'undefined') {
		await ensureCsrfCookie(fetchFn);

		const csrfToken = getCookieValue(csrfCookieName);
		if (!csrfToken) {
			throw new Error('CSRF token cookie is missing for mutating request');
		}

		headers.set('X-CSRF-Token', csrfToken);
	}

	return fetchFn(input, {
		...init,
		method,
		headers,
		credentials: 'include'
	});
}
