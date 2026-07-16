<script lang="ts">
	import ClockIcon from '@lucide/svelte/icons/clock';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import ListChecksIcon from '@lucide/svelte/icons/list-checks';
	import { Button } from '$lib/components/ui/button';
	import type { OverdueItem, OverdueKind } from '$lib/dashboard';
	import { formatDue } from '$lib/time';
	import QuietRow from './quiet-row.svelte';
	import RailCard from './rail-card.svelte';

	let { items, now }: { items: OverdueItem[]; now: number } = $props();

	const ICON = {
		update: ClockIcon,
		'follow-up': ListChecksIcon,
		postmortem: FileTextIcon
	} satisfies Record<OverdueKind, unknown>;
</script>

<RailCard title="Overdue">
	{#if items.length === 0}
		<QuietRow text="Nothing overdue." class="px-4 py-3.5" />
	{:else}
		{#each items as item (item.id)}
			{@const Icon = ICON[item.kind]}
			<div class="flex items-center gap-2.5 border-t px-4 py-2.5 first:border-t-0">
				<span
					class="flex size-[26px] shrink-0 items-center justify-center rounded-full border"
					style="background: var(--{item.tone}-wash); border-color: var(--{item.tone}-edge)"
				>
					<Icon class="size-[13px]" style="color: var(--{item.tone}-ink)" />
				</span>

				<div class="min-w-0 flex-1">
					<div class="text-[13px] leading-[1.35] font-medium">{item.title}</div>
					<div class="mt-px font-mono text-[11px]" style="color: var(--{item.tone}-ink)">
						{formatDue(item.dueAt, now)}
					</div>
				</div>

				<Button variant="ghost" size="sm" href={item.href}>{item.action}</Button>
			</div>
		{/each}
	{/if}
</RailCard>
