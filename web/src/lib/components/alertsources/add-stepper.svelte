<script lang="ts">
	import CheckIcon from '@lucide/svelte/icons/check';
	import { ADD_STEPS } from '$lib/alertsources';

	let { step }: { step: number } = $props();
</script>

<div class="flex items-center gap-2">
	{#each ADD_STEPS as label, index (label)}
		{@const state = index === step ? 'current' : index < step ? 'done' : 'todo'}
		{#if index > 0}
			<span class="h-px min-w-[18px] flex-1 {index <= step ? 'bg-primary' : 'bg-border'}"></span>
		{/if}
		<span
			class="inline-flex items-center gap-[7px] text-[12px] whitespace-nowrap {state === 'current'
				? 'text-foreground font-semibold'
				: state === 'done'
					? 'text-muted-foreground'
					: 'text-subtle-foreground'}"
		>
			<span
				class="flex size-5 shrink-0 items-center justify-center rounded-full border font-mono text-[10.5px] {state ===
				'current'
					? 'bg-brand-wash border-brand-edge text-brand-foreground'
					: state === 'done'
						? 'bg-primary border-primary text-primary-foreground'
						: 'bg-inset border-border'}"
			>
				{#if state === 'done'}
					<CheckIcon class="size-[11px]" />
				{:else}
					{index + 1}
				{/if}
			</span>
			{label}
		</span>
	{/each}
</div>
