<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import GitBranchIcon from '@lucide/svelte/icons/git-branch';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import { Button } from '$lib/components/ui/button';
	import * as Select from '$lib/components/ui/select';
	import { Separator } from '$lib/components/ui/separator';
	import { BRANCH_DEFS, TONE_STYLE, type Branch, type BranchKind } from '$lib/escalation';

	let {
		node,
		onchangecondition,
		onremove,
		ondeselect
	}: {
		node: Branch;
		onchangecondition: (id: string, on: BranchKind) => void;
		onremove: (id: string) => void;
		ondeselect: () => void;
	} = $props();

	const def = $derived(BRANCH_DEFS[node.on]);
	const conditions = $derived(
		(Object.keys(BRANCH_DEFS) as BranchKind[]).map((key) => ({
			value: key,
			label: BRANCH_DEFS[key].title.replace(/^\w/, (c) => c.toUpperCase())
		}))
	);
	const currentLabel = $derived(conditions.find((c) => c.value === node.on)?.label);
</script>

<button
	type="button"
	class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[11.5px] transition-colors"
	onclick={ondeselect}
>
	<ArrowLeftIcon class="size-3" />
	Policy settings
</button>

<div class="flex items-center gap-2 text-[14px] font-semibold">
	<span class="text-brand-foreground flex size-[22px] shrink-0 items-center justify-center rounded-sm border bg-[var(--ink-4)]">
		<GitBranchIcon class="size-[13px]" />
	</span>
	Branch
</div>

<div>
	<div class="text-subtle-foreground mb-2.5 text-[11px] tracking-[0.08em] uppercase">
		Split the escalation by
	</div>
	<Select.Root type="single" value={node.on} onValueChange={(value) => onchangecondition(node.id, value as BranchKind)}>
		<Select.Trigger size="sm" class="w-full" aria-label="Branch condition">{currentLabel}</Select.Trigger>
		<Select.Content>
			<Select.Group>
				{#each conditions as condition (condition.value)}
					<Select.Item value={condition.value} label={condition.label}>{condition.label}</Select.Item>
				{/each}
			</Select.Group>
		</Select.Content>
	</Select.Root>
	<p class="text-subtle-foreground mt-2 text-[11.5px] leading-[1.5]">
		The escalation follows whichever lane matches the incident. Lanes don't rejoin — each ends on its
		own.
	</p>
</div>

<div>
	<div class="text-subtle-foreground mb-2.5 text-[11px] tracking-[0.08em] uppercase">Lanes</div>
	<div class="flex flex-col gap-2">
		{#each def.lanes as lane (lane.key)}
			<div class="text-muted-foreground flex items-center gap-2 text-[12.5px]">
				<span class="size-2 shrink-0 rounded-full" style="background:{TONE_STYLE[lane.tone].color}"></span>
				{lane.label}
			</div>
		{/each}
	</div>
</div>

<Separator />

<Button
	variant="ghost"
	size="sm"
	class="text-critical-ink hover:text-critical-ink self-start"
	onclick={() => onremove(node.id)}
>
	<Trash2Icon data-icon="inline-start" />
	Delete branch & its lanes
</Button>
