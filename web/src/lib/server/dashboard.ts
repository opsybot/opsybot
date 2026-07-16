import type { Alert, Dashboard, Incident, OnCallEntry, OverdueItem, Shift } from '$lib/dashboard';
import { isActive } from '$lib/incidents';
import { scenario } from './fixtures';
import { listIncidents } from './incidents';

const MINUTE = 60_000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

// UTC shift window; an end hour at or before the start rolls into the next day
function shiftOn(day: Date, startHour: number, endHour: number) {
	const start = new Date(day);
	start.setUTCHours(startHour, 0, 0, 0);
	const end = new Date(day);
	end.setUTCHours(endHour, 0, 0, 0);
	if (endHour <= startHour) end.setUTCDate(end.getUTCDate() + 1);
	return { start: start.toISOString(), end: end.toISOString() };
}

function activeIncidents(): Incident[] {
	return listIncidents()
		.filter(isActive)
		.map((incident) => ({
			id: incident.id,
			title: incident.name,
			severity: incident.severity,
			status: incident.status,
			lead: incident.lead,
			declaredAt: incident.declaredAt
		}));
}

function alerts(now: number): Alert[] {
	return [
		{
			id: 'al-1',
			tone: 'high',
			title: 'payments-api p99 above 800 ms',
			source: 'Datadog',
			firedAt: new Date(now - 4 * MINUTE).toISOString()
		},
		{
			id: 'al-2',
			tone: 'warning',
			title: 'Disk usage 85% on db-3',
			source: 'Prometheus',
			firedAt: new Date(now - 12 * MINUTE).toISOString()
		},
		{
			id: 'al-3',
			tone: 'warning',
			title: 'TLS cert for status.acme.dev expires in 7 days',
			source: 'cert-monitor',
			firedAt: new Date(now - HOUR).toISOString()
		}
	];
}

function overdue(now: number): OverdueItem[] {
	return [
		{
			id: 'ov-1',
			kind: 'update',
			tone: 'critical',
			title: 'Status update due — INC-2481',
			dueAt: new Date(now - 6 * MINUTE).toISOString(),
			action: 'Post update',
			href: '/incidents'
		},
		{
			id: 'ov-2',
			kind: 'follow-up',
			tone: 'warning',
			title: 'Follow-up: add rate-limit alarm (INC-2472)',
			dueAt: new Date(now - 2 * DAY).toISOString(),
			action: 'Open',
			href: '/incidents'
		},
		{
			id: 'ov-3',
			kind: 'postmortem',
			tone: 'warning',
			title: 'Postmortem — INC-2468',
			dueAt: new Date(now - 3 * DAY).toISOString(),
			action: 'Start draft',
			href: '/postmortems'
		}
	];
}

function onCallNow(now: number, youAreOnCall: boolean): OnCallEntry[] {
	const today = new Date(now);
	return [
		{
			team: 'payments',
			name: youAreOnCall ? 'Maya Chen' : 'Sana Ito',
			you: youAreOnCall,
			until: shiftOn(today, 9, 18).end
		},
		{ team: 'platform', name: 'Priya Nair', you: false, until: shiftOn(today, 9, 20).end },
		{ team: 'frontend', name: 'Dev Patel', you: false, until: shiftOn(today, 21, 9).end }
	];
}

function myShifts(now: number): Shift[] {
	const today = new Date(now);
	return [
		{ ...shiftOn(today, 9, 18), team: 'payments' },
		{ ...shiftOn(new Date(now + DAY), 9, 18), team: 'payments' },
		{ ...shiftOn(new Date(now + 5 * DAY), 18, 9), team: 'platform cover' }
	];
}

export function getDashboard(): Dashboard {
	const now = Date.now();
	const state = scenario();

	const empty = state === 'empty';
	const busy = state === 'active' || state === 'degraded';
	const youAreOnCall = state !== 'not-on-call' && !empty;

	return {
		now,

		onboarding: empty ? { completed: ['schedule'], dismissed: false } : null,

		incidents: busy ? activeIncidents() : [],
		alerts: busy ? alerts(now) : [],
		alertVolume: busy
			? [2, 1, 0, 0, 1, 3, 2, 1, 0, 2, 4, 3, 1, 1, 0, 1, 2, 5, 3, 2, 1, 2, 3, 4]
			: [],
		overdue: busy ? overdue(now) : [],

		onCallNow: empty ? [] : onCallNow(now, youAreOnCall),
		myShifts: empty ? [] : myShifts(now),

		instance:
			state === 'degraded'
				? {
						selfHosted: true,
						workersHealthy: 1,
						workersTotal: 3,
						checkedAt: new Date(now - 2 * MINUTE).toISOString()
					}
				: {
						selfHosted: false,
						workersHealthy: 3,
						workersTotal: 3,
						checkedAt: new Date(now - 2 * MINUTE).toISOString()
					}
	};
}

export function getNavCounts(): { openIncidents: number } {
	return { openIncidents: listIncidents().filter(isActive).length };
}
