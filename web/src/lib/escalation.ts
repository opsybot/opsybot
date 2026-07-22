export type TargetType = 'person' | 'schedule' | 'team' | 'webhook';
export type Target = { type: TargetType; ref: string; value: string; invalid?: boolean };
export type NotifyMode = 'all' | 'rr';

export type Level = {
	id: string;
	type: 'level';
	targets: Target[];
	mode: NotifyMode;
	wait: string;
	addType: TargetType;
};

export type BranchKind = 'priority' | 'hours';
export type Hours = { days: number[]; start: string; end: string; timezone: string };
export type Lane = { id: string; key: string; nodes: EscNode[] };
export type Branch = { id: string; type: 'branch'; on: BranchKind; hours: Hours; lanes: Lane[] };
export type EscNode = Level | Branch;

export type Tree = { name: string; team: string; repeat: string; ack: string; nodes: EscNode[] };
export type Scenario = { priority: string; hours: string };

export type DirectoryMember = { id: string; name: string; email: string; active: boolean };
export type DirectoryEntry = { id: string; slug: string; name: string };
export type Directory = {
	members: DirectoryMember[];
	schedules: DirectoryEntry[];
	teams: DirectoryEntry[];
	webhooks: DirectoryEntry[];
};

export const DEFAULT_HOURS: Hours = { days: [1, 2, 3, 4, 5], start: '09:00', end: '18:00', timezone: 'UTC' };

export const WEEKDAYS = [
	{ value: 1, label: 'Mon' },
	{ value: 2, label: 'Tue' },
	{ value: 3, label: 'Wed' },
	{ value: 4, label: 'Thu' },
	{ value: 5, label: 'Fri' },
	{ value: 6, label: 'Sat' },
	{ value: 0, label: 'Sun' }
];

export type Tone = 'critical' | 'warning' | 'info' | 'neutral';

export const TARGET_TYPES: { value: TargetType; label: string; icon: string }[] = [
	{ value: 'person', label: 'Person', icon: 'user' },
	{ value: 'schedule', label: 'Schedule', icon: 'calendar-clock' },
	{ value: 'team', label: 'Team', icon: 'users' },
	{ value: 'webhook', label: 'Webhook', icon: 'webhook' }
];

export const WAIT_OPTIONS = [
	{ value: '1', label: '1 minute' },
	{ value: '2', label: '2 minutes' },
	{ value: '5', label: '5 minutes' },
	{ value: '10', label: '10 minutes' },
	{ value: '15', label: '15 minutes' },
	{ value: '30', label: '30 minutes' },
	{ value: '60', label: '1 hour' }
];

export const REPEAT_OPTIONS = [
	{ value: '0', label: "Don't repeat" },
	{ value: '1', label: 'Repeat once' },
	{ value: '2', label: 'Repeat twice' },
	{ value: '3', label: 'Repeat 3 times' }
];

export const ACK_OPTIONS = [
	{ value: '0', label: 'Never' },
	{ value: '10', label: 'After 10 minutes' },
	{ value: '30', label: 'After 30 minutes' },
	{ value: '60', label: 'After 1 hour' },
	{ value: '240', label: 'After 4 hours' }
];

export function targetOptions(directory: Directory, type: TargetType): { ref: string; label: string; invalid?: boolean }[] {
	switch (type) {
		case 'person':
			return directory.members.map((m) => ({ ref: m.id, label: m.name, invalid: !m.active }));
		case 'schedule':
			return directory.schedules.map((s) => ({ ref: s.id, label: s.slug }));
		case 'team':
			return directory.teams.map((t) => ({ ref: t.id, label: t.slug }));
		case 'webhook':
			return directory.webhooks.map((w) => ({ ref: w.id, label: w.slug }));
	}
}

export function targetIcon(type: TargetType): string {
	return TARGET_TYPES.find((entry) => entry.value === type)?.icon ?? 'user';
}

export function targetInvalid(target: Target): boolean {
	return target.invalid === true;
}

type BranchDef = {
	icon: string;
	title: string;
	verb: string;
	lanes: { key: string; label: string; tone: Tone }[];
	pick: (scenario: Scenario) => string;
	control: { id: keyof Scenario; label: string; options: { value: string; label: string }[] };
};

