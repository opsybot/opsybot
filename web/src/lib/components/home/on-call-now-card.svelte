<script lang="ts">
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import type { OnCallEntry } from '$lib/dashboard';
	import { formatUtcDate, formatUtcTime } from '$lib/time';
	import RailCard from './rail-card.svelte';

	let { entries, now }: { entries: OnCallEntry[]; now: number } = $props();

	const today = $derived(formatUtcDate(new Date(now).toISOString()));

	function until(iso: string): string {
		const suffix = formatUtcDate(iso) === today ? '' : ' tomorrow';
		return `until ${formatUtcTime(iso)}${suffix}`;
	}
</script>

<RailCard title="On call now">
	{#each entries as entry (entry.team)}
		<div class="flex items-center gap-2.5 border-t px-4 py-2.5 first:border-t-0">
			<UserAvatar name={entry.name} size="xs" onCall />
			<div class="min-w-0 flex-1">
				<div class="flex items-center gap-1.5 text-[13px] font-medium">
					{entry.name}
					{#if entry.you}
						<Badge tone="brand" size="sm">you</Badge>
					{/if}
				</div>
				<div class="text-subtle-foreground mt-px font-mono text-[11px]">
					{entry.team} · {until(entry.until)}
				</div>
			</div>
		</div>
	{/each}
</RailCard>
