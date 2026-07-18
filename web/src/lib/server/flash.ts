import type { Cookies } from '@sveltejs/kit';
import { FLASH_COOKIE, type Flash } from '$lib/flash';

export function setFlash(cookies: Cookies, flash: Flash): void {
	cookies.set(FLASH_COOKIE, JSON.stringify(flash), {
		path: '/',
		httpOnly: false,
		sameSite: 'lax',
		maxAge: 30
	});
}
