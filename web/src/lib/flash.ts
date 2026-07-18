export const FLASH_COOKIE = 'opsybot_flash';

export type Flash = { tone: 'success' | 'info' | 'error'; title: string; message?: string };

export function takeFlashCookie(): Flash | null {
	if (typeof document === 'undefined') return null;
	const match = document.cookie.match(/(?:^|;\s*)opsybot_flash=([^;]+)/);
	if (!match) return null;
	document.cookie = `${FLASH_COOKIE}=; Max-Age=0; Path=/`;
	try {
		return JSON.parse(decodeURIComponent(match[1])) as Flash;
	} catch {
		return null;
	}
}
