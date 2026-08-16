import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';

/** Databases module removed — Data Access no longer links external databases. */
export const load: LayoutLoad = () => {
	throw redirect(302, '/data-tables');
};
