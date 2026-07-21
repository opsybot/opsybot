import type { Tone } from '$lib/dashboard';

export type AlertSeverity = 'critical' | 'high' | 'warning';
export type AlertStatus = 'open' | 'acked' | 'resolved';

export const SEVERITY_TONE: Record<AlertSeverity, Tone> = {
	critical: 'critical',
	high: 'high',
	warning: 'warning'
};

export const SEVERITY_SHORT: Record<AlertSeverity, string> = {
	critical: 'CRIT',
	high: 'HIGH',
	warning: 'WARN'
};

export type AlertChild = {
	id: string;
	title: string;
	lastSeenAt: string;
};

export type Alert = {
	id: string;
	severity: AlertSeverity;
	title: string;
	description: string;
	source: string;
	service: string;
	status: AlertStatus;
	ackedBy: string | null;
	labels: string[];
	count: number;
	firstSeenAt: string;
	lastSeenAt: string;
	children: AlertChild[];
	links: { kind: 'runbook' | 'dashboard' | 'source'; label: string }[];
	payload: string;
	timeline: EscalationEvent[];
};

export type EscalationEventKind =
	| 'received'
	| 'escalation'
	| 'push'
	| 'sms'
	| 'timeout'
	| 'chat'
	| 'acked'
	| 'resolved';

export type EscalationEvent = {
	id: string;
	at: string;
	kind: EscalationEventKind;
	text: string;
	result?: string;
	tone?: 'success' | 'warning' | 'brand';
};

export type SilenceState = 'active' | 'scheduled' | 'ended';

export type Silence = {
	id: string;
	state: SilenceState;
	scope: string[];
	reason: string;
	creator: string;
	startsAt: string;
	endsAt: string;
};

export type IngestionFailure = {
	id: string;
	source: string;
	at: string;
	reason: string;
	payload: string;
};

export type HeartbeatState = 'healthy' | 'missed';

export type Heartbeat = {
	id: string;
	name: string;
	state: HeartbeatState;
	interval: string;
	grace: string;
	lastSeenAt: string | null;
	policy: string;
};

export const SEVERITIES: AlertSeverity[] = ['critical', 'high', 'warning'];

export const STATUSES: { id: AlertStatus; label: string }[] = [
	{ id: 'open', label: 'Open' },
	{ id: 'acked', label: 'Acked' },
	{ id: 'resolved', label: 'Resolved' }
];

export const ESCALATION_POLICIES = ['platform-default', 'payments-primary', 'frontend-daytime'];

export const SCOPE_FIELDS = ['source', 'service', 'label'];

export const INTERVALS = [
	{ value: '1m', label: 'Every minute' },
	{ value: '5m', label: 'Every 5 minutes' },
	{ value: '1h', label: 'Every hour' },
	{ value: '24h', label: 'Every 24 hours' }
];

export const GRACE_PERIODS = [
	{ value: '2m', label: '2 minutes' },
	{ value: '10m', label: '10 minutes' },
	{ value: '30m', label: '30 minutes' }
];

export const DURATIONS = [
	{ value: '1h', label: '1 hour' },
	{ value: '4h', label: '4 hours' },
	{ value: '8h', label: '8 hours' },
	{ value: '24h', label: '24 hours' }
];
