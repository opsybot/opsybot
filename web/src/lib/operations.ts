import type { Tone } from '$lib/dashboard';

export type HealthState = 'ok' | 'warn' | 'down';

export type Subsystem = { id: string; label: string; icon: string; state: HealthState; detail: string };
export type Queue = { name: string; depth: number; rate: string; state: HealthState };
export type Channel = { ch: string; last: string; state: HealthState };
export type Integration = { name: string; state: HealthState; detail: string };

export type OpsLicense = { title: string; detail: string; tone: 'success' | 'neutral' };
export type OpsUpdate = { current: string; latest: string | null; released: string | null };
export type Backup = { ago: string; at: string; size: string; dest: string; schedule: string } | null;

export type Overall = { degraded: boolean; title: string; detail: string };

export function opsTone(state: HealthState): Tone {
	return state === 'ok' ? 'success' : state === 'warn' ? 'warning' : 'critical';
}

export function dotColor(state: HealthState): string {
	return `var(--${opsTone(state)})`;
}

export function depthClass(state: HealthState): string {
	return state === 'warn'
		? 'text-warning-ink'
		: state === 'down'
			? 'text-critical-ink'
			: 'text-muted-foreground';
}
