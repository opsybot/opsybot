<script lang="ts">
	import { tick } from 'svelte';
	import CheckIcon from '@lucide/svelte/icons/check';
	import SparklesIcon from '@lucide/svelte/icons/sparkles';
	import { toast } from 'svelte-sonner';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import SurfaceOff from '$lib/components/ai/surface-off.svelte';

	let { enabled }: { enabled: boolean } = $props();

	let dismissed = $state(false);
	let strip = $state<HTMLElement | null>(null);
	async function dismiss() {
		dismissed = true;
		await tick();
		strip?.focus();
	}

	const ALERTS = [
		'payments-api p99 above 800 ms',
		'Synthetic checkout failing from us-east-1',
		'db-3 connection pool saturated',
		'events-worker queue depth above 10k'
	];
</script>

{#if !enabled}
	<SurfaceOff>Correlation suggestions need a model — currently unavailable.</SurfaceOff>
{:else if dismissed}
	<div
		bind:this={strip}
		tabindex="-1"
		role="status"
		class="border-border-strong bg-inset text-subtle-foreground flex items-center gap-2.5 rounded-xl border px-4 py-3 text-[12.5px] outline-none"
	>
		<CheckIcon class="size-3.5 shrink-0" />
		<span>Suggestion dismissed. It won't reappear for these alerts.</span>
	</div>
{:else}
	<div class="bg-card border-brand-edge rounded-xl border px-4 py-3.5">
		<header class="mb-2 flex items-center gap-2">
			<SparklesIcon class="text-primary size-3.5" />
			<span class="text-[13.5px] font-semibold">These 4 alerts may be related</span>
			<Badge tone="brand" size="sm">confidence: high</Badge>
		</header>
		<p class="text-muted-foreground mb-2 text-[12.5px] leading-[1.55]">
			Shared service <span class="text-foreground font-mono text-[12px]">payments-api</span> and its dependency
			<span class="text-foreground font-mono text-[12px]">database</span> · all fired within 6 minutes of
			each other.
		</p>
		<div class="mb-2.5 flex flex-col gap-1">
			{#each ALERTS as alert (alert)}
				<span class="text-muted-foreground inline-flex items-center gap-[7px] text-[12px]">
					<span class="size-[5px] shrink-0 rounded-full bg-[var(--warning)]"></span>
					{alert}
				</span>
			{/each}
		</div>
		<div class="flex gap-2">
			<Button
				size="sm"
				onclick={() => toast.success('4 alerts merged into one group and attached to INC-2481.')}
			>
				Merge & attach to INC-2481
			</Button>
			<Button size="sm" variant="ghost" onclick={dismiss}>Dismiss</Button>
		</div>
	</div>
{/if}
