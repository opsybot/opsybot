<script lang="ts">
	import GitBranchIcon from '@lucide/svelte/icons/git-branch';
	import PlusIcon from '@lucide/svelte/icons/plus';

	let { onaddlevel, onaddbranch }: { onaddlevel: () => void; onaddbranch: () => void } = $props();

	const btn =
		'border-border-strong text-muted-foreground hover:border-brand-edge hover:text-brand-foreground hover:bg-brand-wash inline-flex items-center gap-1.5 rounded-md border border-dashed px-[11px] py-[5px] text-[12px] font-medium transition-colors';
</script>

<div class="my-1.5 mb-2.5 flex gap-2">
	<button type="button" class={btn} onclick={onaddlevel}>
		<PlusIcon class="size-3" />
		Level
	</button>
	<button type="button" class={btn} onclick={onaddbranch}>
		<GitBranchIcon class="size-3" />
		Branch
	</button>
</div>
