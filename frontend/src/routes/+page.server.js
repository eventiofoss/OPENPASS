import { env } from '$env/dynamic/private';

/** @type {import('./$types').PageServerLoad} */
export async function load({ fetch }) {
	try {
		const baseUrl = env.INTERNAL_API_BASE_URL;

		if (!baseUrl) {
			return {
				status: 'error',
				message: '',
				error: 'INTERNAL_API_BASE_URL is not set'
			};
		}

		const healthUrl = `${baseUrl.replace(/\/+$/, '')}/api/health`;
		const res = await fetch(healthUrl);
		const data = await res.json();

		return {
			status: data.status ?? (res.ok ? 'ok' : 'error'),
			message: data.message ?? '',
			error: data.error ?? ''
		};
	} catch (e) {
		return {
			status: 'error',
			message: '',
			error: e instanceof Error ? e.message : 'Unknown error'
		};
	}
}
