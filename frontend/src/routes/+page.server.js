/** @type {import('./$types').PageServerLoad} */
export async function load({ fetch }) {
	try {
        // Because this runs server-side inside Docker, we can use the internal Docker network hostname 'api'
		const res = await fetch('http://api:8080/api/health');
		const data = await res.json();
		return {
			status: data.status,
			message: data.message,
			error: data.error
		};
	} catch (e) {
		return {
			status: 'error',
			message: '',
			error: e.message
		};
	}
}
