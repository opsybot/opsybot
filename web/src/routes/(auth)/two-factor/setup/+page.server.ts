import { fail, redirect } from '@sveltejs/kit';
import { toString as qrToString } from 'qrcode';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { totpSchema } from '$lib/schemas/auth';
import { RECOVERY_CODES, TOTP_SECRET, verifyTotp } from '$lib/server/auth';
import type { Actions, PageServerLoad } from './$types';

const ACCOUNT = 'maya@acme.dev';

function groupSecret(secret: string): string {
	return secret.replace(/(.{4})/g, '$1 ').trim();
}

export const load: PageServerLoad = async ({ url }) => {
	const step = url.searchParams.get('step') === 'recovery' ? 'recovery' : 'scan';

	const form = await superValidate(zod4(totpSchema));

	if (step === 'recovery') {
		return { step, secret: '', groupedSecret: '', qr: '', recoveryCodes: RECOVERY_CODES, form };
	}

	const uri = `otpauth://totp/Opsybot:${ACCOUNT}?secret=${TOTP_SECRET}&issuer=Opsybot`;
	const qr = await qrToString(uri, {
		type: 'svg',
		margin: 1,
		width: 148,
		color: { dark: '#0a0f11', light: '#ffffff' }
	});

	return {
		step,
		secret: TOTP_SECRET,
		groupedSecret: groupSecret(TOTP_SECRET),
		qr,
		recoveryCodes: [],
		form
	};
};

export const actions: Actions = {
	verify: async ({ request }) => {
		const form = await superValidate(request, zod4(totpSchema));
		if (!form.valid) return fail(400, { form });
		if (!verifyTotp(form.data.code)) return message(form, 'wrong', { status: 400 });

		redirect(303, '/two-factor/setup?step=recovery');
	},
	finish: async () => {
		redirect(303, '/');
	}
};
