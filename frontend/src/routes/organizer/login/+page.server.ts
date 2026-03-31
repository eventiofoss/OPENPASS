import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent }) => {
    const { user } = await parent();
    
    // If the user is already logged in, don't show the login form
    if (user) {
        throw redirect(302, '/organizer/dashboard');
    }
    return {};
};
