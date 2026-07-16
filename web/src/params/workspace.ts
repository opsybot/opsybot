import type { ParamMatcher } from '@sveltejs/kit';

export const match: ParamMatcher = (value) => /^[a-z][a-z0-9-]{0,39}$/.test(value);
