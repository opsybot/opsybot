<script lang="ts">
	import { toast } from 'svelte-sonner';
	import Sparkline from '$lib/components/sparkline.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import type { Alert } from '$lib/dashboard';
	import { formatSince } from '$lib/time';
	import Card from './card.svelte';
	import QuietRow from './quiet-row.svelte';

	let {
		alerts: incoming,
		volume,
		now
	}: {
		alerts: Alert[];
		volume: number[];
		now: number;
	} = $props();

	let acknowledged = $state(new Set<string>());
	let resolved = $state(new Set<string>());

	const alerts = $derived(incoming.filter((alert) => !resolved.has(alert.id)));

	function acknowledge(id: string) {
		acknowledged.add(id);
	}

	function resolve(alert: Alert) {
		resolved.add(alert.id);
		toast.success(`Alert resolved: ${alert.title}.`);
	}
</script>

<Card
	title="Alerts needing action"
	count={alerts.length}
	countTone={alerts.length ? 'warning' : 'neutral'}
>
	{#snippet aside()}
		{#if alerts.length}
			<div class="flex items-center gap-2" title="Alert volume, last 24 h">
				<Sparkline data={volume} tone="warning" height={22} class="w-24" />
				<span class="text-subtle-foreground font-mono text-[10.5px]">24 h</span>
			</div>
		{/if}
	{/snippet}

	{#if alerts.length === 0}
		<QuietRow text="Nothing needs action right now." />
	{:else}
		<div>
			{#each alerts as alert (alert.id)}
				<div
					class="flex items-center gap-3 border-t py-[11px] pr-4 pl-[18px] first:border-t-0"
					style="box-shadow: inset 3px 0 0 var(--{alert.tone})"
				>
					<div class="min-w-0 flex-1">
						<div class="text-[13.5px] font-medium">{alert.title}</div>
						<div class="text-subtle-foreground mt-px font-mono text-[11.5px]">
							{alert.source} · fired {formatSince(now - Date.parse(alert.firedAt))}
						</div>
					</div>

					{#if acknowledged.has(alert.id)}
						<Badge tone="info" size="sm">Acked by you</Badge>
					{:else}
						<Button variant="ghost" size="sm" onclick={() => acknowledge(alert.id)}>Ack</Button>
					{/if}
					<Button variant="secondary" size="sm" onclick={() => resolve(alert)}>Resolve</Button>
				</div>
			{/each}
		</div>
	{/if}
</Card>
