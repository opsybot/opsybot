<script lang="ts">
	import { onDestroy } from 'svelte';
	import RotateCwIcon from '@lucide/svelte/icons/rotate-cw';
	import SparklesIcon from '@lucide/svelte/icons/sparkles';
	import XIcon from '@lucide/svelte/icons/x';
	import { toast } from 'svelte-sonner';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import SurfaceOff from '$lib/components/ai/surface-off.svelte';

	let { enabled }: { enabled: boolean } = $props();

	let loading = $state(false);
	let timer: ReturnType<typeof setTimeout>;
	function regenerate() {
		loading = true;
		clearTimeout(timer);
		timer = setTimeout(() => (loading = false), 1500);
	}
	onDestroy(() => clearTimeout(timer));
</script>

{#if !enabled}
	<SurfaceOff>
		Catch-me-up needs a model. AI is off or no model is configured —
		<a href="/ai" class="text-brand-foreground hover:underline">connect one</a>.
	</SurfaceOff>
{:else}
	<div class="bg-card rounded-xl border px-4 py-3.5">
		<header class="mb-2.5 flex items-center gap-2">
			<SparklesIcon class="text-primary size-3.5" />
			<span class="text-[13.5px] font-semibold">Catch me up — INC-2481</span>
			<Badge tone="brand" size="sm">generated 09:52 UTC</Badge>
			<div class="flex-1"></div>
			<Button variant="ghost" size="icon-sm" aria-label="Regenerate summary" onclick={regenerate}>
				<RotateCwIcon />
			</Button>
			<Button
				variant="ghost"
				size="icon-sm"
				aria-label="Discard summary"
				onclick={() => toast('Summary discarded. Nothing was saved to the timeline.')}
			>
				<XIcon />
			</Button>
		</header>
		{#if loading}
			<div class="flex items-center gap-2.5 py-2.5">
				<span
					class="border-border border-t-primary size-4 shrink-0 animate-spin rounded-full border-2 [animation-duration:0.8s] motion-reduce:animate-none"
					aria-hidden="true"
				></span>
				<span class="text-muted-foreground text-[12.5px]">Reading 9 timeline entries…</span>
			</div>
		{:else}
			<p class="text-muted-foreground m-0 text-[13px] leading-[1.65]">
				Checkout errors in EU spiked at 09:12 UTC, ten minutes after deploy 84c2. Failover to us-east-1
				began 09:31 UTC; error rate is down from 12% to 3.1%. EU deploys are held. Fix owner: Marcus.
				Next status update due 10:07 UTC.
			</p>
			<p class="text-subtle-foreground mt-2 mb-0 text-[10.5px]">
				Drafted by Opsybot from the timeline — verify before acting on it.
			</p>
		{/if}
	</div>
{/if}
