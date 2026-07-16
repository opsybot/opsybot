import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => ({
	failure: {
		code: 'invalid_signature',
		entityId: 'https://sso.acme.dev/saml/metadata',
		requestId: 'saml_9f27_20260711T0914Z',
		at: '2026-07-11 09:14 UTC'
	}
});
