<script lang="ts">
	import BellOffIcon from '@lucide/svelte/icons/bell-off';
	import CheckIcon from '@lucide/svelte/icons/check';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import HistoryIcon from '@lucide/svelte/icons/history';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import AlertsTabs from '$lib/components/alerts/alerts-tabs.svelte';
	import CreateSilenceDialog from '$lib/components/alerts/create-silence-dialog.svelte';
	import Page from '$lib/components/layout/page.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import { formatRemaining, formatSpan, formatUtc, formatUtcTime } from '$lib/time';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let creating = $state(false);

	$effect(() => {
		if (!page.url.searchParams.has('source')) return;
		creating = true;
		const url = new URL(page.url);
		url.searchParams.delete('source');
		goto(url, { replaceState: true, noScroll: true, keepFocus: true });
	});

	function timing(silence: (typeof data.silences)[number]): string {
		const runs = formatRemaining(Date.parse(silence.endsAt) - Date.parse(silence.startsAt));

		return silence.state === 'active'
			? `ends in ${formatRemaining(Date.parse(silence.endsAt) - data.now)} · ${formatUtcTime(silence.endsAt)}`
			: `starts ${formatUtc(silence.startsAt)} · ${runs}`;
	}
</script>

<Page title="Alerts" subtitle="Deduplicated signals from every connected source">
	<AlertsTabs current="silences" />

	<div class="flex flex-col gap-3.5">
		<section class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2.5 border-b px-4 py-3">
				<span class="text-sm font-semibold">Active &amp; scheduled</span>
				<Badge tone="neutral" size="sm">{data.silences.length}</Badge>
				<div class="flex-1"></div>
				<Button size="sm" onclick={() => (creating = true)}>
					<BellOffIcon data-icon="inline-start" />
					Create silence
				</Button>
			</header>

			{#each data.silences as silence (silence.id)}
				<div class="flex items-start gap-3 border-t px-4 py-[13px] first:border-t-0">
					<span
						class="flex size-7 shrink-0 items-center justify-center rounded-full border {silence.state ===
						'active'
							? 'bg-brand-wash border-brand-edge text-brand-foreground'
							: 'bg-inset text-subtle-foreground'}"
					>
						{#if silence.state === 'active'}
							<BellOffIcon class="size-3.5" />
						{:else}
							<ClockIcon class="size-3.5" />
						{/if}
					</span>

					<div class="min-w-0 flex-1">
						<div class="flex flex-wrap items-center gap-1.5">
							{#each silence.scope as condition (condition)}
								<Tag>{condition}</Tag>
							{/each}
							{#if silence.state === 'scheduled'}
								<Badge tone="info" size="sm">scheduled</Badge>
							{:else}
								<Badge tone="brand" size="sm" dot>active</Badge>
							{/if}
						</div>
						<div class="text-muted-foreground mt-[5px] text-[12.5px]">{silence.reason}</div>
						<div class="text-subtle-foreground mt-0.5 font-mono text-[11px]">
							{timing(silence)} · by {silence.creator}
						</div>
					</div>

					<form method="POST" action="?/end" use:enhance>
						<input type="hidden" name="id" value={silence.id} />
						<Button type="submit" variant="ghost" size="sm">
							{silence.state === 'active' ? 'End now' : 'Cancel'}
						</Button>
					</form>
				</div>
			{:else}
				<div class="flex flex-col items-center gap-2.5 px-5 py-8">
					<div class="text-sm font-medium">No silences</div>
					<p class="text-subtle-foreground m-0 text-[12.5px]">Everything that fires will page.</p>
				</div>
			{/each}
		</section>

		<section class="bg-card overflow-hidden rounded-xl border">
			<Collapsible.Root>
				<Collapsible.Trigger
					class="hover:bg-accent text-foreground flex w-full items-center gap-2 px-4 py-3.5 text-[13.5px] font-semibold"
				>
					<HistoryIcon class="size-3.5" />
					History
					<Badge tone="neutral" size="sm">{data.history.length}</Badge>
				</Collapsible.Trigger>

				<Collapsible.Content>
					{#each data.history as silence (silence.id)}
						<div class="flex items-start gap-3 border-t px-4 py-[13px] opacity-75">
							<span
								class="bg-inset text-subtle-foreground flex size-7 shrink-0 items-center justify-center rounded-full border"
							>
								<CheckIcon class="size-3.5" />
							</span>
							<div class="min-w-0 flex-1">
								<div class="flex flex-wrap gap-1.5">
									{#each silence.scope as condition (condition)}
										<Tag>{condition}</Tag>
									{/each}
								</div>
								<div class="text-subtle-foreground mt-1 font-mono text-[11px]">
									{formatSpan(silence.startsAt, silence.endsAt)} · by {silence.creator} · {silence.reason}
								</div>
							</div>
						</div>
					{/each}
				</Collapsible.Content>
			</Collapsible.Root>
		</section>
	</div>
</Page>

<CreateSilenceDialog bind:open={creating} source={data.source} />
