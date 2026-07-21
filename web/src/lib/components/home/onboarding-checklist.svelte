<script lang="ts">
	import CheckIcon from '@lucide/svelte/icons/check';
	import FlaskConicalIcon from '@lucide/svelte/icons/flask-conical';
	import XIcon from '@lucide/svelte/icons/x';
	import Progress from '$lib/components/progress.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import type { Onboarding } from '$lib/dashboard';
	import { ws } from '$lib/navigation';
	import { ONBOARDING_STEPS } from '$lib/onboarding';

	let {
		onboarding,
		selfHosted,
		ondismiss
	}: {
		onboarding: Onboarding;
		selfHosted: boolean;
		ondismiss: () => void;
	} = $props();

	const done = $derived(new Set(onboarding.completed));
	const doneCount = $derived(ONBOARDING_STEPS.filter((step) => done.has(step.id)).length);
	const complete = $derived(doneCount === ONBOARDING_STEPS.length);
</script>

<section class="bg-card overflow-hidden rounded-xl border" aria-label="Onboarding checklist">
	<header class="flex items-center gap-[18px] border-b px-[18px] py-4">
		<div class="min-w-0">
			<div class="text-[15px] font-semibold">
				{complete ? 'Setup complete' : 'Get set up'}
			</div>
			<div class="text-subtle-foreground mt-0.5 text-[12.5px]">
				{complete
					? 'This workspace is ready to take its first real page.'
					: 'Five steps from empty workspace to your first real page.'}
			</div>
		</div>

		<div class="ml-auto flex items-center gap-2.5">
			<span class="text-muted-foreground font-mono text-xs">
				{doneCount} of {ONBOARDING_STEPS.length}
			</span>
			<Progress value={doneCount} max={ONBOARDING_STEPS.length} size="sm" class="w-[120px]" />
		</div>

		<Button variant="ghost" size="icon-sm" aria-label="Dismiss checklist" onclick={ondismiss}>
			<XIcon />
		</Button>
	</header>

	{#if complete}
		<div class="flex items-center gap-3 px-[18px] py-4">
			<span
				class="bg-primary shadow-glow flex size-[30px] shrink-0 items-center justify-center rounded-full"
			>
				<CheckIcon class="text-primary-foreground size-[15px]" />
			</span>
			<span class="text-muted-foreground text-[13.5px]">
				All five steps done. This card disappears once you dismiss it.
			</span>
			<Button size="sm" class="ml-auto" onclick={ondismiss}>Dismiss checklist</Button>
		</div>
	{:else}
		<div class="flex flex-col">
			{#each ONBOARDING_STEPS as step, index (step.id)}
				{@const stepDone = done.has(step.id)}
				<div class="flex items-center gap-3.5 border-t px-[18px] py-[13px] first:border-t-0">
					<span
						aria-hidden="true"
						class="flex size-[26px] shrink-0 items-center justify-center rounded-full border text-xs font-semibold {stepDone
							? 'bg-primary border-primary'
							: 'border-input text-muted-foreground'}"
					>
						{#if stepDone}
							<CheckIcon class="text-primary-foreground size-3.5" />
						{:else}
							{index + 1}
						{/if}
					</span>

					<div class="min-w-0 flex-1">
						<div
							class="text-[13.5px] font-medium {stepDone ? 'text-subtle-foreground' : ''}"
						>
							{step.title}
						</div>
						<div class="text-subtle-foreground mt-px text-[12.5px]">{step.description}</div>
					</div>

					{#if stepDone}
						<span class="text-brand-foreground px-2 text-[12.5px] font-medium">Done</span>
					{:else}
						<Button variant="secondary" size="sm" href={ws(step.href)}>
							<step.icon data-icon="inline-start" />
							{step.action}
						</Button>
					{/if}
				</div>
			{/each}
		</div>
	{/if}

	{#if !selfHosted && !complete}
		<footer class="bg-inset flex items-center gap-2.5 border-t px-[18px] py-3">
			<FlaskConicalIcon class="text-subtle-foreground size-4 shrink-0" />
			<span class="text-muted-foreground flex-1 text-[12.5px] leading-normal">
				Want to look around first? Load sample data: 3 incidents, 2 services, and a schedule, all
				labelled <Badge tone="neutral" size="sm">Sample</Badge> and removable in one click.
			</span>
			<Button variant="ghost" size="sm">Load sample data</Button>
		</footer>
	{/if}
</section>
