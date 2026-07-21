<script lang="ts">
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import SparklesIcon from '@lucide/svelte/icons/sparkles';
	import { enhance } from '$app/forms';
	import Panel from '$lib/components/incidents/panel.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { ws } from '$lib/navigation';
	import { postmortemId } from '$lib/postmortems';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const incident = $derived(data.incident);
	const stage = $derived(incident.postmortem);

	let drafting = $state(false);
</script>

<div class="flex max-w-[720px] flex-col gap-3.5">
	{#if drafting}
		<Panel class="flex items-center gap-3 p-5">
			<span
				class="border-border border-t-primary motion-safe:animate-spin size-4 shrink-0 rounded-full border-2"
				aria-hidden="true"
			></span>
			<span class="text-muted-foreground text-[13px]">
				Opsybot is drafting from {incident.timeline.length} timeline entries and {incident.alerts
					.length} linked alerts…
			</span>
		</Panel>
	{:else if stage === 'not-started'}
		<div class="bg-card flex flex-col items-center gap-2.5 rounded-xl border px-5 py-10">
			<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
				<FileTextIcon class="text-subtle-foreground size-[18px]" />
			</span>
			<div class="text-sm font-medium">No postmortem yet</div>
			<p class="text-subtle-foreground m-0 max-w-[420px] text-center text-[12.5px] leading-[1.55]">
				{incident.status === 'resolved'
					? 'Start from the timeline: most of the story is already captured.'
					: 'Usually written after resolve, but you can start any time.'}
			</p>
			<div class="flex gap-2">
				<form
					method="POST"
					action="?/postmortem"
					use:enhance={() => {
						drafting = true;
						return async ({ update }) => {
							await new Promise((resolve) => setTimeout(resolve, 1600));
							await update();
							drafting = false;
						};
					}}
				>
					<input type="hidden" name="state" value="draft" />
					<Button type="submit" size="sm">
						<SparklesIcon data-icon="inline-start" />
						Draft from timeline
					</Button>
				</form>

				<form method="POST" action="?/postmortem" use:enhance>
					<input type="hidden" name="state" value="draft" />
					<Button type="submit" size="sm" variant="secondary">Start blank</Button>
				</form>
			</div>
		</div>
	{:else}
		<Panel>
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<FileTextIcon class="text-subtle-foreground size-3.5" />
				<span class="text-[13.5px] font-semibold">Postmortem: {incident.id}</span>
				<Badge
					tone={stage === 'published' ? 'success' : stage === 'in-review' ? 'info' : 'neutral'}
					size="sm"
				>
					{stage === 'in-review' ? 'in review' : stage}
				</Badge>
				<div class="flex-1"></div>

				{#if stage === 'draft'}
					<form method="POST" action="?/postmortem" use:enhance>
						<input type="hidden" name="state" value="in-review" />
						<Button type="submit" size="sm" variant="secondary">Request review</Button>
					</form>
				{:else if stage === 'in-review'}
					<form method="POST" action="?/postmortem" use:enhance>
						<input type="hidden" name="state" value="published" />
						<Button type="submit" size="sm">Publish</Button>
					</form>
				{/if}
			</header>

			<div class="p-4">
				<p class="text-muted-foreground m-0 text-[13px] leading-[1.6]">
					{incident.summary ||
						'The narrative sections are drafted from the timeline, so the author starts editing rather than staring at a blank page.'}
				</p>
				<p class="text-subtle-foreground mt-2.5 mb-0 text-[11px]">
					Drafted by Opsybot. Review before sharing. Timeline, contributing factors, and follow-ups
					are pre-filled.
				</p>
				<Button variant="ghost" size="sm" class="mt-3" href={ws(`/postmortems/${postmortemId(incident.id)}`)}>
					<PencilIcon data-icon="inline-start" />
					Open editor
				</Button>
			</div>
		</Panel>
	{/if}
</div>
