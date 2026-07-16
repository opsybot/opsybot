<script lang="ts">
	import ActivityIcon from '@lucide/svelte/icons/activity';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import MegaphoneIcon from '@lucide/svelte/icons/megaphone';
	import ScaleIcon from '@lucide/svelte/icons/scale';
	import WrenchIcon from '@lucide/svelte/icons/wrench';
	import { Badge } from '$lib/components/ui/badge';
	import { ENTRY_TYPES, type EntryType, type TimelineEntry } from '$lib/incidents';
	import { formatUtcTime } from '$lib/time';

	let { entry, last }: { entry: TimelineEntry; last: boolean } = $props();

	const ICON: Record<EntryType, typeof ActivityIcon> = {
		status: ActivityIcon,
		communication: MegaphoneIcon,
		action: WrenchIcon,
		observation: EyeIcon,
		decision: ScaleIcon
	};

	const label = $derived(ENTRY_TYPES.find((type) => type.id === entry.type)?.label ?? entry.type);
	const Icon = $derived(ICON[entry.type]);

	const local = $derived(
		new Date(entry.at).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
	);
</script>

<div class="grid grid-cols-[92px_26px_minmax(0,1fr)] gap-x-2.5">
	<div class="flex flex-col gap-px pt-[3px] text-right">
		<span class="text-muted-foreground font-mono text-xs">
			{formatUtcTime(entry.at).replace(' UTC', '')} UTC
		</span>
		<span class="text-subtle-foreground font-mono text-[10px]">{local} local</span>
	</div>

	<div class="flex flex-col items-center">
		<span
			title={label}
			class="bg-popover text-muted-foreground flex size-[22px] shrink-0 items-center justify-center rounded-full border"
		>
			<Icon class="size-[11px]" />
		</span>
		{#if !last}
			<span class="bg-border min-h-2 w-px flex-1"></span>
		{/if}
	</div>

	<div class="min-w-0 pt-0.5 pb-4">
		<div class="mb-[3px] flex flex-wrap items-center gap-[7px]">
			<span class="text-[12.5px] font-semibold">{entry.actor}</span>
			{#if entry.ai}
				<Badge tone="brand" size="sm">Opsybot</Badge>
			{/if}
			<Badge tone="neutral" size="sm">{label}</Badge>
			{#if entry.retro}
				<Badge tone="info" size="sm">retroactive</Badge>
			{/if}
			{#if entry.edited}
				<span class="text-subtle-foreground font-mono text-[10px]">edited</span>
			{/if}
		</div>
		<div class="text-muted-foreground text-[13.5px] leading-[1.55] [overflow-wrap:anywhere]">
			{entry.text}
		</div>
	</div>
</div>
