import { z } from 'zod';

const password = z.string().min(12, 'Use at least 12 characters.');
const timezone = z.string().min(1, 'Pick a timezone.');

export const loginSchema = z.object({
	email: z.email('Enter a valid email address.'),
	password: z.string().min(1, 'Enter your password.'),
	remember: z.boolean().default(true)
});

const confirmMatches = (fields: { password: string; confirm: string }) =>
	fields.password === fields.confirm;
const confirmMessage = { message: 'Both passwords must match.', path: ['confirm'] };

const slug = z
	.string()
	.min(1, 'Choose a workspace URL.')
	.regex(/^[a-z][a-z0-9-]{0,39}$/, 'Lowercase letters, numbers, and hyphens; start with a letter.');

export const signupSchema = z
	.object({
		name: z.string().min(1, 'Enter your name.'),
		email: z.email('Enter a valid work email address.'),
		password,
		confirm: z.string().min(1, 'Repeat the password.'),
		workspace: z.string().min(1, 'Name the workspace.'),
		slug,
		timezone
	})
	.refine(confirmMatches, confirmMessage);

export const setupSchema = z
	.object({
		name: z.string().min(1, 'Enter your name.'),
		email: z.email('Enter a valid email address.'),
		password,
		confirm: z.string().min(1, 'Repeat the password.'),
		workspace: z.string().min(1, 'Name the workspace.'),
		slug,
		timezone
	})
	.refine(confirmMatches, confirmMessage);

export const inviteSchema = z
	.object({
		name: z.string().min(1, 'Enter your name.'),
		password,
		confirm: z.string().min(1, 'Repeat the password.'),
		timezone
	})
	.refine(confirmMatches, confirmMessage);

export const forgotPasswordSchema = z.object({
	email: z.email('Enter the email you log in with.')
});

export const resetPasswordSchema = z
	.object({
		password,
		confirm: z.string().min(1, 'Repeat the new password.')
	})
	.refine((fields) => fields.password === fields.confirm, {
		message: 'Both passwords must match.',
		path: ['confirm']
	});

export const totpSchema = z.object({
	code: z.string().regex(/^\d{6}$/, 'Enter the 6-digit code from your app.')
});

export const recoveryCodeSchema = z.object({
	code: z
		.string()
		.trim()
		.toLowerCase()
		.regex(/^[a-z0-9]{4}-[a-z0-9]{4}$/, 'A recovery code is two groups of four, like k7f2-9mqa.')
});

export const profileSchema = z.object({
	name: z.string().min(1, 'Enter your name.'),
	timezone
});

export const channelSchema = z.object({
	type: z.enum(['slack', 'teams', 'discord', 'telegram', 'ntfy', 'email', 'webhook']),
	detail: z.string().trim().min(1, 'Enter the address or URL.').max(200, 'Keep it under 200 characters.')
});

export const changePasswordSchema = z
	.object({
		currentPassword: z.string().min(1, 'Enter your current password.'),
		newPassword: password,
		confirm: z.string().min(1, 'Repeat the new password.')
	})
	.refine((fields) => fields.newPassword === fields.confirm, {
		message: 'Both passwords must match.',
		path: ['confirm']
	});

export type LoginSchema = typeof loginSchema;
export type SignupSchema = typeof signupSchema;
export type SetupSchema = typeof setupSchema;
export type InviteSchema = typeof inviteSchema;
export type ForgotPasswordSchema = typeof forgotPasswordSchema;
export type ResetPasswordSchema = typeof resetPasswordSchema;
export type TotpSchema = typeof totpSchema;
export type RecoveryCodeSchema = typeof recoveryCodeSchema;
export type ChannelSchema = typeof channelSchema;
