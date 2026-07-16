import type { Status } from '$lib/components/status-badge.svelte';
import type { OnboardingStepId } from '$lib/onboarding';

export type Severity = 'SEV1' | 'SEV2' | 'SEV3' | 'SEV4';
export type Tone = 'critical' | 'high' | 'warning' | 'info' | 'success' | 'neutral';

export const SEVERITY_TONE: Record<Severity, Tone> = {
	SEV1: 'critical',
	SEV2: 'high',
	SEV3: 'warning',
	SEV4: 'info'
};

export type Incident = {
	id: string;
	title: string;
	severity: Severity;
	status: Status;
	lead: string;
	declaredAt: string;
};

export type Alert = {
	id: string;
	tone: Tone;
	title: string;
	source: string;
	firedAt: string;
};

export type OnCallEntry = {
	team: string;
	name: string;
	you: boolean;
	until: string;
};

export type Shift = {
	start: string;
	end: string;
	team: string;
};

export type OverdueKind = 'update' | 'follow-up' | 'postmortem';

export type OverdueItem = {
	id: string;
	kind: OverdueKind;
	tone: Tone;
	title: string;
	dueAt: string;
	action: string;
	href: string;
};

export type InstanceHealth = {
	selfHosted: boolean;
	workersHealthy: number;
	workersTotal: number;
	checkedAt: string;
};

export type Onboarding = {
	completed: OnboardingStepId[];
	dismissed: boolean;
};

export type Dashboard = {
	// Server clock so the first paint and the hydrated page agree on now
	now: number;
	onboarding: Onboarding | null;
	incidents: Incident[];
	alerts: Alert[];
	// Alerts per hour over the last 24h
	alertVolume: number[];
	onCallNow: OnCallEntry[];
	myShifts: Shift[];
	overdue: OverdueItem[];
	instance: InstanceHealth;
};