export const BRANCH_DEFS: Record<BranchKind, BranchDef> = {
	priority: {
		icon: 'flag',
		title: 'alert severity',
		verb: 'Route by severity',
		lanes: [
			{ key: 'high', label: 'Critical · High', tone: 'critical' },
			{ key: 'low', label: 'Warning', tone: 'neutral' }
		],
		pick: (scenario) => scenario.priority,
		control: {
			id: 'priority',
			label: 'Alert severity',
			options: [
				{ value: 'high', label: 'Critical · High' },
				{ value: 'low', label: 'Warning' }
			]
		}
	},
	hours: {
		icon: 'clock',
		title: 'time of day',
		verb: 'Route by working hours',
		lanes: [
			{ key: 'business', label: 'Working hours', tone: 'info' },
			{ key: 'off', label: 'Off hours', tone: 'warning' }
		],
		pick: (scenario) => scenario.hours,
		control: {
			id: 'hours',
			label: 'Time of day',
			options: [
				{ value: 'business', label: 'Working hours' },
				{ value: 'off', label: 'Off hours' }
			]
		}
	}
};

export function laneMeta(branch: Branch, lane: Lane): { key: string; label: string; tone: Tone; icon: string } {
	const def = BRANCH_DEFS[branch.on];
	const meta = def.lanes.find((entry) => entry.key === lane.key) ?? def.lanes[0];
	return { ...meta, icon: def.icon };
}

export const TONE_STYLE: Record<Tone, { bg: string; border: string; color: string }> = {
	critical: { bg: 'var(--critical-wash)', border: 'var(--critical-edge)', color: 'var(--critical-ink)' },
	warning: { bg: 'var(--warning-wash)', border: 'var(--warning-edge)', color: 'var(--warning-ink)' },
	info: { bg: 'var(--info-wash)', border: 'var(--info-edge)', color: 'var(--info-ink)' },
	neutral: { bg: 'var(--surface-inset)', border: 'var(--border)', color: 'var(--muted-foreground)' }
};

let counter = 0;
export function uid(prefix: string): string {
	counter += 1;
	return `${prefix}-${counter.toString(36)}${Math.random().toString(36).slice(2, 5)}`;
}

export function mkLevel(over: Partial<Level> = {}): Level {
	return { id: uid('lv'), type: 'level', targets: [], mode: 'all', wait: '5', addType: 'person', ...over };
}

export function mkBranch(on: BranchKind): Branch {
	return {
		id: uid('br'),
		type: 'branch',
		on,
		hours: { ...DEFAULT_HOURS, days: [...DEFAULT_HOURS.days] },
		lanes: BRANCH_DEFS[on].lanes.map((lane) => ({ id: uid('ln'), key: lane.key, nodes: [mkLevel()] }))
	};
}

export function updateNodes(nodes: EscNode[], id: string, fn: (level: Level) => Level): EscNode[] {
	return nodes.map((node) => {
		if (node.id === id && node.type === 'level') return fn(node);
		if (node.type === 'branch') {
			return { ...node, lanes: node.lanes.map((lane) => ({ ...lane, nodes: updateNodes(lane.nodes, id, fn) })) };
		}
		return node;
	});
}

export function changeCondition(nodes: EscNode[], id: string, on: BranchKind): EscNode[] {
	return nodes.map((node) => {
		if (node.id === id && node.type === 'branch') {
			return { ...node, on, lanes: node.lanes.map((lane, index) => ({ ...lane, key: BRANCH_DEFS[on].lanes[index].key })) };
		}
		if (node.type === 'branch') {
			return { ...node, lanes: node.lanes.map((lane) => ({ ...lane, nodes: changeCondition(lane.nodes, id, on) })) };
		}
		return node;
	});
}

export function removeNodeDeep(nodes: EscNode[], id: string): EscNode[] {
	return nodes
		.filter((node) => node.id !== id)
		.map((node) =>
			node.type === 'branch'
				? { ...node, lanes: node.lanes.map((lane) => ({ ...lane, nodes: removeNodeDeep(lane.nodes, id) })) }
				: node
		);
}

