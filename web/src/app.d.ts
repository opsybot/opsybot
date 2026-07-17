import type { Theme } from '$lib/theme';
import type { AuthUser } from '$lib/session';

declare global {
	namespace App {
		interface Locals {
			theme: Theme;
			user: AuthUser | null;
		}
		interface PageData {
			theme: Theme;
		}
	}
}

export {};
