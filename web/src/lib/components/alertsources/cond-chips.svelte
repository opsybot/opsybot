<script lang="ts">
	import type { Condition } from '$lib/alertsources';

	let { conditions }: { conditions: Condition[] } = $props();
</script>

<div class="flex flex-wrap items-center gap-1.5">
	{#each conditions as condition, index (index)}
		{#if index > 0}
			<span class="text-subtle-foreground text-[10.5px] tracking-[0.05em] uppercase">and</span>
		{/if}
		<span
			class="bg-inset inline-flex items-center gap-[5px] rounded-sm border px-2 py-[3px] font-mono text-[11.5px]"
		>
			<span class="text-subtle-foreground">{condition.field}</span>
			<span class="text-brand-foreground">{condition.op}</span>
			<span>{condition.value}</span>
		</span>
	{/each}
</div>
