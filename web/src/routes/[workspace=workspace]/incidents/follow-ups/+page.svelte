<script lang="ts">
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import IncidentsTabs from '$lib/components/incidents/incidents-tabs.svelte';
	import Panel from '$lib/components/incidents/panel.svelte';
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import Page from '$lib/components/layout/page.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Select from '$lib/components/ui/select';
	import { ws } from '$lib/navigation';
	import { formatUtcDate } from '$lib/time';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const params = $derived(page.url.searchParams);
	const owner = $derived(params.get('owner') ?? '');
	const stateFilter = $derived(params.get('state') ?? '');
	const overdueOnly = $derived(stateFilter === 'overdue');

	let toggleForms: Record<string, HTMLFormElement | undefined> = $state({});

	function set(key: string, value: string | undefined) {
		const next = new URLSearchParams(params);
		if (value) next.set(key, value);
		else next.delete(key);
		goto(`?${next}`, { keepFocus: true, noScroll: true, replaceState: true });
	}

	const overdue = (followUp: { dueAt: string; done: boolean }) =>
		!followUp.done && Date.parse(followUp.dueAt) < data.now;

	const groups = $derived([...new Set(data.followUps.map((followUp) => followUp.incidentId))]);
</script>

<Page title="Incidents" subtitle="From alert to postmortem">
	<IncidentsTabs current="follow-ups" />

	<div class="flex max-w-[780px] flex-col gap-3.5">
		<div class="bg-card flex flex-wrap items-center gap-2.5 rounded-xl border px-3 py-2.5">
			<Select.Root
				type="single"
				value={owner}
				onValueChange={(value) => set('owner', value || undefined)}
			>
				<Select.Trigger size="sm" class="w-[140px]">{owner || 'Owner'}</Select.Trigger>
				<Select.Content>
					<Select.Group>
						{#each data.people as person (person)}
							<Select.Item value={person} label={person}>{person}</Select.Item>
						{/each}
					</Select.Group>
				</Select.Content>
			</Select.Root>

			<Select.Root
				type="single"
				value={stateFilter}
				onValueChange={(value) => set('state', value || undefined)}
			>
				<Select.Trigger size="sm" class="w-[120px]">{stateFilter || 'State'}</Select.Trigger>
				<Select.Content>
					<Select.Group>
						{#each ['open', 'done', 'overdue'] as option (option)}
							<Select.Item value={option} label={option}>{option}</Select.Item>
						{/each}
					</Select.Group>
				</Select.Content>
			</Select.Root>

			<Tag
				selected={overdueOnly}
				onclick={() => set('state', overdueOnly ? undefined : 'overdue')}
			>
				Overdue only
			</Tag>

			{#if owner || stateFilter}
				<button
					type="button"
					onclick={() => goto(ws('/incidents/follow-ups'), { replaceState: true })}
					class="text-muted-foreground hover:text-brand-foreground text-[12.5px]"
				>
					Clear
				</button>
			{/if}

			<div class="flex-1"></div>
			<span class="text-subtle-foreground text-[12.5px]">
				{data.followUps.length} items · grouped by incident
			</span>
		</div>

		{#each groups as group (group)}
			{@const incident = data.incidents.find((entry) => entry.id === group)}
			<Panel>
				<header class="flex items-center gap-2 border-b px-4 py-3">
					<a href={ws(`/incidents/${group}`)} class="text-foreground font-mono text-[12.5px] font-semibold">
						{incident?.ref ?? group}
					</a>
					<span class="text-subtle-foreground text-[12.5px]">{incident?.name}</span>
				</header>

				{#each data.followUps.filter((followUp) => followUp.incidentId === group) as followUp (followUp.id)}
					<div class="flex items-center gap-2.5 border-t px-3.5 py-[11px]">
						<form method="POST" action="?/toggle" use:enhance bind:this={toggleForms[followUp.id]}>
							<input type="hidden" name="id" value={followUp.id} />
							<input type="hidden" name="incident" value={followUp.incidentId} />
							<input type="hidden" name="done" value={!followUp.done} />
							<Checkbox
								checked={followUp.done}
								onCheckedChange={() => toggleForms[followUp.id]?.requestSubmit()}
								aria-label={followUp.title}
							/>
						</form>

						<span
							class="min-w-0 flex-1 text-[13.5px] {followUp.done
								? 'text-subtle-foreground line-through'
								: 'text-foreground'}"
						>
							{followUp.title}
						</span>

						<UserAvatar name={followUp.owner} size="xs" />
						<span
							class="font-mono text-[11px] {overdue(followUp)
								? 'text-critical-ink'
								: 'text-subtle-foreground'}"
						>
							due {formatUtcDate(followUp.dueAt)}
						</span>
						{#if overdue(followUp)}
							<Badge tone="critical" size="sm">overdue</Badge>
						{/if}
					</div>
				{/each}
			</Panel>
		{:else}
			<div class="bg-card flex flex-col items-center gap-2.5 rounded-xl border px-5 py-9">
				<div class="text-sm font-medium">Nothing matches</div>
			</div>
		{/each}
	</div>
</Page>
