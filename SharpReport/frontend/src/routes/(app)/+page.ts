import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

/** Home redirects to Data tables; AI chat lives in the header assistant drawer. */
export const load: PageLoad = () => {
	throw redirect(302, '/data-tables');
};
