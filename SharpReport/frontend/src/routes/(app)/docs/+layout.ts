import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';

/** Docs module removed from Data Access — send legacy URLs to Data tables. */
export const load: LayoutLoad = () => {
	throw redirect(302, '/data-tables');
};
