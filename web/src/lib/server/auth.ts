export type TokenState = 'valid' | 'expired';

export function tokenState(token: string | null): TokenState {
	return token === 'expired' ? 'expired' : 'valid';
}
