<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import CheckIcon from '@lucide/svelte/icons/check';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import GlobeIcon from '@lucide/svelte/icons/globe';
	import HistoryIcon from '@lucide/svelte/icons/history';
	import ListChecksIcon from '@lucide/svelte/icons/list-checks';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import Page from '$lib/components/layout/page.svelte';
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import FactorCard from '$lib/components/postmortems/factor-card.svelte';
	import PublishOption from '$lib/components/postmortems/publish-option.svelte';
	import SectionEditor from '$lib/components/postmortems/section-editor.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { REVIEWERS, stateLabel, STATE_TONE } from '$lib/postmortems';
	import { formatUtc } from '$lib/time';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	const postmortem = $derived(data.postmortem);
	const readonly = $derived(data.state === 'published');

	let reviewer = $state(REVIEWERS[0]);
	let title = $state('');
	let owner = $state('Priya Nair');
	let due = $state(new Date(Date.now() + 7 * 86_400_000).toISOString().slice(0, 10));

	$effect(() => {
		if (form?.reviewer) toast.success(`Sent to ${form.reviewer} for review.`);
	});

	$effect(() => {
		if (form?.published) {
			toast.success(
				form.announced ? 'Postmortem published and announced in the incident channel.' : 'Postmortem published.'
			);
		}
	});

	$effect(() => {
		if (form?.added) toast.success('Follow-up added — it is on the global follow-ups list.');
	});
</script>

