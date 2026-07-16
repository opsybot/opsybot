<script lang="ts">
	import PlusIcon from '@lucide/svelte/icons/plus';
	import { enhance } from '$app/forms';
	import Panel from '$lib/components/incidents/panel.svelte';
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { PEOPLE } from '$lib/incidents';
	import { formatUtcDate } from '$lib/time';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let title = $state('');
	let owner = $state(PEOPLE[0]);
	let due = $state(new Date(Date.now() + 7 * 86_400_000).toISOString().slice(0, 10));

	let toggleForms: Record<string, HTMLFormElement | undefined> = $state({});

	const overdue = (followUp: { dueAt: string; done: boolean }) =>
		!followUp.done && Date.parse(followUp.dueAt) < data.now;
</script>

<div class="flex max-w-[720px] flex-col gap-3.5">
	<Panel>
		{#if data.followUps.length === 0}
			<div class="text-subtle-foreground px-4 py-6 text-center text-[13px]">
				No follow-ups yet. Capture them while the context is fresh.
			</div>
		{:else}
			{#each data.followUps as followUp (followUp.id)}
				<div class="flex items-center gap-2.5 border-t px-3.5 py-[11px] first:border-t-0">
					<form
						method="POST"
						action="?/toggle-follow-up"
						use:enhance
						bind:this={toggleForms[followUp.id]}
					>
						<input type="hidden" name="id" value={followUp.id} />
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
					<span class="font-mono text-[11px] {overdue(followUp) ? 'text-critical-ink' : 'text-subtle-foreground'}">
						due {formatUtcDate(followUp.dueAt)}
					</span>
					{#if overdue(followUp)}
						<Badge tone="critical" size="sm">overdue</Badge>
					{/if}
				</div>
			{/each}
		{/if}
	</Panel>

	<Panel class="p-3.5">
		<form
			method="POST"
			action="?/add-follow-up"
			use:enhance={() => async ({ update }) => {
				await update();
				title = '';
			}}
			class="flex flex-wrap items-start gap-2"
		>
			<Input
				name="title"
				bind:value={title}
				placeholder="What should we fix so this doesn't repeat?"
				class="h-[34px] min-w-[240px] flex-1 text-[13px]"
			/>

			<Select.Root type="single" name="owner" bind:value={owner}>
				<Select.Trigger size="sm" class="w-[140px]">{owner}</Select.Trigger>
				<Select.Content>
					<Select.Group>
						{#each PEOPLE as person (person)}
							<Select.Item value={person} label={person}>{person}</Select.Item>
						{/each}
					</Select.Group>
				</Select.Content>
			</Select.Root>

			<Input name="due" type="date" bind:value={due} class="h-[34px] w-[150px] text-[13px]" />

			<Button type="submit" size="sm" disabled={!title.trim()}>
				<PlusIcon data-icon="inline-start" />
				Add
			</Button>
		</form>
	</Panel>
</div>
