import { fail, redirect } from '@sveltejs/kit';
import { toString as qrToString } from 'qrcode';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { changePasswordSchema, totpSchema } from '$lib/schemas/auth';
import { apiClient } from '$lib/server/api';
import type { Actions, PageServerLoad } from './$types';

function groupSecret(secret: string): string {
	return secret.replace(/(.{4})/g, '$1 ').trim();
}

export const load: PageServerLoad = async ({ cookies }) => {
	const client = apiClient(cookies);
	const { data: me } = await client.GET('/me');
	const enabled = me?.twoFactorEnabled ?? false;

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

	const { data: enroll, error: enrollError } = await client.POST('/me/two-factor/enroll');
	const secret = enroll?.secret ?? '';
	const qr = enroll?.otpauthUri
		? await qrToString(enroll.otpauthUri, {
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
		unavailable: !!enrollError,
		unavailableDetail: enrollError?.detail ?? ''
	};
};

export const actions: Actions = {
	changePassword: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(changePasswordSchema), { id: 'password' });
		if (!form.valid) return fail(400, { form });
		const { error } = await apiClient(cookies).PUT('/me/password', {
			body: { currentPassword: form.data.currentPassword, newPassword: form.data.newPassword }
		});
		if (error) return message(form, 'wrong', { status: 400 });
		return message(form, 'changed');
	},
	verify: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(totpSchema), { id: 'verify' });
		if (!form.valid) return fail(400, { form });
		const { data, error } = await apiClient(cookies).POST('/me/two-factor/verify', {
			body: { code: form.data.code }
		});
		if (error) return message(form, 'wrong', { status: 400 });
		return { form, recoveryCodes: data?.codes ?? [] };
	},
	regenerate: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(totpSchema), { id: 'regenerate' });
		if (!form.valid) return fail(400, { form });
		const { data, error } = await apiClient(cookies).POST('/me/two-factor/recovery-codes', {
			body: { code: form.data.code }
		});
		if (error) return message(form, 'wrong', { status: 400 });
		return { form, recoveryCodes: data?.codes ?? [] };
	},
	disable: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(totpSchema), { id: 'disable' });
		if (!form.valid) return fail(400, { form });
		const { error } = await apiClient(cookies).POST('/me/two-factor/disable', {
			body: { code: form.data.code }
		});
		if (error) return message(form, 'wrong', { status: 400 });
		redirect(303, '/account/security');
	}
};
