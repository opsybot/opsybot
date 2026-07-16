<script lang="ts">
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import AlertsTabs from '$lib/components/alerts/alerts-tabs.svelte';
	import Page from '$lib/components/layout/page.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import * as Empty from '$lib/components/ui/empty';
	import { ws } from '$lib/navigation';
	import { formatUtc } from '$lib/time';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<Page title="Alerts" subtitle="Deduplicated signals from every connected source">
	<AlertsTabs current="failures" />

	<section class="bg-card overflow-hidden rounded-xl border">
		<header class="border-b px-4 py-3">
			<div class="text-sm font-semibold">Rejected payloads</div>
			<p class="text-subtle-foreground m-0 mt-0.5 text-[12.5px]">
				A payload that never became an alert. Fix the sender, or the alert stays invisible.
			</p>
		</header>

		{#each data.failures as failure (failure.id)}
			<Collapsible.Root class="border-t first:border-t-0">
				<Collapsible.Trigger
					class="hover:bg-accent group flex w-full items-center gap-2.5 px-4 py-3 text-left"
				>
					<ChevronRightIcon
						class="text-subtle-foreground size-3.5 shrink-0 transition-transform group-data-[state=open]:rotate-90"
					/>
					<span class="w-[104px] shrink-0 font-mono text-[12.5px] font-medium">
						{failure.source}
					</span>
					<span class="text-warning-ink min-w-0 flex-1 truncate text-[13px]">{failure.reason}</span>
					<span class="text-subtle-foreground shrink-0 font-mono text-[11px]">
						{formatUtc(failure.at)}
					</span>
				</Collapsible.Trigger>

				<Collapsible.Content>
					<div class="bg-inset border-t px-4 py-3.5">
						<pre
							class="text-muted-foreground m-0 overflow-x-auto font-mono text-[11.5px] leading-[1.6] whitespace-pre">{failure.payload}</pre>
						<div class="mt-3 flex gap-2">
							<Button
								variant="secondary"
								size="sm"
								onclick={() => navigator.clipboard.writeText(failure.payload)}
							>
								Copy payload
							</Button>
							<Button variant="ghost" size="sm" href={ws('/alert-sources')}>Open source config</Button>
						</div>
					</div>
				</Collapsible.Content>
			</Collapsible.Root>
		{:else}
			<Empty.Root class="py-10">
				<Empty.Header>
					<Empty.Title>No rejected payloads</Empty.Title>
					<Empty.Description>Every payload in the last 7 days became an alert.</Empty.Description>
				</Empty.Header>
			</Empty.Root>
		{/each}
	</section>
</Page>
