<script lang="ts">
	let { step, total }: { step: number; total: number } = $props();
</script>

<div class="mb-5 flex items-center gap-1.5" aria-label="Step {step} of {total}">
	{#each Array.from({ length: total }, (_, index) => index + 1) as position (position)}
		<span
			class="h-[3px] w-[22px] rounded-full {position <= step ? 'bg-primary' : 'bg-input'}"
		></span>
	{/each}
	<span class="text-subtle-foreground ml-1.5 text-[11px]">Step {step} of {total}</span>
</div>
