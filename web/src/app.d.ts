import type { Theme } from '$lib/theme';

declare global {
	namespace App {
		interface Locals {
			theme: Theme;
		}
		interface PageData {
			theme: Theme;
		}
	}
}

export {};
