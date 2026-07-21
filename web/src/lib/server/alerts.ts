import type {
	Alert,
	AlertStatus,
	Heartbeat,
	IngestionFailure,
	Silence
} from '$lib/alerts';
import { scenario } from './fixtures';

const MINUTE = 60_000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

const iso = (offsetMs: number) => new Date(Date.now() + offsetMs).toISOString();
const id = (prefix: string) => prefix + Math.random().toString(36).slice(2, 8);

function seed() {
	const alerts: Alert[] = [
		{
			id: 'al-6',
			severity: 'critical',
			title: 'Synthetic checkout failing from us-east-1',
			description:
				'Synthetic checkout flow from us-east-1 failed 27 consecutive runs. Payment authorization step times out after 10 s. Real-user checkout error rate is climbing in the same region.',
			source: 'Checkly',
			service: 'payments-api',
			status: 'open',
			ackedBy: null,
			labels: ['env:prod', 'region:us-east-1'],
			count: 27,
			firstSeenAt: iso(-38 * MINUTE),
			lastSeenAt: iso(-2 * MINUTE),
			children: [],
			links: [
				{ kind: 'runbook', label: 'Runbook: checkout failures' },
				{ kind: 'dashboard', label: 'Dashboard: payments-api' },
				{ kind: 'source', label: 'Source rule in Checkly' }
			],
			payload: `{
  "check": "checkout-flow-us-east-1",
  "state": "FAILED",
  "consecutive_failures": 27,
  "step": "payment_authorization",
  "error": "timeout after 10000 ms",
  "region": "us-east-1",
  "tags": ["env:prod", "team:payments"],
  "result_url": "https://app.checklyhq.com/checks/8f31"
}`,
			timeline: [
				{
					id: 'e1',
					at: iso(-38 * MINUTE),
					kind: 'received',
					text: 'Alert received from Checkly: matched routing rule payments-prod'
				},
				{
					id: 'e2',
					at: iso(-38 * MINUTE),
					kind: 'escalation',
					text: 'Escalation policy payments-primary started at step 1'
				},
				{
					id: 'e3',
					at: iso(-38 * MINUTE),
					kind: 'push',
					text: 'Push to Maya Chen',
					result: 'delivered',
					tone: 'success'
				},
				{
					id: 'e4',
					at: iso(-33 * MINUTE),
					kind: 'timeout',
					text: 'No ack after 5 min: moved to step 2',
					result: 'timeout',
					tone: 'warning'
				},
				{
					id: 'e5',
					at: iso(-33 * MINUTE),
					kind: 'sms',
					text: 'SMS to Maya Chen',
					result: 'delivered',
					tone: 'success'
				},
				{
					id: 'e6',
					at: iso(-33 * MINUTE),
					kind: 'push',
					text: 'Push to Priya Nair (step 2)',
					result: 'delivered',
					tone: 'success'
				}
			]
		},
		{
			id: 'grp-1',
			severity: 'high',
			title: 'payments-api p99 above 800 ms',
			description:
				'Latency on payments-api is above its objective across three pods. Grouped by service and label.',
			source: 'Datadog',
			service: 'payments-api',
			status: 'open',
			ackedBy: null,
			labels: ['env:prod', 'region:eu-west-1'],
			count: 12,
			firstSeenAt: iso(-52 * MINUTE),
			lastSeenAt: iso(-4 * MINUTE),
			children: [
				{ id: 'grp-1a', title: 'p99 812 ms · pod payments-7c9f', lastSeenAt: iso(-4 * MINUTE) },
				{ id: 'grp-1b', title: 'p99 977 ms · pod payments-b02d', lastSeenAt: iso(-8 * MINUTE) },
				{ id: 'grp-1c', title: 'p99 1.2 s · pod payments-51ae', lastSeenAt: iso(-15 * MINUTE) }
			],
			links: [{ kind: 'dashboard', label: 'Dashboard: payments-api' }],
			payload: `{
  "metric": "trace.http.request.duration.by.service.99p",
  "service": "payments-api",
  "threshold_ms": 800,
  "value_ms": 977,
  "tags": ["env:prod", "region:eu-west-1"]
}`,
			timeline: [
				{
					id: 'g1',
					at: iso(-52 * MINUTE),
					kind: 'received',
					text: 'Alert received from Datadog: grouped with 2 related alerts'
				},
				{
					id: 'g2',
					at: iso(-52 * MINUTE),
					kind: 'escalation',
					text: 'Escalation policy payments-primary started at step 1'
				},
				{
					id: 'g3',
					at: iso(-52 * MINUTE),
					kind: 'push',
					text: 'Push to Maya Chen',
					result: 'delivered',
					tone: 'success'
				}
			]
		},
		{
			id: 'grp-2',
			severity: 'high',
			title: '5xx burst on gateway',
			description: 'Elevated 502 rate on the gateway across two regions.',
			source: 'Grafana',
			service: 'gateway',
			status: 'open',
			ackedBy: null,
			labels: ['env:prod'],
			count: 4,
			firstSeenAt: iso(-70 * MINUTE),
			lastSeenAt: iso(-21 * MINUTE),
			children: [
				{ id: 'grp-2a', title: '502 rate 3.1% · eu-west-1', lastSeenAt: iso(-21 * MINUTE) },
				{ id: 'grp-2b', title: '502 rate 1.9% · us-east-1', lastSeenAt: iso(-30 * MINUTE) }
			],
			links: [{ kind: 'dashboard', label: 'Dashboard: gateway' }],
			payload: `{ "rule": "gateway-5xx", "rate": 0.031, "region": "eu-west-1" }`,
			timeline: [
				{
					id: 'h1',
					at: iso(-70 * MINUTE),
					kind: 'received',
					text: 'Alert received from Grafana: matched routing rule platform-prod'
				},
				{
					id: 'h2',
					at: iso(-70 * MINUTE),
					kind: 'escalation',
					text: 'Escalation policy platform-default started at step 1'
				},
				{
					id: 'h3',
					at: iso(-70 * MINUTE),
					kind: 'push',
					text: 'Push to Priya Nair',
					result: 'delivered',
					tone: 'success'
				},
				{
					id: 'h4',
					at: iso(-65 * MINUTE),
					kind: 'chat',
					text: 'Posted to #platform-alerts',
					result: 'delivered',
					tone: 'success'
				}
			]
		},
		{
			id: 'al-2',
			severity: 'warning',
			title: 'Disk usage 85% on db-3',
			description: 'Disk on db-3 crossed the 85% warning threshold.',
			source: 'Prometheus',
			service: 'database',
			status: 'open',
			ackedBy: null,
			labels: ['env:prod'],
			count: 1,
			firstSeenAt: iso(-11 * MINUTE),
			lastSeenAt: iso(-11 * MINUTE),
			children: [],
			links: [{ kind: 'runbook', label: 'Runbook: disk pressure' }],
			payload: `{ "alertname": "DiskUsageHigh", "instance": "db-3", "value": 0.85 }`,
			timeline: [
				{
					id: 'i1',
					at: iso(-11 * MINUTE),
					kind: 'received',
					text: 'Alert received from Prometheus: matched routing rule platform-prod'
				},
				{
					id: 'i2',
					at: iso(-11 * MINUTE),
					kind: 'chat',
					text: 'Posted to #platform-alerts: warning severity does not page',
					result: 'delivered',
					tone: 'success'
				}
			]
		},
		{
			id: 'al-3',
			severity: 'warning',
			title: 'TLS cert for status.acme.dev expires in 7 days',
			description: 'The certificate for status.acme.dev expires on 2026-07-18.',
			source: 'cert-monitor',
			service: 'edge',
			status: 'acked',
			ackedBy: 'Maya Chen',
			labels: ['env:prod'],
			count: 1,
			firstSeenAt: iso(-60 * MINUTE),
			lastSeenAt: iso(-60 * MINUTE),
			children: [],
			links: [{ kind: 'runbook', label: 'Runbook: certificate rotation' }],
			payload: `{ "domain": "status.acme.dev", "expires_in_days": 7 }`,
			timeline: [
				{
					id: 'j1',
					at: iso(-60 * MINUTE),
					kind: 'received',
					text: 'Alert received from cert-monitor'
				},
				{
					id: 'j2',
					at: iso(-58 * MINUTE),
					kind: 'acked',
					text: 'Acknowledged by Maya Chen via Slack',
					result: 'acked',
					tone: 'brand'
				}
			]
		},
		{
			id: 'al-5',
			severity: 'high',
			title: 'Queue depth above 10k',
			description: 'The events-worker queue backed up past 10,000 messages.',
			source: 'Prometheus',
			service: 'events-worker',
			status: 'resolved',
			ackedBy: 'Dev Patel',
			labels: ['env:prod'],
			count: 3,
			firstSeenAt: iso(-3 * HOUR),
			lastSeenAt: iso(-2 * HOUR - 40 * MINUTE),
			children: [],
			links: [],
			payload: `{ "alertname": "QueueDepthHigh", "queue": "events", "value": 10432 }`,
			timeline: [
				{ id: 'k1', at: iso(-3 * HOUR), kind: 'received', text: 'Alert received from Prometheus' },
				{
					id: 'k2',
					at: iso(-2 * HOUR - 40 * MINUTE),
					kind: 'resolved',
					text: 'Source reported recovery: auto-resolved',
					result: 'resolved',
					tone: 'success'
				}
			]
		}
	];

	const silences: Silence[] = [
		{
			id: 's-1',
			state: 'active',
			scope: ['source = Datadog', 'service = payments-api'],
			creator: 'Maya Chen',
			reason: 'Planned failover test',
			startsAt: iso(-50 * MINUTE),
			endsAt: iso(2 * HOUR + 10 * MINUTE)
		},
		{
			id: 's-2',
			state: 'scheduled',
			scope: ['service = database', 'label env = staging'],
			creator: 'Dev Patel',
			reason: 'DB maintenance window',
			startsAt: iso(DAY + 12 * HOUR),
			endsAt: iso(DAY + 16 * HOUR)
		}
	];

	const history: Silence[] = [
		{
			id: 'h-1',
			state: 'ended',
			scope: ['source = Grafana'],
			creator: 'Priya Nair',
			reason: 'Grafana upgrade',
			startsAt: iso(-3 * DAY),
			endsAt: iso(-3 * DAY + 2 * HOUR)
		},
		{
			id: 'h-2',
			state: 'ended',
			scope: ['service = edge', 'label region = eu-west-1'],
			creator: 'Maya Chen',
			reason: 'CDN migration',
			startsAt: iso(-8 * DAY),
			endsAt: iso(-8 * DAY + 2 * HOUR)
		},
		{
			id: 'h-3',
			state: 'ended',
			scope: ['source = cert-monitor'],
			creator: 'Dev Patel',
			reason: 'Cert rotation',
			startsAt: iso(-12 * DAY),
			endsAt: iso(-12 * DAY + HOUR)
		}
	];

	const failures: IngestionFailure[] = [
		{
			id: 'f-1',
			source: 'webhook/legacy-nagios',
			at: iso(-2 * HOUR),
			reason: 'JSON parse error at position 214: unexpected token',
			payload: '{"host":"db-3","service":"disk","state":"WARNING","output":"DISK WARNING - free space / 12%",,}'
		},
		{
			id: 'f-2',
			source: 'email/alerts@acme.dev',
			at: iso(-11 * HOUR),
			reason: 'No severity mapping matched subject prefix "URGENT"',
			payload:
				'Subject: URGENT disk space on db-3\nFrom: nagios@legacy.acme.dev\n\nDISK WARNING - free space / 12%'
		},
		{
			id: 'f-3',
			source: 'webhook/custom-cron',
			at: iso(-20 * HOUR),
			reason: 'Missing required field: title',
			payload: '{"severity":"warning","details":"job runtime exceeded budget","job":"nightly-rollup"}'
		}
	];

	const heartbeats: Heartbeat[] = [
		{
			id: 'hb-1',
			name: 'nightly-db-backup',
			state: 'healthy',
			interval: '24 h',
			grace: '30 m',
			lastSeenAt: iso(-7 * HOUR),
			policy: 'platform-default'
		},
		{
			id: 'hb-2',
			name: 'billing-cron',
			state: 'missed',
			interval: '1 h',
			grace: '10 m',
			lastSeenAt: iso(-2 * HOUR - 14 * MINUTE),
			policy: 'payments-primary'
		},
		{
			id: 'hb-3',
			name: 'edge-sync',
			state: 'healthy',
			interval: '5 m',
			grace: '2 m',
			lastSeenAt: iso(-42_000),
			policy: 'platform-default'
		}
	];

	return { alerts, silences, history, failures, heartbeats };
}