<Page title="Postmortems" subtitle="Blameless, drafted from the timeline">
	<div class="flex max-w-[860px] flex-col gap-3.5">
		<a
			href="/postmortems"
			class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[12.5px]"
		>
			<ArrowLeftIcon class="size-3.5" />
			Library
		</a>

		<div class="flex flex-wrap items-center gap-2.5">
			<h2 class="tracking-heading m-0 text-[19px] font-semibold">{data.title}</h2>
			<span class="text-subtle-foreground font-mono text-xs">
				{postmortem.id} · {data.incidentId}
			</span>
			<Badge tone={STATE_TONE[data.state]}>{stateLabel(data.state)}</Badge>

			<div class="flex-1"></div>

			{#if data.state === 'draft' || data.state === 'not-started'}
				<form method="POST" action="?/review" use:enhance class="flex items-center gap-2">
					<Select.Root type="single" name="reviewer" bind:value={reviewer}>
						<Select.Trigger size="sm" class="w-[150px]">{reviewer}</Select.Trigger>
						<Select.Content>
							<Select.Group>
								{#each REVIEWERS as person (person)}
									<Select.Item value={person} label={person}>{person}</Select.Item>
								{/each}
							</Select.Group>
						</Select.Content>
					</Select.Root>
					<Button type="submit" size="sm">
						<EyeIcon data-icon="inline-start" />
						Request review
					</Button>
				</form>
			{:else if data.state === 'in-review'}
				<form method="POST" action="?/publish" use:enhance>
					<Button type="submit" size="sm">
						<CheckIcon data-icon="inline-start" />
						Publish
					</Button>
				</form>
			{/if}

			{#if data.state === 'published' && postmortem.publicLink}
				<Button size="sm" variant="secondary" href="/postmortems/{postmortem.id}/public">
					<GlobeIcon data-icon="inline-start" />
					View public page
				</Button>
			{/if}
		</div>

		<section class="bg-card overflow-hidden rounded-xl border">
			<div class="grid grid-cols-[repeat(auto-fit,minmax(150px,1fr))] gap-3.5 px-4 py-3.5">
				{#each data.facts as fact (fact.label)}
					<div>
						<div class="text-subtle-foreground tracking-[0.07em] mb-1 text-[10.5px] uppercase">
							{fact.label}
						</div>
						<div class="text-[13px] font-medium {fact.mono ? 'font-mono' : ''}">{fact.value}</div>
					</div>
				{/each}
			</div>
			<p class="text-subtle-foreground m-0 border-t px-3.5 py-2 text-[11.5px]">
				Facts come from the incident record and cannot drift from it.
			</p>
		</section>

		<SectionEditor
			section="summary"
			title="Summary"
			value={postmortem.summary}
			{readonly}
		/>
		<SectionEditor section="impact" title="Impact" value={postmortem.impact} {readonly} />

		<section class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Contributing factors</span>
				<div class="flex-1"></div>
				{#if !readonly}
					<form method="POST" action="?/addFactor" use:enhance>
						<Button type="submit" size="sm" variant="ghost">
							<PlusIcon data-icon="inline-start" />
							Add factor
						</Button>
					</form>
				{/if}
			</header>

			<p class="bg-inset text-subtle-foreground m-0 border-b px-4 py-[9px] text-xs leading-[1.55]">
				Blameless means systemic: describe conditions and mechanisms, never who. “The pipeline had
				no canary stage” — not “someone deployed without checking”.
			</p>

			<div class="flex flex-col gap-3 p-3.5">
				{#each postmortem.factors as factor (factor.id)}
					<FactorCard {factor} incidentId={data.incidentId} {readonly} />
				{:else}
					<p class="text-subtle-foreground m-0 py-2 text-center text-[12.5px]">
						No contributing factors yet. A postmortem without one has not finished asking why.
					</p>
				{/each}
			</div>
		</section>

		<SectionEditor
			section="wentWell"
			title="What went well"
			value={postmortem.wentWell}
			{readonly}
		/>
		<SectionEditor
			section="improve"
			title="What could be improved"
			value={postmortem.improve}
			{readonly}
		/>

		<section class="bg-card overflow-hidden rounded-xl border">
			<header class="flex flex-wrap items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Follow-up actions</span>
				<span class="text-subtle-foreground ml-1 text-[11.5px]">
					the same list the incident carries
				</span>
			</header>

			{#each data.followUps as followUp (followUp.id)}
				<div class="flex items-center gap-3 border-t px-4 py-3 first:border-t-0">
					<ListChecksIcon class="text-subtle-foreground size-[13px] shrink-0" />
					<span class="min-w-0 flex-1 text-[13px]">{followUp.title}</span>
					<UserAvatar name={followUp.owner} size="xs" />
					<span class="text-subtle-foreground font-mono text-[11px]">due {followUp.dueAt.slice(0, 10)}</span>
				</div>
			{/each}

			{#if !readonly}
				<form
					method="POST"
					action="?/followUp"
					use:enhance={() =>
						async ({ update }) => {
							await update();
							title = '';
						}}
					class="flex flex-wrap items-start gap-2 border-t px-3.5 py-3"
				>
					<Input
						name="title"
						bind:value={title}
						placeholder="Action that prevents a repeat"
						aria-label="Follow-up"
						class="h-[34px] min-w-[220px] flex-1 text-[13px]"
					/>

					<Select.Root type="single" name="owner" bind:value={owner}>
						<Select.Trigger size="sm" class="w-[140px]">{owner}</Select.Trigger>
						<Select.Content>
							<Select.Group>
								{#each data.people as person (person)}
									<Select.Item value={person} label={person}>{person}</Select.Item>
								{/each}
							</Select.Group>
						</Select.Content>
					</Select.Root>

					<Input
						name="due"
						type="date"
						bind:value={due}
						aria-label="Due date"
						class="h-[34px] w-[150px] text-[13px]"
					/>

					<Button type="submit" size="sm" disabled={!title.trim()}>
						<PlusIcon data-icon="inline-start" />
						Add
					</Button>
				</form>
			{/if}
		</section>

		<section class="bg-card flex flex-col gap-3 rounded-xl border p-3.5">
			<div class="text-[13.5px] font-semibold">Publish options</div>

			<PublishOption
				option="announce"
				on={postmortem.announce}
				disabled={readonly}
				label="Announce in the incident channel when published"
			>
				Announce in #{data.incidentId.toLowerCase()} when published
			</PublishOption>

			<PublishOption
				option="publicLink"
				on={postmortem.publicLink}
				label="Public link — readable without login"
				onchanged={(next) => {
					if (next) toast.warning('Public link enabled. Internal-only fields never appear on it.');
				}}
			>
				Public link — readable without login
				<span class="text-subtle-foreground">(off by default)</span>
			</PublishOption>
		</section>

		<section class="bg-card overflow-hidden rounded-xl border">
			<Collapsible.Root>
				<Collapsible.Trigger
					class="hover:bg-accent text-foreground flex w-full items-center gap-2 px-4 py-3.5 text-[13px] font-semibold"
				>
					<HistoryIcon class="size-[13px]" />
					Edit history
					<Badge tone="neutral" size="sm">{postmortem.history.length}</Badge>
				</Collapsible.Trigger>

				<Collapsible.Content>
					{#each postmortem.history as edit (edit.id)}
						<div class="flex items-center gap-3 border-t px-4 py-3">
							<UserAvatar name={edit.by} size="xs" />
							<span class="flex-1 text-[12.5px]">{edit.what}</span>
							<span class="text-subtle-foreground shrink-0 font-mono text-[10.5px]">
								{formatUtc(edit.at)} · {edit.by.split(' ')[0]}
							</span>
						</div>
					{/each}
				</Collapsible.Content>
			</Collapsible.Root>
		</section>
	</div>
</Page>
