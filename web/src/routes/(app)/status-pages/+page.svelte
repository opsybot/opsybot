<script lang="ts">
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import GlobeIcon from '@lucide/svelte/icons/globe';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import { toast } from 'svelte-sonner';
	import Page from '$lib/components/layout/page.svelte';
	import SpTabs from '$lib/components/statuspages/sp-tabs.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<Page title="Status pages" subtitle="Tell customers before they ask">
	<SpTabs current="pages" />

	<div class="mt-3.5 flex flex-col gap-3.5">
		<div class="flex items-center">
			<span class="text-subtle-foreground text-[13px]">
				{data.pages.length}
				{data.pages.length === 1 ? 'page' : 'pages'}
			</span>
			<div class="flex-1"></div>
			<Button
				size="sm"
				onclick={() =>
					toast.info('New pages reuse the settings form — open a page to see it.')}
			>
				<PlusIcon data-icon="inline-start" />
				New page
			</Button>
		</div>

		{#if !data.anyPage}
			<div
				class="text-muted-foreground flex flex-col items-center gap-2.5 rounded-xl border border-dashed px-5 py-14"
			>
				<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
					<GlobeIcon class="text-subtle-foreground size-5" />
				</span>
				<div class="text-[15px] font-medium">No status pages</div>
				<p class="text-subtle-foreground m-0 max-w-[420px] text-center text-[13px] leading-[1.55]">
					A status page tells customers what is happening before they file tickets. One public page
					covers most teams.
				</p>
			</div>
		{:else}
			<div class="bg-card overflow-hidden rounded-xl border">
				{#each data.pages as page (page.id)}
					<a
						href="/status-pages/{page.id}"
						data-page={page.id}
						class="hover:bg-accent flex items-center gap-3 border-t px-4 py-3 first:border-t-0"
					>
						<span
							class="bg-inset text-subtle-foreground flex size-8 shrink-0 items-center justify-center rounded-sm border"
						>
							<GlobeIcon class="size-[15px]" />
						</span>

						<div class="min-w-0 flex-1">
							<div class="flex flex-wrap items-center gap-2">
								<span class="text-[13.5px] font-semibold">{page.name}</span>
								<Badge tone={page.visibility === 'public' ? 'info' : 'neutral'} size="sm">
									{page.visibility}
								</Badge>
								{#if !page.published}
									<Badge tone="neutral" size="sm">unpublished</Badge>
								{/if}
							</div>
							<div class="text-subtle-foreground mt-[3px] font-mono text-[11px]">
								{page.domain} · {page.subscribers.toLocaleString('en-US')} subscribers
							</div>
						</div>

						<Badge tone={page.overallTone} size="sm" dot>{page.overallLabel}</Badge>
						<ChevronRightIcon class="text-subtle-foreground size-[15px] shrink-0" />
					</a>
				{/each}
			</div>
		{/if}
	</div>
</Page>
