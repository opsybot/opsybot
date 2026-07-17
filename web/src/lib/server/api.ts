import { env } from '$env/dynamic/private';
import type { Cookies } from '@sveltejs/kit';

const BASE = env.OPSYBOT_API_URL ?? 'http://127.0.0.1:8099';

export const SESSION_COOKIE = 'opsybot_session';
export const PENDING_COOKIE = 'opsybot_2fa';

export type ApiProblem = { type?: string; title?: string; status?: number; detail?: string };

export type ApiResult<T> = {
	status: number;
	ok: boolean;
	data: T | null;
	problem: ApiProblem | null;
	setCookie: string | null;
};

type Init = { body?: unknown; session?: string; pending?: string };

async function call<T>(method: string, path: string, cookies: Cookies, init: Init = {}): Promise<ApiResult<T>> {
	const headers: Record<string, string> = {};

	const cookieParts: string[] = [];
	const session = init.session ?? cookies.get(SESSION_COOKIE);
	if (session) cookieParts.push(`${SESSION_COOKIE}=${session}`);
	const pending = init.pending ?? cookies.get(PENDING_COOKIE);
	if (pending) cookieParts.push(`${PENDING_COOKIE}=${pending}`);
	if (cookieParts.length) headers['cookie'] = cookieParts.join('; ');

	let body: string | undefined;
	if (init.body !== undefined) {
		headers['content-type'] = 'application/json';
		body = JSON.stringify(init.body);
	}

	const res = await fetch(`${BASE}/v1${path}`, { method, headers, body });
	const setCookie = res.headers.get('set-cookie');

	let data: T | null = null;
	let problem: ApiProblem | null = null;
	const text = await res.text();
	if (text) {
		try {
			const parsed = JSON.parse(text);
			if (res.ok) data = parsed as T;
			else problem = parsed as ApiProblem;
		} catch {
			/* non-JSON response */
		}
	}

	return { status: res.status, ok: res.ok, data, problem, setCookie };
}

export const api = {
	get: <T>(path: string, cookies: Cookies, init?: Init) => call<T>('GET', path, cookies, init),
	post: <T>(path: string, cookies: Cookies, init?: Init) => call<T>('POST', path, cookies, init),
	put: <T>(path: string, cookies: Cookies, init?: Init) => call<T>('PUT', path, cookies, init),
	patch: <T>(path: string, cookies: Cookies, init?: Init) => call<T>('PATCH', path, cookies, init),
	del: <T>(path: string, cookies: Cookies, init?: Init) => call<T>('DELETE', path, cookies, init)
};

export function cookieValue(setCookie: string | null, name: string): string | null {
	if (!setCookie) return null;
	const match = new RegExp(`${name}=([^;]+)`).exec(setCookie);
	return match ? match[1] : null;
}
