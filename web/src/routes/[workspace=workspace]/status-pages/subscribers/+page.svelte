<script lang="ts">
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import MailIcon from '@lucide/svelte/icons/mail';
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';
	import RotateCwIcon from '@lucide/svelte/icons/rotate-cw';
	import RssIcon from '@lucide/svelte/icons/rss';
	import SearchIcon from '@lucide/svelte/icons/search';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import WebhookIcon from '@lucide/svelte/icons/webhook';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import Page from '$lib/components/layout/page.svelte';
	import SpTabs from '$lib/components/statuspages/sp-tabs.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	$effect(() => {
		if (form?.removed) toast.success(`${form.removed} removed — deletion requests are honored immediately.`);
	});
	$effect(() => {
		if (form?.redelivered) toast.success('Redelivered — endpoint returned 200.');
	});

	let typed: ReturnType<typeof setTimeout>;
	function search(value: string) {
		clearTimeout(typed);
		typed = setTimeout(() => {
			const url = new URL(page.url);
			if (value) url.searchParams.set('q', value);
			else url.searchParams.delete('q');
			goto(url, { keepFocus: true, noScroll: true, replaceState: true });
		}, 200);
	}

	const CHANNELS = [
		{ icon: MailIcon, label: 'Email', get: () => data.counts.email },
		{ icon: RssIcon, label: 'Feed', get: () => data.counts.feed },
		{ icon: WebhookIcon, label: 'Webhook', get: () => data.counts.webhook }
	];
</script>

<Page title="Status pages" subtitle="Tell customers before they ask">
	<SpTabs current="subscribers" />

	<div class="mt-3.5 flex max-w-[760px] flex-col gap-3.5">
		<div class="grid grid-cols-3 gap-2.5">
			{#each CHANNELS as channel (channel.label)}
				<div class="bg-card flex items-center gap-2.5 rounded-xl border px-4 py-3.5">
					<channel.icon class="text-subtle-foreground size-[15px] shrink-0" />
					<span class="font-mono text-lg font-semibold">
						{channel.get().toLocaleString('en-US')}
					</span>
					<span class="text-subtle-foreground text-[11.5px]">{channel.label}</span>
				</div>
			{/each}
		</div>

		<section class="bg-card overflow-hidden rounded-xl border">
			<header class="flex flex-wrap items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Email subscribers</span>
				<span class="text-subtle-foreground text-[11.5px]">visible to page admins only</span>
				<div class="flex-1"></div>
				<div class="relative w-[190px]">
					<SearchIcon
						class="text-subtle-foreground pointer-events-none absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2"
					/>
					<Input
						value={data.query}
						placeholder="Search emails"
						aria-label="Search emails"
						class="h-[34px] pl-[30px] text-[13px]"
						oninput={(event: Event) => search((event.currentTarget as HTMLInputElement).value)}
					/>
				</div>
				<Button
					size="sm"
					variant="ghost"
					onclick={() =>
						toast.success(`CSV export of ${data.counts.email.toLocaleString('en-US')} subscribers started.`)}
				>
					<DownloadIcon data-icon="inline-start" />
					Export
				</Button>
			</header>

			{#each data.emails as subscriber (subscriber.address)}
				<div class="flex items-center gap-3 border-t px-4 py-2.5">
					<span class="text-foreground min-w-0 flex-1 truncate font-mono text-[12.5px]">
						{subscriber.address}
					</span>
					<span class="text-subtle-foreground shrink-0 font-mono text-[11px]">
						since {subscriber.since}
					</span>
					<form method="POST" action="?/remove" use:enhance>
						<input type="hidden" name="address" value={subscriber.address} />
						<Button type="submit" size="sm" variant="ghost">
							<Trash2Icon data-icon="inline-start" />
							Delete on request
						</Button>
					</form>
				</div>
			{:else}
				<p class="text-subtle-foreground m-0 px-4 py-[18px] text-[12.5px]">No matches.</p>
			{/each}

			<p class="text-subtle-foreground m-0 border-t px-4 py-2 text-[11.5px]">
				{#if data.query}
					Showing {data.emails.length} matching “{data.query}”.
				{:else}
					Showing {data.emails.length} of {data.counts.email.toLocaleString('en-US')} — search to find
					a specific address.
				{/if}
			</p>
		</section>

		<section class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2.5 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Webhook subscribers</span>
				<Badge tone="neutral" size="sm">{data.webhooks.length}</Badge>
			</header>

			{#each data.webhooks as webhook (webhook.url)}
				<div class="flex items-center gap-3 border-t px-4 py-2.5">
					{#if webhook.ok}
						<CircleCheckIcon class="text-success-ink size-3.5 shrink-0" />
					{:else}
						<OctagonAlertIcon class="text-critical-ink size-3.5 shrink-0" />
					{/if}
					<div class="min-w-0 flex-1">
						<div class="text-foreground truncate font-mono text-xs">{webhook.url}</div>
						<div
							class="mt-0.5 font-mono text-[10.5px] {webhook.ok
								? 'text-subtle-foreground'
								: 'text-critical-ink'}"
						>
							last delivery {webhook.last}
						</div>
					</div>
					{#if !webhook.ok}
						<form method="POST" action="?/redeliver" use:enhance>
							<input type="hidden" name="url" value={webhook.url} />
							<Button type="submit" size="sm" variant="secondary">
								<RotateCwIcon data-icon="inline-start" />
								Redeliver
							</Button>
						</form>
					{/if}
				</div>
			{/each}
		</section>
	</div>
</Page>
