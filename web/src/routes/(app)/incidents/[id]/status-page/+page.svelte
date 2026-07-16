<script lang="ts">
	import EyeOffIcon from '@lucide/svelte/icons/eye-off';
	import GlobeIcon from '@lucide/svelte/icons/globe';
	import MegaphoneIcon from '@lucide/svelte/icons/megaphone';
	import Panel from '$lib/components/incidents/panel.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { STAGE_TONE, type PublishStage } from '$lib/statuspages';
	import { formatUtc } from '$lib/time';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const statusPage = $derived(data.incident.statusPage);

	const stageTone = (stage: string) => STAGE_TONE[stage.toLowerCase() as PublishStage] ?? 'neutral';
</script>

<div class="flex max-w-[720px] flex-col gap-3.5">
	<Panel>
		<header class="flex flex-wrap items-center gap-2 border-b px-4 py-3">
			<GlobeIcon class="text-subtle-foreground size-3.5" />
			<span class="text-[13.5px] font-semibold">{statusPage.domain}</span>
			{#if statusPage.stage !== 'none'}
				<Badge tone={stageTone(statusPage.stage)} size="sm">public stage: {statusPage.stage}</Badge>
			{:else}
				<Badge tone="neutral" size="sm">nothing published</Badge>
			{/if}
			<div class="flex-1"></div>
			<Button size="sm" href="/status-pages/publish?incident={data.incident.id}">
				<MegaphoneIcon data-icon="inline-start" />
				Publish update
			</Button>
		</header>

		{#if statusPage.title}
			<div class="border-b px-4 py-2.5">
				<span class="text-subtle-foreground text-[11px]">Public title:</span>
				<span class="text-foreground text-[13px] font-medium">{statusPage.title}</span>
			</div>
		{/if}

		{#each statusPage.updates as update (update.at)}
			<div class="flex items-start gap-2.5 border-t px-3.5 py-[11px] first:border-t-0">
				<Badge tone={stageTone(update.stage)} size="sm" class="shrink-0">
					{update.stage}
				</Badge>
				<div class="min-w-0 flex-1">
					<div class="text-muted-foreground text-[13px] leading-[1.55]">{update.text}</div>
					<div class="text-subtle-foreground mt-[3px] font-mono text-[10.5px]">
						{formatUtc(update.at)}
					</div>
				</div>
			</div>
		{:else}
			<div class="text-subtle-foreground px-4 py-6 text-center text-[13px]">
				Nothing published for this incident. Publishing is always a deliberate act.
			</div>
		{/each}
	</Panel>

	<Panel class="flex items-center gap-2.5 p-3.5">
		<EyeOffIcon class="text-subtle-foreground size-3.5 shrink-0" />
		<span class="text-muted-foreground flex-1 text-[12.5px]">
			internal.acme.dev — not published for this incident.
		</span>
		<Button variant="ghost" size="sm" href="/status-pages">Publish here</Button>
	</Panel>
</div>