export function insertLevelAfterDeep(nodes: EscNode[], afterId: string, newNode: EscNode): EscNode[] {
	const out: EscNode[] = [];
	for (const node of nodes) {
		const mapped =
			node.type === 'branch'
				? { ...node, lanes: node.lanes.map((lane) => ({ ...lane, nodes: insertLevelAfterDeep(lane.nodes, afterId, newNode) })) }
				: node;
		out.push(mapped);
		if (node.id === afterId) out.push(newNode);
	}
	return out;
}

function appendToLaneDeep(nodes: EscNode[], laneId: string, newNode: EscNode): EscNode[] {
	return nodes.map((node) =>
		node.type === 'branch'
			? {
					...node,
					lanes: node.lanes.map((lane) =>
						lane.id === laneId
							? { ...lane, nodes: [...lane.nodes, newNode] }
							: { ...lane, nodes: appendToLaneDeep(lane.nodes, laneId, newNode) }
					)
				}
			: node
	);
}

export function appendNode(tree: Tree, ownerId: string, newNode: EscNode): Tree {
	if (ownerId === 'root') return { ...tree, nodes: [...tree.nodes, newNode] };
	return { ...tree, nodes: appendToLaneDeep(tree.nodes, ownerId, newNode) };
}

export function moveNodeDeep(nodes: EscNode[], id: string, dir: -1 | 1): EscNode[] {
	const index = nodes.findIndex((node) => node.id === id);
	if (index !== -1) {
		const to = index + dir;
		if (to < 0 || to >= nodes.length || nodes[to].type === 'branch') return nodes;
		const out = nodes.slice();
		[out[index], out[to]] = [out[to], out[index]];
		return out;
	}
	return nodes.map((node) =>
		node.type === 'branch'
			? { ...node, lanes: node.lanes.map((lane) => ({ ...lane, nodes: moveNodeDeep(lane.nodes, id, dir) })) }
			: node
	);
}

export function nodeSiblings(
	nodes: EscNode[],
	id: string
): { canUp: boolean; canDown: boolean } | null {
	const index = nodes.findIndex((node) => node.id === id);
	if (index !== -1) {
		return { canUp: index > 0, canDown: index < nodes.length - 1 && nodes[index + 1].type !== 'branch' };
	}
	for (const node of nodes) {
		if (node.type === 'branch') {
			for (const lane of node.lanes) {
				const found = nodeSiblings(lane.nodes, id);
				if (found) return found;
			}
		}
	}
	return null;
}

export function findNode(nodes: EscNode[], id: string): EscNode | null {
	for (const node of nodes) {
		if (node.id === id) return node;
		if (node.type === 'branch') {
			for (const lane of node.lanes) {
				const found = findNode(lane.nodes, id);
				if (found) return found;
			}
		}
	}
	return null;
}

export function levelUnreachable(level: Level): boolean {
	return level.targets.length > 0 && level.targets.every(targetInvalid);
}

export type Analysis = {
	emptyLevel: boolean;
	unreachableLevel: boolean;
	deactivated: boolean;
	deadEnds: string[];
	hasDeadEnd: boolean;
	reach: number;
	reachValid: number;
	maxMin: number;
	maxLevels: number;
	branches: number;
};

export function analyzeTree(tree: Tree): Analysis {
	let emptyLevel = false;
	let unreachableLevel = false;
	let deactivated = false;
	const deadEnds: string[] = [];
	const targets = new Set<string>();
	const valid = new Set<string>();
	let maxMin = 0;
	let maxLevels = 0;
	let branches = 0;

	function walk(nodes: EscNode[], accMin: number, accLevels: number) {
		let min = accMin;
		let levels = accLevels;
		let branched = false;
		for (const node of nodes) {
			if (node.type === 'level') {
				if (node.targets.length === 0) emptyLevel = true;
				if (levelUnreachable(node)) unreachableLevel = true;
				for (const target of node.targets) {
					const key = `${target.type}:${target.value}`;
					targets.add(key);
					if (targetInvalid(target)) deactivated = true;
					else valid.add(key);
				}
				levels += 1;
				min += Number(node.wait) || 0;
			} else {
				branched = true;
				branches += 1;
				for (const lane of node.lanes) {
					if (lane.nodes.length === 0) deadEnds.push(lane.id);
					walk(lane.nodes, min, levels);
				}
			}
		}
		if (!branched) {
			maxMin = Math.max(maxMin, min);
			maxLevels = Math.max(maxLevels, levels);
		}
	}
	walk(tree.nodes, 0, 0);

	return {
		emptyLevel,
		unreachableLevel,
		deactivated,
		deadEnds,
		hasDeadEnd: deadEnds.length > 0,
		reach: targets.size,
		reachValid: valid.size,
		maxMin,
		maxLevels,
		branches
	};
}

