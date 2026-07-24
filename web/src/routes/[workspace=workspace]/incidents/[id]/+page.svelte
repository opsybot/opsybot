<script lang="ts">
	import LinkIcon from '@lucide/svelte/icons/link';
	import BookOpenIcon from '@lucide/svelte/icons/book-open';
	import XIcon from '@lucide/svelte/icons/x';
	import { enhance } from '$app/forms';
	import Panel from '$lib/components/incidents/panel.svelte';
	import StatusBadge from '$lib/components/status-badge.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { RUNBOOKS } from '$lib/incidents';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const incident = $derived(data.incident);
	const RELATIONS = ['caused by', 'duplicates', 'related to'];

	let relation = $state('caused by');
	let other = $state('');

	const otherName = $derived(
		data.candidates.find((candidate) => candidate.id === other)?.name ?? 'Incident'
	);

	const fieldValue = (id: string) => incident.customFieldsRaw?.[id] ?? '';
	const fieldValues = (id: string) =>
		fieldValue(id)
			.split(',')
			.map((entry) => entry.trim())
			.filter(Boolean);
	const fieldOptions = (options?: string) =>
		(options ?? '')
			.split(',')
			.map((entry) => entry.trim())
			.filter(Boolean);
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
					<form method="POST" action="?/unlink-alert" use:enhance>
						<input type="hidden" name="alert" value={alert.id} />
						<Button type="submit" variant="ghost" size="icon-sm" aria-label="Unlink alert">
							<XIcon />
						</Button>
					</form>
				</div>
			{/each}

			{#if incident.alerts.length === 0}
				<div class="text-subtle-foreground border-t px-3.5 py-2.5 text-[11.5px]">
					No alerts linked yet. Link one from an alert's page.
				</div>
			{/if}

			<div class="text-subtle-foreground border-t px-3.5 py-2.5 text-[11.5px]">
				Resolving the incident offers to resolve these together.
			</div>
		</Panel>

		<Panel title="Related incidents">
			{#each incident.related as related (related.relationId ?? related.id)}
				<div class="flex items-center gap-2.5 border-t px-3.5 py-[11px] first:border-t-0">
					<LinkIcon class="text-subtle-foreground size-3.5 shrink-0" />
					<span class="flex-1 text-[13px]">
						<Badge tone="neutral" size="sm">{related.relation}</Badge>
						<span class="text-foreground font-mono text-xs">{related.id}</span>: {related.name}
					</span>
					{#if related.relationId}
						<form method="POST" action="?/unrelate" use:enhance>
							<input type="hidden" name="relation" value={related.relationId} />
							<Button type="submit" variant="ghost" size="icon-sm" aria-label="Remove relation">
								<XIcon />
							</Button>
						</form>
					{/if}
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
					<Select.Trigger size="sm" class="w-[180px]">{otherName}</Select.Trigger>
					<Select.Content>
						<Select.Group>
							{#each data.candidates as candidate (candidate.id)}
								<Select.Item value={candidate.id} label={candidate.name}>{candidate.name}</Select.Item>
							{/each}
						</Select.Group>
					</Select.Content>
				</Select.Root>

				<Button type="submit" size="sm" variant="secondary" disabled={!other}>Link</Button>
			</form>
		</Panel>
	</div>

	<div class="flex flex-col gap-3.5">
		{#if data.fieldDefs.length}
			<Panel title="Custom fields">
				<form
					method="POST"
					action="?/custom-fields"
					use:enhance
					class="flex flex-col gap-3 p-3.5"
				>
					{#each data.fieldDefs as def (def.id)}
						<div class="flex flex-col gap-1.5">
							<span class="text-subtle-foreground text-[12px] font-medium">{def.name}</span>
							{#if def.type === 'select'}
								<select
									name="cf:{def.id}"
									class="border-border-strong bg-inset h-[34px] rounded-md border px-2.5 text-[13px]"
								>
									<option value="" selected={!fieldValue(def.id)}>—</option>
									{#each fieldOptions(def.options) as opt (opt)}
										<option value={opt} selected={fieldValue(def.id) === opt}>{opt}</option>
									{/each}
								</select>
							{:else if def.type === 'multi-select'}
								<div class="flex flex-wrap gap-2">
									{#each fieldOptions(def.options) as opt (opt)}
										<label class="text-foreground flex items-center gap-1.5 text-[12.5px]">
											<input
												type="checkbox"
												name="cf:{def.id}"
												value={opt}
												checked={fieldValues(def.id).includes(opt)}
											/>
											{opt}
										</label>
									{/each}
								</div>
							{:else}
								<Input
									name="cf:{def.id}"
									type={def.type === 'number' ? 'number' : 'text'}
									value={fieldValue(def.id)}
									class="h-[34px] text-[13px]"
								/>
							{/if}
						</div>
					{/each}
					<Button type="submit" size="sm" variant="secondary" class="self-start">Save fields</Button>
				</form>
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
