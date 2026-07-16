export type LoginOutcome = 'ok' | 'invalid' | 'deactivated' | 'sso-required';

const ACCOUNTS: Record<string, LoginOutcome> = {
	'maya@acme.dev': 'ok',
	'jordan@acme.dev': 'deactivated',
	'sam@sso.acme.dev': 'sso-required'
};

export function attemptLogin(email: string): LoginOutcome {
	return ACCOUNTS[email.trim().toLowerCase()] ?? 'invalid';
}

export function verifyTotp(code: string): boolean {
	return /^\d{6}$/.test(code) && code !== '000000';
}

export function verifyRecoveryCode(code: string): boolean {
	return /^[a-z0-9]{4}-[a-z0-9]{4}$/.test(code.trim().toLowerCase());
}

export type TokenState = 'valid' | 'expired';

export function tokenState(token: string | null): TokenState {
	return token === 'expired' ? 'expired' : 'valid';
}

export type Invite = {
	invitedBy: string;
	workspace: string;
	email: string;
	sentAt: string;
};

export function getInvite(): Invite {
	return {
		invitedBy: 'Maya Chen',
		workspace: 'Acme Corp',
		email: 'jordan@acme.dev',
		sentAt: '2026-06-28T09:14:00Z'
	};
}

export const RECOVERY_CODES = [
	'k7f2-9mqa',
	'x3nd-04rt',
	'b8vw-2lch',
	'p5js-7yke',
	'm1qz-6dnu',
	'r9th-3wgo',
	'c4lx-8bfs',
	'v6ka-1pjm'
];

export const TOTP_SECRET = 'JBSWY3DPEHPK3PXPQK7F2M9A';
