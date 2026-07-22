export function formatUtc(iso: string): string {
	const d = new Date(iso);
	const pad = (n: number) => String(n).padStart(2, '0');
	return (
		`${d.getUTCFullYear()}-${pad(d.getUTCMonth() + 1)}-${pad(d.getUTCDate())} ` +
		`${pad(d.getUTCHours())}:${pad(d.getUTCMinutes())} UTC`
	);
}

export function formatUtcDate(iso: string): string {
	return formatUtc(iso).slice(0, 10);
}

export function formatUtcTime(iso: string): string {
	return formatUtc(iso).slice(11);
}

export function formatAge(ms: number): string {
	const total = Math.max(0, Math.floor(ms / 1000));
	const minutes = Math.floor(total / 60);
	const seconds = total % 60;
	return `${minutes}m ${String(seconds).padStart(2, '0')}s`;
}

export function formatDuration(seconds: number): string {
	if (seconds <= 0) return 'none';
	if (seconds % 86_400 === 0) return `${seconds / 86_400} d`;
	if (seconds % 3_600 === 0) return `${seconds / 3_600} h`;
	if (seconds < 60) return `${seconds} s`;
	return `${Math.round(seconds / 60)} m`;
}

export function formatSince(ms: number): string {
	const minutes = Math.round(ms / 60_000);
	if (minutes < 1) return 'just now';
	if (minutes < 60) return `${minutes} m ago`;
	const hours = Math.round(minutes / 60);
	if (hours < 24) return `${hours} h ago`;
	return `${Math.round(hours / 24)} d ago`;
}

export function formatDue(dueIso: string, now: number): string {
	const overdueMs = now - Date.parse(dueIso);
	if (overdueMs <= 0) return `due ${formatUtcDate(dueIso)}`;

	const minutes = Math.round(overdueMs / 60_000);
	if (minutes < 60) return `${minutes} min overdue`;
	const hours = Math.round(minutes / 60);
	if (hours < 24) return `${hours} h overdue`;
	const days = Math.round(hours / 24);
	return `${days} ${days === 1 ? 'day' : 'days'} overdue`;
}

export function formatRemaining(ms: number): string {
	const minutes = Math.max(0, Math.round(ms / 60_000));
	if (minutes < 60) return `${minutes} m`;

	const hours = Math.floor(minutes / 60);
	if (hours < 24) {
		const rest = minutes % 60;
		return rest ? `${hours} h ${rest} m` : `${hours} h`;
	}

	return `${Math.round(hours / 24)} d`;
}

export function formatSpan(startIso: string, endIso: string): string {
	const sameDay = formatUtcDate(startIso) === formatUtcDate(endIso);
	const start = formatUtc(startIso).replace(' UTC', '');
	const end = sameDay ? formatUtcTime(endIso).replace(' UTC', '') : formatUtc(endIso).replace(' UTC', '');
	return `${start}–${end} UTC`;
}
