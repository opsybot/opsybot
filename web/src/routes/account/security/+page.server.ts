import { fail, redirect } from '@sveltejs/kit';
import { toString as qrToString } from 'qrcode';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { changePasswordSchema, totpSchema } from '$lib/schemas/auth';
import { api } from '$lib/server/api';
import type { Actions, PageServerLoad } from './$types';

type Profile = { twoFactorEnabled: boolean };
type Enrollment = { secret: string; otpauthUri: string };
type RecoveryCodes = { codes: string[] };

function groupSecret(secret: string): string {
	return secret.replace(/(.{4})/g, '$1 ').trim();
}

export const load: PageServerLoad = async ({ cookies }) => {
	const me = await api.get<Profile>('/me', cookies);
	const enabled = me.data?.twoFactorEnabled ?? false;

	const base = {
		enabled,
		passwordForm: await superValidate(zod4(changePasswordSchema), { id: 'password' }),
		verifyForm: await superValidate(zod4(totpSchema), { id: 'verify' }),
		regenerateForm: await superValidate(zod4(totpSchema), { id: 'regenerate' }),
		disableForm: await superValidate(zod4(totpSchema), { id: 'disable' })
	};

	if (enabled) {
		return { ...base, secret: '', groupedSecret: '', qr: '', unavailable: false, unavailableDetail: '' };
	}

	const enroll = await api.post<Enrollment>('/me/two-factor/enroll', cookies);
	const secret = enroll.data?.secret ?? '';
	const qr = enroll.data?.otpauthUri
		? await qrToString(enroll.data.otpauthUri, {
				type: 'svg',
				margin: 1,
				width: 150,
				color: { dark: '#0a0f11', light: '#ffffff' }
			})
		: '';

	return {
		...base,
		secret,
		groupedSecret: groupSecret(secret),
		qr,
		unavailable: !enroll.ok,
		unavailableDetail: enroll.problem?.detail ?? ''
	};
};

export const actions: Actions = {
	changePassword: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(changePasswordSchema), { id: 'password' });
		if (!form.valid) return fail(400, { form });
		const res = await api.put('/me/password', cookies, {
			body: { currentPassword: form.data.currentPassword, newPassword: form.data.newPassword }
		});
		if (!res.ok) return message(form, 'wrong', { status: 400 });
		return message(form, 'changed');
	},
	verify: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(totpSchema), { id: 'verify' });
		if (!form.valid) return fail(400, { form });
		const res = await api.post<RecoveryCodes>('/me/two-factor/verify', cookies, {
			body: { code: form.data.code }
		});
		if (!res.ok) return message(form, 'wrong', { status: 400 });
		return { form, recoveryCodes: res.data?.codes ?? [] };
	},
	regenerate: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(totpSchema), { id: 'regenerate' });
		if (!form.valid) return fail(400, { form });
		const res = await api.post<RecoveryCodes>('/me/two-factor/recovery-codes', cookies, {
			body: { code: form.data.code }
		});
		if (!res.ok) return message(form, 'wrong', { status: 400 });
		return { form, recoveryCodes: res.data?.codes ?? [] };
	},
	disable: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(totpSchema), { id: 'disable' });
		if (!form.valid) return fail(400, { form });
		const res = await api.post('/me/two-factor/disable', cookies, { body: { code: form.data.code } });
		if (!res.ok) return message(form, 'wrong', { status: 400 });
		redirect(303, '/account/security');
	}
};
