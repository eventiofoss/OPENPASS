export const organizerEventsDependency = 'organizer:events';

/** Returns the dependency key for an organizer-owned event detail. */
export function organizerEventDependency(eventId: string): string {
	return `organizer:event:${eventId}`;
}

/** Returns the dependency key for organizer event analytics. */
export function organizerAnalyticsDependency(eventId: string): string {
	return `organizer:analytics:${eventId}`;
}