export function saveBlocked(analysis: Analysis): boolean {
	return analysis.emptyLevel || analysis.unreachableLevel || analysis.hasDeadEnd || analysis.reachValid === 0;
}

export type TraceStep =
	| { id: string; kind: 'level'; t: number; targets: Target[]; mode: NotifyMode; wait: string }
	| { id: string; kind: 'branch'; on: BranchKind; laneKey: string }
	| { id: string; kind: 'end'; t: number; repeat: string };

export type Trace = { steps: TraceStep[]; totalMin: number; laneChoices: Record<string, string>; endId: string };

export function computeTrace(tree: Tree, scenario: Scenario): Trace {
	const steps: TraceStep[] = [];
	const laneChoices: Record<string, string> = {};
	let t = 0;
	let endOwner = 'root';

	function walk(nodes: EscNode[]) {
		for (const node of nodes) {
			if (node.type === 'level') {
				steps.push({ id: node.id, kind: 'level', t, targets: node.targets, mode: node.mode, wait: node.wait });
				t += Number(node.wait) || 0;
			} else {
				const key = BRANCH_DEFS[node.on].pick(scenario);
				const lane = node.lanes.find((entry) => entry.key === key) ?? node.lanes[0];
				laneChoices[node.id] = lane.id;
				steps.push({ id: node.id, kind: 'branch', on: node.on, laneKey: lane.key });
				endOwner = lane.id;
				walk(lane.nodes);
				return;
			}
		}
	}
	walk(tree.nodes);
	steps.push({ id: `end-${endOwner}`, kind: 'end', t, repeat: tree.repeat });
	return { steps, totalMin: t, laneChoices, endId: `end-${endOwner}` };
}

export function usedConditions(tree: Tree): BranchKind[] {
	const found = new Set<BranchKind>();
	function walk(nodes: EscNode[]) {
		for (const node of nodes) {
			if (node.type === 'branch') {
				found.add(node.on);
				for (const lane of node.lanes) walk(lane.nodes);
			}
		}
	}
	walk(tree.nodes);
	return [...found];
}

export type SummaryPart = { text: string; kind: 'step' | 'wait' | 'branch' };

function levelText(level: Level): string {
	const primary = level.targets[0];
	const prefix = level.mode === 'rr' ? 'Round-robin ' : '';
	if (!primary) return `${prefix}notify no one`;
	const extra = level.targets.length > 1 ? ` +${level.targets.length - 1}` : '';
	if (primary.type === 'schedule') return `Schedule ${primary.value}${extra}`;
	if (primary.type === 'team') return `Team ${primary.value}${extra}`;
	if (primary.type === 'webhook') return `Webhook ${primary.value}${extra}`;
	return `${prefix}${primary.value}${extra}`;
}

export function stepSummary(tree: Tree): SummaryPart[] {
	const parts: SummaryPart[] = [];
	for (const node of tree.nodes) {
		if (node.type === 'level') {
			parts.push({ text: levelText(node), kind: 'step' });
			if (Number(node.wait)) parts.push({ text: `wait ${node.wait} m`, kind: 'wait' });
		} else {
			parts.push({ text: `branch by ${node.on}`, kind: 'branch' });
		}
	}
	if (parts.length && parts[parts.length - 1].kind === 'wait') parts.pop();
	return parts;
}

export function firstBranchKind(tree: Tree): BranchKind | null {
	const branch = tree.nodes.find((node): node is Branch => node.type === 'branch');
	return branch ? branch.on : null;
}

export function firstDeactivatedTarget(tree: Tree): Target | null {
	let found: Target | null = null;
	function walk(nodes: EscNode[]) {
		for (const node of nodes) {
			if (found) return;
			if (node.type === 'level') {
				const bad = node.targets.find(targetInvalid);
				if (bad) {
					found = bad;
					return;
				}
			} else {
				for (const lane of node.lanes) walk(lane.nodes);
			}
		}
	}
	walk(tree.nodes);
	return found;
}

