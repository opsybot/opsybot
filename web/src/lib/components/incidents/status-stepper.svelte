<script lang="ts">
	import CheckIcon from '@lucide/svelte/icons/check';
	import { enhance } from '$app/forms';
	import { canMoveTo, STAGES, type IncidentStage } from '$lib/incidents';
	import { cn } from '$lib/utils';

	let {
		stage,
		action,
		onresolve
	}: {
		stage: IncidentStage;
		action: string;
		onresolve: () => void;
	} = $props();

	const current = $derived(STAGES.indexOf(stage));
</script>

<form
	method="POST"
	{action}
	use:enhance
	class="bg-card flex flex-wrap items-center gap-1.5 rounded-xl border px-3.5 py-2.5"
	aria-label="Incident status"
>
	{#each STAGES as step, index (step)}
		{@const done = index < current}
		{@const active = index === current}
		{@const allowed = canMoveTo(stage, step)}

		{#if index > 0}
			<span class={cn('h-px w-4 shrink-0', index <= current ? 'bg-primary' : 'bg-border')}></span>
		{/if}

		<button
			type={step === 'resolved' ? 'button' : 'submit'}
			name="status"
			value={step}
			disabled={!allowed}
			aria-current={active ? 'step' : undefined}
			title={allowed
				? `Move to ${step}`
				: index > current + 1
					? 'Move one stage at a time'
					: undefined}
			onclick={() => {
				if (step === 'resolved') onresolve();
			}}
			class={cn(
				'inline-flex items-center gap-[7px] rounded-full border border-transparent px-[11px] py-[5px] text-xs transition-colors duration-[120ms] ease-out',
				active
					? 'text-brand-foreground bg-brand-wash border-brand-edge font-semibold'
					: done
						? 'text-muted-foreground'
						: 'text-subtle-foreground',
				allowed
					? 'hover:text-foreground hover:border-border-strong cursor-pointer'
					: 'cursor-default disabled:opacity-60'
			)}
		>
			<span
				class={cn(
					'bg-inset flex size-4 shrink-0 items-center justify-center rounded-full border',
					done && 'bg-primary border-primary text-primary-foreground',
					active && 'border-primary shadow-glow'
				)}
			>
				{#if done}
					<CheckIcon class="size-2.5" />
				{/if}
			</span>
			{step}
		</button>
	{/each}
</form>