const store = seed();
const empty = !['active', 'degraded'].includes(scenario());

function record(alert: Alert, kind: Alert['timeline'][number]['kind'], text: string, extra?: Partial<Alert['timeline'][number]>) {
	alert.timeline.push({ id: id('e'), at: new Date().toISOString(), kind, text, ...extra });
}

export function listAlerts(): Alert[] {
	return empty ? [] : store.alerts;
}

export function getAlert(alertId: string): Alert | undefined {
	return empty ? undefined : store.alerts.find((alert) => alert.id === alertId);
}

export function setStatus(alertId: string, status: AlertStatus, actor = 'Maya Chen') {
	const alert = getAlert(alertId);
	if (!alert) return;

	alert.status = status;

	if (status === 'acked') {
		alert.ackedBy = actor;
		record(alert, 'acked', `Acknowledged by ${actor}`, { result: 'acked', tone: 'brand' });
	} else if (status === 'resolved') {
		record(alert, 'resolved', `Resolved by ${actor}`, { result: 'resolved', tone: 'success' });
	}
}

export function escalate(alertId: string, actor = 'Maya Chen') {
	const alert = getAlert(alertId);
	if (!alert) return;
	record(alert, 'escalation', `Manually escalated to the next step by ${actor}`, {
		result: 'escalated',
		tone: 'warning'
	});
}