function isTargetType(value: unknown): value is TargetType {
	return typeof value === 'string' && TARGET_TYPES.some((entry) => entry.value === value);
}

function sanitizeTargets(input: unknown[]): Target[] {
	const out: Target[] = [];
	for (const raw of input) {
		if (!raw || typeof raw !== 'object') continue;
		const entry = raw as Record<string, unknown>;
		if (isTargetType(entry.type) && typeof entry.ref === 'string' && entry.ref.trim()) {
			out.push({
				type: entry.type,
				ref: entry.ref.trim(),
				value: typeof entry.value === 'string' ? entry.value : entry.ref.trim(),
				invalid: entry.invalid === true
			});
		}
	}
	return out;
}

function sanitizeHours(raw: unknown): Hours {
	if (!raw || typeof raw !== 'object') return { ...DEFAULT_HOURS, days: [...DEFAULT_HOURS.days] };
	const entry = raw as Record<string, unknown>;
	const days = Array.isArray(entry.days)
		? entry.days.filter((d): d is number => typeof d === 'number' && d >= 0 && d <= 6)
		: [];
	const time = (value: unknown, fallback: string) =>
		typeof value === 'string' && /^\d{2}:\d{2}$/.test(value) ? value : fallback;
	return {
		days: days.length ? days : [...DEFAULT_HOURS.days],
		start: time(entry.start, DEFAULT_HOURS.start),
		end: time(entry.end, DEFAULT_HOURS.end),
		timezone: typeof entry.timezone === 'string' && entry.timezone ? entry.timezone : DEFAULT_HOURS.timezone
	};
}

const MAX_DEPTH = 40;

function sanitizeNodes(input: unknown[], depth = 0): EscNode[] {
	if (depth > MAX_DEPTH) return [];
	const out: EscNode[] = [];
	for (const raw of input) {
		if (!raw || typeof raw !== 'object') continue;
		const node = raw as Record<string, unknown>;
		if (node.type === 'branch') {
			const on: BranchKind = node.on === 'hours' ? 'hours' : 'priority';
			const incoming = Array.isArray(node.lanes) ? node.lanes : [];
			const lanes: Lane[] = BRANCH_DEFS[on].lanes.map((def, index) => {
				const source =
					incoming[index] && typeof incoming[index] === 'object' ? (incoming[index] as Record<string, unknown>) : {};
				return {
					id: typeof source.id === 'string' ? source.id : uid('ln'),
					key: def.key,
					nodes: Array.isArray(source.nodes) ? sanitizeNodes(source.nodes, depth + 1) : []
				};
			});
			out.push({
				id: typeof node.id === 'string' ? node.id : uid('br'),
				type: 'branch',
				on,
				hours: sanitizeHours(node.hours),
				lanes
			});
			break;
		}
		out.push({
			id: typeof node.id === 'string' ? node.id : uid('lv'),
			type: 'level',
			targets: Array.isArray(node.targets) ? sanitizeTargets(node.targets) : [],
			mode: node.mode === 'rr' ? 'rr' : 'all',
			wait: typeof node.wait === 'string' && WAIT_OPTIONS.some((w) => w.value === node.wait) ? node.wait : '5',
			addType: isTargetType(node.addType) ? node.addType : 'person'
		});
	}
	return out;
}

export function parseTree(raw: string): { tree: Tree } | { error: string } {
	let data: unknown;
	try {
		data = JSON.parse(raw);
	} catch {
		return { error: 'Could not read the policy.' };
	}
	if (!data || typeof data !== 'object') return { error: 'Could not read the policy.' };
	const object = data as Record<string, unknown>;

	const nodes = Array.isArray(object.nodes) ? sanitizeNodes(object.nodes) : [];
	if (!nodes.length) return { error: 'A policy needs at least one level.' };

	return {
		tree: {
			name: typeof object.name === 'string' ? object.name.trim() : '',
			team: typeof object.team === 'string' ? object.team : '',
			repeat:
				typeof object.repeat === 'string' && REPEAT_OPTIONS.some((r) => r.value === object.repeat)
					? object.repeat
					: '0',
			ack:
				typeof object.ack === 'string' && ACK_OPTIONS.some((a) => a.value === object.ack)
					? object.ack
					: '0',
			nodes
		}
	};
}
