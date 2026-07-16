<script lang="ts">
	import { onDestroy, tick } from 'svelte';
	import SparklesIcon from '@lucide/svelte/icons/sparkles';
	import { Button } from '$lib/components/ui/button';
	import Tag from '$lib/components/tag.svelte';
	import SurfaceOff from '$lib/components/ai/surface-off.svelte';

	let { enabled }: { enabled: boolean } = $props();

	let phase = $state<'idle' | 'loading' | 'ready'>('idle');
	let region = $state<HTMLElement | null>(null);
	let timer: ReturnType<typeof setTimeout>;
	async function run() {
		phase = 'loading';
		clearTimeout(timer);
		timer = setTimeout(() => (phase = 'ready'), 1600);
		await tick();
		region?.focus();
	}
	onDestroy(() => clearTimeout(timer));

	const GROUPS = [
		['database', '2 alerts — disk usage on db-3 and slow vacuum on db-1. Same cluster, likely one cleanup task.'],
		['edge', '3 alerts — cert expiries within 30 days across three domains. One renewal run covers all.'],
		['events-worker', '2 alerts — queue depth briefly above threshold, self-recovered twice. Consider raising the threshold.']
	];
</script>

{#if !enabled}
	<SurfaceOff>The digest needs a model — currently unavailable.</SurfaceOff>
{:else}
	<div class="bg-card rounded-xl border px-4 py-3.5">
		<header class="flex items-center gap-2 {phase ==='idle' ? '' : 'mb-2.5'}">
			<SparklesIcon class="text-primary size-3.5" />
			<span class="text-[13.5px] font-semibold">Low-urgency digest</span>
			<div class="flex-1"></div>
			{#if phase ==='idle'}
				<Button size="sm" variant="secondary" onclick={run}>Summarize 7 open alerts</Button>
			{/if}
		</header>
		<div bind:this={region} tabindex="-1" role="status" aria-live="polite" class="outline-none">
			{#if phase ==='loading'}
				<div class="flex items-center gap-2.5">
					<span
						class="border-border border-t-primary size-4 shrink-0 animate-spin rounded-full border-2 [animation-duration:0.8s] motion-reduce:animate-none"
						aria-hidden="true"
					></span>
					<span class="text-muted-foreground text-[12.5px]">Grouping by service and pattern…</span>
				</div>
			{:else if phase ==='ready'}
				<div class="flex flex-col gap-2">
					{#each GROUPS as [service, text] (service)}
						<div class="flex items-start gap-2.5">
							<Tag>{service}</Tag>
							<span class="text-muted-foreground text-[12.5px] leading-[1.55]">{text}</span>
						</div>
					{/each}
					<p class="text-subtle-foreground mt-0.5 mb-0 text-[10.5px]">
						Drafted by Opsybot — alerts stay untouched until you act.
					</p>
				</div>
			{/if}
		</div>
	</div>
{/if}
