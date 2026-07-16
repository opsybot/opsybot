<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import GlobeIcon from '@lucide/svelte/icons/globe';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import EditDialog from '$lib/components/catalog/edit-dialog.svelte';
	import Page from '$lib/components/layout/page.svelte';
	import StatusBadge from '$lib/components/status-badge.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { LINK_KINDS } from '$lib/catalog';
	import { ws } from '$lib/navigation';
	import { formatUtcTime } from '$lib/time';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	const service = $derived(data.service);
	const activity = $derived(data.activity);

	const links = $derived(LINK_KINDS.map((link) => ({ ...link, url: service.links[link.kind] })).filter((link) => link.url));

	let open = $state(false);
	$effect(() => {
		if (data.dialogOpen) open = true;
	});

	const incidentStatus = (status: string) =>
		status === 'resolved' ? 'resolved' : (status as 'investigating' | 'identified' | 'monitoring');
</script>

<Page title="Catalog" subtitle="Services, owners, and what depends on what">
	<div class="flex flex-col gap-3.5">
		<a
			href={ws('/catalog')}
			class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[12.5px]"
		>
			<ArrowLeftIcon class="size-3.5" />
			Catalog
		</a>

		<div class="flex flex-wrap items-center gap-2.5">
			<h2 class="m-0 font-mono text-[18px] font-semibold">{service.id}</h2>
			<Tag>{service.team}</Tag>
			{#if activity.openIncidents}
				<Badge tone="critical" size="sm" dot>
					{activity.openIncidents} open {activity.openIncidents === 1 ? 'incident' : 'incidents'}
				</Badge>
			{/if}
			<div class="flex-1"></div>
			<Button size="sm" variant="secondary" href={ws(`/catalog/${service.id}?edit`)}>
				<PencilIcon data-icon="inline-start" />
				Edit service
			</Button>
		</div>

		<p class="text-muted-foreground m-0 max-w-[640px] text-[13.5px] leading-[1.6]">
			{service.description}
		</p>

		<div class="grid items-start gap-3.5 min-[1100px]:grid-cols-[minmax(0,1fr)_320px]">
			<div class="flex min-w-0 flex-col gap-3.5">
				<section class="bg-card overflow-hidden rounded-xl border">
					<header class="border-b px-4 py-3">
						<span class="text-[13.5px] font-semibold">Recent incidents</span>
					</header>

					{#each activity.incidents as incident (incident.id)}
						<a
							href={ws(`/incidents/${incident.id}`)}
							class="hover:bg-accent flex items-center gap-2.5 border-t px-3.5 py-[11px] first:border-t-0"
						>
							<Badge tone={incident.tone} size="sm">{incident.severity}</Badge>
							<span class="min-w-0 flex-1 truncate text-[13px] font-medium">{incident.name}</span>
							<StatusBadge status={incidentStatus(incident.status)} size="sm" />
							<span class="text-subtle-foreground shrink-0 font-mono text-[11px]">{incident.id}</span>
						</a>
					{:else}
						<p class="text-subtle-foreground m-0 px-4 py-[18px] text-[12.5px]">
							Nothing in the last 30 days.
						</p>
					{/each}
				</section>

				<section class="bg-card overflow-hidden rounded-xl border">
					<header class="border-b px-4 py-3">
						<span class="text-[13.5px] font-semibold">Open alerts</span>
					</header>

					{#each activity.alerts as alert (alert.id)}
						<a
							href={ws(`/alerts/${alert.id}`)}
							class="hover:bg-accent flex items-center gap-2.5 border-t px-3.5 py-[11px] first:border-t-0"
							style="box-shadow: inset 3px 0 0 var(--{alert.tone})"
						>
							<span class="min-w-0 flex-1 truncate text-[13px]">{alert.title}</span>
							<span class="text-subtle-foreground shrink-0 font-mono text-[11px]">
								{formatUtcTime(alert.lastSeenAt)}
							</span>
						</a>
					{:else}
						<p class="text-subtle-foreground m-0 px-4 py-[18px] text-[12.5px]">None right now.</p>
					{/each}
				</section>
			</div>

			<div class="flex flex-col gap-3.5">
				<section class="bg-card overflow-hidden rounded-xl border">
					<header class="border-b px-4 py-3">
						<span class="text-[13.5px] font-semibold">Links</span>
					</header>

					{#each links as link (link.kind)}
						<div class="flex items-center gap-2.5 border-t px-3.5 py-[11px] first:border-t-0">
							<link.icon class="text-subtle-foreground size-[13px] shrink-0" />
							<span class="text-foreground flex-1 text-[12.5px]">{link.label}</span>
							<span class="text-subtle-foreground shrink-0 font-mono text-[10.5px]">{link.url}</span>
						</div>
					{:else}
						<p class="text-subtle-foreground m-0 px-4 py-[14px] text-[12.5px]">No links yet.</p>
					{/each}
				</section>

				<section class="bg-card overflow-hidden rounded-xl border">
					<header class="border-b px-4 py-3">
						<span class="text-[13.5px] font-semibold">Dependencies</span>
					</header>

					<div class="flex flex-col gap-3 px-3.5 py-3">
						<div>
							<div class="text-subtle-foreground tracking-label mb-1.5 text-[10.5px] uppercase">
								Depends on
							</div>
							<div class="flex flex-wrap gap-1.5">
								{#each service.deps as dep (dep)}
									<Tag href={ws(`/catalog/${dep}`)}>{dep}</Tag>
								{:else}
									<span class="text-subtle-foreground text-xs">nothing</span>
								{/each}
							</div>
						</div>

						<div>
							<div class="text-subtle-foreground tracking-label mb-1.5 text-[10.5px] uppercase">
								Depended on by
							</div>
							<div class="flex flex-wrap gap-1.5">
								{#each activity.dependedOnBy as dep (dep)}
									<Tag href={ws(`/catalog/${dep}`)}>{dep}</Tag>
								{:else}
									<span class="text-subtle-foreground text-xs">nothing</span>
								{/each}
							</div>
						</div>
					</div>
				</section>

				<section class="bg-card overflow-hidden rounded-xl border">
					<header class="border-b px-4 py-3">
						<span class="text-[13.5px] font-semibold">Status page components</span>
					</header>

					{#each service.statusComponents as component (component)}
						<div class="flex items-center gap-2.5 border-t px-3.5 py-[11px] first:border-t-0">
							<GlobeIcon class="text-subtle-foreground size-[13px] shrink-0" />
							<span class="flex-1 text-[12.5px]">{component}</span>
							<span class="text-subtle-foreground shrink-0 font-mono text-[10.5px]">status.acme.dev</span>
						</div>
					{:else}
						<p class="text-subtle-foreground m-0 px-4 py-[14px] text-[12.5px]">
							Not mapped to any public component.
						</p>
					{/each}
				</section>
			</div>
		</div>
	</div>
</Page>

<EditDialog bind:open service={data.service} names={data.names} error={form?.error} />
