import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ url }) => ({
	code: url.searchParams.get('code') ?? 'error'
});
