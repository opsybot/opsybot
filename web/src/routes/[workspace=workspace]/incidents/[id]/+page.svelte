<script lang="ts">
	import LinkIcon from '@lucide/svelte/icons/link';
	import BookOpenIcon from '@lucide/svelte/icons/book-open';
	import { enhance } from '$app/forms';
	import Panel from '$lib/components/incidents/panel.svelte';
	import StatusBadge from '$lib/components/status-badge.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Select from '$lib/components/ui/select';
	import { RUNBOOKS } from '$lib/incidents';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const incident = $derived(data.incident);
	const RELATIONS = ['caused by', 'duplicates', 'related to'];

	let relation = $state('caused by');
	let other = $state('');
</script>

<div class="grid items-start gap-3.5 min-[1100px]:grid-cols-[minmax(0,1fr)_300px]">
	<div class="flex min-w-0 flex-col gap-3.5">
		{#if incident.summary}
			<Panel class="p-4">
				<div class="mb-2.5 flex items-center gap-2">
					<span class="text-[13.5px] font-semibold">Summary</span>
					<Badge tone="brand" size="sm">kept current by Opsybot</Badge>
				</div>
				<p class="text-muted-foreground m-0 text-[13.5px] leading-[1.65]">{incident.summary}</p>
				<p class="text-subtle-foreground mt-2.5 mb-0 text-[11px]">
					Drafted by Opsybot from the timeline. Review before sharing.
				</p>
			</Panel>
		{/if}

		<Panel title="Linked alerts">
			{#snippet aside()}
				<Badge tone="neutral" size="sm">{incident.alerts.length}</Badge>
			{/snippet}

			{#each incident.alerts as alert (alert.id)}
				<div class="flex items-center gap-2.5 border-t px-3.5 py-[11px] first:border-t-0">
					<Badge tone={alert.tone} size="sm">{alert.severity}</Badge>
					<span class="min-w-0 flex-1 text-[13px]">{alert.title}</span>
					{#if alert.status === 'resolved'}
						<Badge tone="success" size="sm">resolved</Badge>
					{:else}
						<StatusBadge
							status={alert.status === 'open' ? 'firing' : 'acknowledged'}
							label={alert.status}
							size="sm"
						/>
					{/if}
				</div>
			{/each}

			<div class="text-subtle-foreground border-t px-3.5 py-2.5 text-[11.5px]">
				Resolving the incident offers to resolve these together.
			</div>
		</Panel>

		<Panel title="Related incidents">
			{#each incident.related as related (related.id)}
				<div class="flex items-center gap-2.5 border-t px-3.5 py-[11px] first:border-t-0">
					<LinkIcon class="text-subtle-foreground size-3.5 shrink-0" />
					<span class="flex-1 text-[13px]">
						<Badge tone="neutral" size="sm">{related.relation}</Badge>
						<span class="text-foreground font-mono text-xs">{related.id}</span>: {related.name}
					</span>
				</div>
			{/each}

			<form
				method="POST"
				action="?/link"
				use:enhance
				class="flex flex-wrap gap-2 border-t px-3.5 py-2.5"
			>
				<Select.Root type="single" name="relation" bind:value={relation}>
					<Select.Trigger size="sm" class="w-[120px]">{relation}</Select.Trigger>
					<Select.Content>
						<Select.Group>
							{#each RELATIONS as option (option)}
								<Select.Item value={option} label={option}>{option}</Select.Item>
							{/each}
						</Select.Group>
					</Select.Content>
				</Select.Root>

				<Select.Root type="single" name="incident" bind:value={other}>
					<Select.Trigger size="sm" class="w-[130px]">{other || 'Incident'}</Select.Trigger>
					<Select.Content>
						<Select.Group>
							{#each data.candidates as candidate (candidate.id)}
								<Select.Item value={candidate.id} label={candidate.id}>{candidate.id}</Select.Item>
							{/each}
						</Select.Group>
					</Select.Content>
				</Select.Root>

				<Button type="submit" size="sm" variant="secondary" disabled={!other}>Link</Button>
			</form>
		</Panel>
	</div>

	<div class="flex flex-col gap-3.5">
		{#if incident.customFields.length}
			<Panel title="Custom fields">
				{#each incident.customFields as field (field.label)}
					<div class="flex justify-between gap-3 border-t px-3.5 py-2.5 text-[12.5px] first:border-t-0">
						<span class="text-subtle-foreground">{field.label}</span>
						<span class="text-foreground text-right {field.mono ? 'font-mono text-xs' : ''}">
							{field.value}
						</span>
					</div>
				{/each}
			</Panel>
		{/if}

		<Panel title="Runbooks">
			{#snippet aside()}
				<span class="text-subtle-foreground ml-1 text-[11.5px]">for affected services</span>
			{/snippet}

			{#each RUNBOOKS.filter((book) => incident.services.includes(book.service)) as book (book.label)}
				<div class="flex items-center gap-2.5 border-t px-3.5 py-[11px] first:border-t-0">
					<BookOpenIcon class="text-subtle-foreground size-3.5 shrink-0" />
					<a href={ws('/catalog')} class="text-brand-foreground flex-1 text-[12.5px] hover:underline">
						{book.label}
					</a>
					<span class="text-subtle-foreground font-mono text-[10.5px]">{book.service}</span>
				</div>
			{/each}
		</Panel>
	</div>
</div>