export function attachToIncident(alertId: string, incidentId: string) {
	const alert = getAlert(alertId);
	if (!alert) return;
	record(alert, 'chat', `Attached to ${incidentId}`, { result: 'attached', tone: 'brand' });
}

export function listSilences(): Silence[] {
	return empty ? [] : store.silences;
}

export function listSilenceHistory(): Silence[] {
	return empty ? [] : store.history;
}

export function createSilence(input: {
	scope: string[];
	reason: string;
	startsNow: boolean;
	startsAt?: string;
	durationHours: number;
}): void {
	const start = input.startsNow ? Date.now() : Date.parse(input.startsAt ?? '');
	if (Number.isNaN(start)) return;

	store.silences.unshift({
		id: id('s-'),
		state: input.startsNow ? 'active' : 'scheduled',
		scope: input.scope,
		reason: input.reason || 'No reason given',
		creator: 'Maya Chen',
		startsAt: new Date(start).toISOString(),
		endsAt: new Date(start + input.durationHours * HOUR).toISOString()
	});
}

export function endSilence(silenceId: string) {
	const index = store.silences.findIndex((silence) => silence.id === silenceId);
	if (index === -1) return;

	const [silence] = store.silences.splice(index, 1);
	store.history.unshift({ ...silence, state: 'ended', endsAt: new Date().toISOString() });
}

export function listFailures(): IngestionFailure[] {
	return empty ? [] : store.failures;
}

export function listHeartbeats(): Heartbeat[] {
	return empty ? [] : store.heartbeats;
}

const spaced = (duration: string) => duration.replace(/^(\d+)/, '$1 ');

export function createHeartbeat(input: {
	name: string;
	interval: string;
	grace: string;
	policy: string;
}): { slug: string; token: string } {
	const slug = input.name
		.trim()
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-|-$/g, '');

	store.heartbeats.push({
		id: id('hb-'),
		name: input.name.trim(),
		state: 'healthy',
		interval: spaced(input.interval),
		grace: spaced(input.grace),
		lastSeenAt: null,
		policy: input.policy
	});

	return { slug, token: Math.random().toString(16).slice(2, 8) };
}
