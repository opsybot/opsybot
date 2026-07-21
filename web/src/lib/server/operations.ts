import type { Backup, Channel, Integration, OpsLicense, OpsUpdate, Overall, Queue, Subsystem } from '$lib/operations';
import { scenario } from './fixtures';

const LAST_CHECK = '2026-07-12 09:41 UTC';

type OpsStore = {
	subsystems: Subsystem[];
	queues: Queue[];
	channels: Channel[];
	integrations: Integration[];
	license: OpsLicense;
	update: OpsUpdate;
	backup: Backup;
};

function seed(): OpsStore {
	const subsystems: Subsystem[] = [
		{ id: 'api', label: 'API server', icon: 'server', state: 'ok', detail: '3/3 replicas · p99 42 ms' },
		{ id: 'db', label: 'Database', icon: 'database', state: 'ok', detail: 'primary + 2 replicas · 41 GB · lag 0.2 s' },
		{ id: 'storage', label: 'Object storage', icon: 'hard-drive', state: 'ok', detail: 'attachments · 18 GB used' },
		{ id: 'workers', label: 'Delivery workers', icon: 'workflow', state: 'ok', detail: '3/3 healthy · all workers up' },
		{ id: 'scheduler', label: 'Scheduler', icon: 'calendar-clock', state: 'ok', detail: 'on-call rotations · next tick 12 s' }
	];
	const queues: Queue[] = [
		{ name: 'notifications', depth: 4, rate: '120/min', state: 'ok' },
		{ name: 'alert-ingest', depth: 0, rate: '38/min', state: 'ok' },
		{ name: 'webhooks-out', depth: 12, rate: '15/min', state: 'ok' },
		{ name: 'postmortem-ai', depth: 1, rate: '2/min', state: 'ok' }
	];
	const channels: Channel[] = [
		{ ch: 'Push', last: '12 s ago', state: 'ok' },
		{ ch: 'SMS (bridge)', last: '3 m ago', state: 'ok' },
		{ ch: 'Voice (bridge)', last: '1 h ago', state: 'ok' },
		{ ch: 'Email', last: '44 s ago', state: 'ok' },
		{ ch: 'Slack', last: '8 s ago', state: 'ok' }
	];
	const integrations: Integration[] = [
		{ name: 'prometheus-prod', state: 'ok', detail: 'last event 2 m ago' },
		{ name: 'grafana-main', state: 'ok', detail: 'last event 18 m ago' },
		{ name: 'legacy-nagios', state: 'ok', detail: 'last event 6 m ago' },
		{ name: 'sso.acme.dev (OIDC)', state: 'ok', detail: 'handshake 0.3 s' }
	];
	const license: OpsLicense = { title: 'License active', detail: 'Business · expires 2027-01-14', tone: 'success' };
	const update: OpsUpdate = { current: 'v3.4.1', latest: 'v3.5.0', released: '2 days ago' };
	const backup: Backup = { ago: '7 h ago', at: '2026-07-12 02:00 UTC', size: '41 GB', dest: 's3://acme-opsybot-backups', schedule: 'nightly' };
	return { subsystems, queues, channels, integrations, license, update, backup };
}

const store = seed();
const state = scenario();

if (state === 'degraded') {
	const workers = store.subsystems.find((s) => s.id === 'workers');
	if (workers) {
		workers.state = 'warn';
		workers.detail = '2/3 healthy · worker-3 restarting';
	}
	const webhooks = store.queues.find((q) => q.name === 'webhooks-out');
	if (webhooks) {
		webhooks.state = 'warn';
		webhooks.depth = 112;
	}
	const nagios = store.integrations.find((i) => i.name === 'legacy-nagios');
	if (nagios) {
		nagios.state = 'warn';
		nagios.detail = '3 parse failures / 24 h';
	}
}

if (state === 'empty') {
	store.queues = store.queues.map((q) => ({ ...q, depth: 0, rate: '0/min', state: 'ok' }));
	store.channels = [];
	store.integrations = [];
	store.backup = null;
	store.license = { title: 'Running the AGPL core', detail: 'Community · unrestricted paging', tone: 'neutral' };
	store.update = { current: 'v3.5.0', latest: null, released: null };
}

function overall(): Overall {
	const degraded =
		store.subsystems.some((s) => s.state !== 'ok') ||
		store.queues.some((q) => q.state !== 'ok') ||
		store.integrations.some((i) => i.state !== 'ok');
	return degraded
		? {
				degraded: true,
				title: 'Degraded: 1 worker recovering',
				detail: `Paging is unaffected: a healthy worker is covering. Last check ${LAST_CHECK}.`
			}
		: { degraded: false, title: 'All systems healthy', detail: `Last check ${LAST_CHECK}.` };
}

export function getDiagnostics() {
	return {
		overall: overall(),
		subsystems: store.subsystems,
		queues: store.queues,
		channels: store.channels,
		integrations: store.integrations,
		license: store.license,
		update: store.update
	};
}

export function getBackup() {
	return { backup: store.backup };
}
