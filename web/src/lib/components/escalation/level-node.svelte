<script lang="ts">
	import BellIcon from '@lucide/svelte/icons/bell';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import TargetRow from '$lib/components/escalation/target-row.svelte';
	import { targetInvalid, type Level } from '$lib/escalation';

	let {
		node,
		editable = false,
		selected = false,
		active = false,
		onselect,
		ondelete
	}: {
		node: Level;
		editable?: boolean;
		selected?: boolean;
		active?: boolean;
		onselect?: (id: string) => void;
		ondelete?: (id: string) => void;
	} = $props();

	const noTargets = $derived(node.targets.length === 0);
	const err = $derived(noTargets || node.targets.every(targetInvalid));
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div
	class="ep-node bg-card w-[288px] rounded-xl border shadow-[var(--shadow-xs)] {editable
		? 'cursor-pointer outline-none transition-[border-color,box-shadow,transform] hover:border-border-strong focus-visible:border-primary focus-visible:shadow-[var(--focus-ring)]'
		: ''} {selected ? 'border-primary shadow-[var(--focus-ring)]' : ''} {active
		? 'border-primary -translate-y-px shadow-[0_0_0_1px_var(--primary),var(--glow-brand)]'
		: ''} {err ? 'border-critical-edge' : ''}"
	tabindex={editable ? 0 : undefined}
	role={editable ? 'button' : undefined}
	onclick={editable ? () => onselect?.(node.id) : undefined}
	onkeydown={editable
		? (event) => {
				if (event.target !== event.currentTarget) return;
				if (event.key === 'Enter' || event.key === ' ') {
					event.preventDefault();
					onselect?.(node.id);
				}
			}
		: undefined}
>
	<header class="bg-inset flex items-center gap-2 border-b py-2 pr-2.5 pl-3">
		<span
			class="text-muted-foreground flex size-[22px] shrink-0 items-center justify-center rounded-sm border bg-[var(--ink-4)]"
		>
			<BellIcon class="size-[13px]" />
		</span>
		<span class="text-[12.5px] font-semibold">Notify</span>
		<Badge tone={node.mode === 'rr' ? 'info' : 'neutral'} size="sm" class="ml-auto">
			{node.mode === 'rr' ? 'round-robin' : 'all at once'}
		</Badge>
		{#if editable}
			<Button
				variant="ghost"
				size="icon-sm"
				aria-label="Delete level"
				onclick={(event) => {
					event.stopPropagation();
					ondelete?.(node.id);
				}}
			>
				<Trash2Icon />
			</Button>
		{/if}
	</header>
	<div class="flex flex-col gap-1.5 px-3 py-2.5">
		{#if noTargets}
			<div class="text-critical-ink flex items-center gap-1.5 text-[12px]">
				<TriangleAlertIcon class="size-3" />
				Notifies no one. Add a target
			</div>
		{:else}
			{#each node.targets as target, index (index)}
				<TargetRow {target} />
			{/each}
			{#if err}
				<div class="text-critical-ink flex items-center gap-1.5 text-[11.5px]">
					<TriangleAlertIcon class="size-[11px]" />
					No one here can be paged
				</div>
			{/if}
		{/if}
	</div>
</div>
