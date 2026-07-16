<script lang="ts">
	import GitBranchIcon from '@lucide/svelte/icons/git-branch';
	import { BRANCH_DEFS, type Branch } from '$lib/escalation';

	let {
		node,
		editable = false,
		selected = false,
		active = false,
		onselect
	}: {
		node: Branch;
		editable?: boolean;
		selected?: boolean;
		active?: boolean;
		onselect?: (id: string) => void;
	} = $props();

	const def = $derived(BRANCH_DEFS[node.on]);
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div
	class="ep-node bg-inset w-[288px] rounded-xl border shadow-[var(--shadow-xs)] {editable
		? 'cursor-pointer outline-none transition-[border-color,box-shadow,transform] hover:border-border-strong focus-visible:border-primary focus-visible:shadow-[var(--focus-ring)]'
		: ''} {selected ? 'border-primary shadow-[var(--focus-ring)]' : ''} {active
		? 'border-primary -translate-y-px shadow-[0_0_0_1px_var(--primary),var(--glow-brand)]'
		: ''}"
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
	<div class="flex items-center gap-2 px-3 py-[11px]">
		<span
			class="text-brand-foreground flex size-[22px] shrink-0 items-center justify-center rounded-sm border bg-[var(--ink-4)]"
		>
			<GitBranchIcon class="size-[13px]" />
		</span>
		<span class="text-[12.5px] font-semibold">Split by {def.title}</span>
	</div>
</div>
