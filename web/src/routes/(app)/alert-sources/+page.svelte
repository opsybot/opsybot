<script lang="ts">
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import PlugIcon from '@lucide/svelte/icons/plug';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import Sparkline from '$lib/components/sparkline.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { ICON } from '$lib/components/alertsources/icons';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { healthBadge } from '$lib/alertsources';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<div class="flex flex-col gap-3.5">
	<div class="flex items-center">
		<span class="text-subtle-foreground text-[13px]">
			{data.sources.length} connected {data.sources.length === 1 ? 'source' : 'sources'}
		</span>
		<div class="flex-1"></div>
		<Button size="sm" href="/alert-sources/new">
			<PlusIcon data-icon="inline-start" />
			Add integration
		</Button>
	</div>

	{#if data.sources.length === 0}
		<div
			class="text-muted-foreground flex flex-col items-center gap-2.5 rounded-xl border border-dashed px-5 py-14"
		>
			<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
				<PlugIcon class="text-subtle-foreground size-5" />
			</span>
			<div class="text-[15px] font-medium">No alert sources yet</div>
			<p class="text-subtle-foreground m-0 max-w-[420px] text-center text-[13px] leading-[1.55]">
				Point your monitoring at Opsybot and alerts start flowing. Prometheus, Grafana, Uptime Kuma,
				heartbeats, or any tool that can POST JSON.
			</p>
			<Button size="sm" variant="secondary" href="/alert-sources/new">
				<PlusIcon data-icon="inline-start" />
				Connect your first source
			</Button>
		</div>
	{:else}
		<div class="bg-card overflow-hidden rounded-xl border">
			{#each data.sources as source (source.id)}
				{@const Icon = ICON[source.icon]}
				{@const health = healthBadge(source)}
				<a
					href="/alert-sources/{source.id}"
					data-source={source.id}
					class="hover:bg-accent flex items-center gap-[14px] border-t px-4 py-[13px] first:border-t-0"
				>
					<span
						class="bg-inset text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-sm border"
					>
						<Icon class="size-4" />
					</span>
					<div class="min-w-0 flex-1">
						<div class="flex flex-wrap items-center gap-2">
							<span class="font-mono text-[13.5px] font-medium">{source.name}</span>
							<Tag>{source.format}</Tag>
							{#if source.status === 'paused'}
								<Badge tone="neutral" size="sm">paused</Badge>
							{/if}
						</div>
						<div class="text-subtle-foreground mt-[3px] font-mono text-[11px]">
							last event {source.lastEvent}
						</div>
					</div>

					<Badge tone={health.tone} size="sm" dot>{health.label}</Badge>

					<div class="flex w-[130px] items-center justify-end gap-2" title="Event volume, last 24 h">
						<div class="w-24">
							<Sparkline
								data={source.volume}
								tone={source.health === 'failing' ? 'critical' : 'brand'}
								height={22}
							/>
						</div>
						<span class="text-subtle-foreground font-mono text-[10px]">24 h</span>
					</div>

					<ChevronRightIcon class="text-subtle-foreground size-4 shrink-0" />
				</a>
			{/each}
		</div>
	{/if}
</div>
