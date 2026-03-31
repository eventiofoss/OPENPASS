/** Resolves scanned QR content into an in-app event path when possible. */
export function resolveEventPathFromQr(
	value: string,
	origin: string
): string | null {
	const trimmed = value.trim();
	if (trimmed === '') {
		return null;
	}

	try {
		const url = new URL(trimmed, origin);
		if (url.origin !== origin) {
			return null;
		}

		const path = url.pathname.replace(/\/+$/, '');
		if (
			path.startsWith('/events/') ||
			path.startsWith('/join/')
		) {
			return `${url.pathname}${url.search}`;
		}
	} catch {
		return null;
	}

	const slug = trimmed.replace(/^\/+|\/+$/g, '');
	if (/^[a-z0-9]+(?:-[a-z0-9]+)*$/i.test(slug)) {
		return `/events/${slug}`;
	}

	return null;
}
