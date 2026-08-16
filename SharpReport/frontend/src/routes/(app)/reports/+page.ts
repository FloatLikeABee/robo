import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

/** Reports hub redirects to the report builder. */
export const load: PageLoad = () => {
	throw redirect(302, '/reports/builder');
};
