import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

/** "new" is not a dashboard id; it used to match `[id]` by mistake. */
export const load: PageLoad = () => {
	throw redirect(303, '/dashboards');
};
