<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import EyeOffIcon from '@lucide/svelte/icons/eye-off';
	import RotateCwIcon from '@lucide/svelte/icons/rotate-cw';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import CopyField from '$lib/components/alertsources/copy-field.svelte';
	import MappingTable from '$lib/components/alertsources/mapping-table.svelte';
	import SourceToggle from '$lib/components/alertsources/source-toggle.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { ICON } from '$lib/components/alertsources/icons';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import { endpointUrl, healthBadge } from '$lib/alertsources';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const source = $derived(data.source);
	const Icon = $derived(ICON[source.icon]);
	const url = $derived(endpointUrl(source.slug));
	const health = $derived(healthBadge(source));

	let revealed = $state(false);
	let rotateOpen = $state(false);

	// Reseed only when the source id changes so a save does not discard an in-progress edit
	let rows = $state(untrack(() => data.source.mapping.map((m) => ({ ...m }))));
	$effect(() => {
		source.id;
		untrack(() => (rows = source.mapping.map((m) => ({ ...m }))));
	});
	const dirty = $derived(JSON.stringify(rows) !== JSON.stringify(source.mapping));
	const mappingJson = $derived(JSON.stringify(rows));
</script>

<div class="flex flex-col gap-3.5">
	<a
		href="/alert-sources"
		class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[12.5px] transition-colors"
	>
		<ArrowLeftIcon class="size-3.5" />
		Alert sources
	</a>

	<div class="flex flex-wrap items-center gap-2.5">
		<span
			class="bg-inset text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-sm border"
		>
			<Icon class="size-4" />
		</span>
		<h2 class="font-mono text-[18px] font-semibold">{source.name}</h2>
		<Tag>{source.format}</Tag>
		<Badge tone={health.tone} size="sm" dot>{health.label}</Badge>
		<div class="flex-1"></div>
		<SourceToggle id={source.id} paused={source.status === 'paused'} />
	</div>

	<div class="grid items-start gap-3.5 min-[1100px]:[grid-template-columns:minmax(0,1fr)_320px]">
		<div class="flex min-w-0 flex-col gap-3.5">
			<div class="bg-card flex flex-col gap-3.5 rounded-xl border p-4">
				<CopyField label="Endpoint URL" value={url} />
				<div>
					<div class="text-subtle-foreground mb-[7px] text-[11px] tracking-[0.08em] uppercase">
						Signing secret
					</div>
					<div class="flex items-center gap-2">
						<code
							class="bg-inset text-foreground flex-1 rounded-md border px-[11px] py-[9px] font-mono text-[12px] [overflow-wrap:anywhere]"
						>
							{revealed ? source.secret : source.secret.slice(0, 6) + '••••••••••••'}
						</code>
						<Button variant="ghost" size="sm" onclick={() => (revealed = !revealed)}>
							{#if revealed}
								<EyeOffIcon data-icon="inline-start" />
								Hide
							{:else}
								<EyeIcon data-icon="inline-start" />
								Reveal
							{/if}
						</Button>
						<Button variant="secondary" size="sm" onclick={() => (rotateOpen = true)}>
							<RotateCwIcon data-icon="inline-start" />
							Rotate
						</Button>
					</div>
				</div>
			</div>

			<section class="bg-card overflow-hidden rounded-xl border">
				<header class="flex items-center gap-2 border-b px-4 py-3">
					<span class="text-[14px] font-semibold">Recent events</span>
				</header>
				<table class="w-full border-collapse text-[13px]">
					<thead>
						<tr class="text-subtle-foreground text-left text-[11px] tracking-[0.05em] uppercase">
							<th class="py-[9px] pr-[10px] pl-4 font-semibold">Received</th>
							<th class="px-[10px] py-[9px] font-semibold">Event</th>
							<th class="py-[9px] pr-4 pl-[10px] font-semibold">Outcome</th>
						</tr>
					</thead>
					<tbody>
						{#each data.events as event (event.at)}
							<tr class="border-t">
								<td class="text-subtle-foreground py-[10px] pr-[10px] pl-4 font-mono whitespace-nowrap">
									{event.at}
								</td>
								<td class="px-[10px] py-[10px] {event.tone === 'critical' ? 'text-subtle-foreground' : ''}">
									{event.title}
								</td>
								<td class="py-[10px] pr-4 pl-[10px]">
									<Badge tone={event.tone} size="sm">{event.outcome}</Badge>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</section>

			{#if source.failures}
				<div
					class="bg-warning-wash border-warning-edge flex items-center gap-2.5 rounded-md border px-[14px] py-2.5"
				>
					<TriangleAlertIcon class="text-warning-ink size-3.5 shrink-0" />
					<span class="text-muted-foreground flex-1 text-[12.5px]">
						{source.failures} payloads from this source failed to parse in the last 24 h.
					</span>
					<Button variant="ghost" size="sm" href="/alerts/failures">View ingestion failures</Button>
				</div>
			{/if}
		</div>

		<section class="bg-card self-start overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Field mapping</span>
				{#if dirty}
					<form
						method="POST"
						action="?/saveMapping"
						class="ml-auto"
						use:enhance={() =>
							async ({ result, update }) => {
								await update({ reset: false });
								if (result.type === 'success') toast.success('Mapping saved. Applies to the next event.');
							}}
					>
						<input type="hidden" name="mapping" value={mappingJson} />
						<Button type="submit" size="sm">Save</Button>
					</form>
				{/if}
			</header>
			<MappingTable bind:rows editable flush />
		</section>
	</div>
</div>

<Dialog.Root bind:open={rotateOpen}>
	<Dialog.Content class="sm:max-w-[440px]">
		<form
			method="POST"
			action="?/rotate"
			use:enhance={() =>
				async ({ result, update }) => {
					await update({ reset: false });
					if (result.type !== 'success') return;
					revealed = true;
					rotateOpen = false;
					toast.success('Secret rotated', {
						description: 'The old secret stops working in 24 h — update your sender before then.'
					});
				}}
		>
			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-warning-wash text-warning-ink flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						<RotateCwIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">
							Rotate the signing secret?
						</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							A new secret is generated immediately. The old one keeps working for 24 h so you can
							update the sender without dropping events.
						</Dialog.Description>
					</div>
				</div>
			</div>
			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (rotateOpen = false)}>Cancel</Button>
				<Button type="submit">Rotate secret</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
