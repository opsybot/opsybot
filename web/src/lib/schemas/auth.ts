import { z } from 'zod';

const password = z.string().min(12, 'Use at least 12 characters.');
const timezone = z.string().min(1, 'Pick a timezone.');

export const loginSchema = z.object({
	email: z.email('Enter a valid email address.'),
	password: z.string().min(1, 'Enter your password.'),
	remember: z.boolean().default(true)
});

export const signupAccountSchema = z.object({
	name: z.string().min(1, 'Enter your name.'),
	email: z.email('Enter a valid work email address.'),
	password
});

export const workspaceSchema = z.object({
	workspace: z.string().min(1, 'Name the workspace.'),
	timezone
});

export const setupSchema = z.object({
	name: z.string().min(1, 'Enter your name.'),
	email: z.email('Enter a valid email address.'),
	password,
	workspace: z.string().min(1, 'Name the workspace.'),
	timezone
});

export const inviteSchema = z.object({
	name: z.string().min(1, 'Enter your name.'),
	password,
	timezone
});

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
export type SignupAccountSchema = typeof signupAccountSchema;
export type WorkspaceSchema = typeof workspaceSchema;
export type SetupSchema = typeof setupSchema;
export type InviteSchema = typeof inviteSchema;
export type ForgotPasswordSchema = typeof forgotPasswordSchema;
export type ResetPasswordSchema = typeof resetPasswordSchema;
export type TotpSchema = typeof totpSchema;
export type RecoveryCodeSchema = typeof recoveryCodeSchema;
