<script lang="ts">
	import BoxesIcon from '@lucide/svelte/icons/boxes';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import { goto } from '$app/navigation';
	import EditDialog from '$lib/components/catalog/edit-dialog.svelte';
	import Page from '$lib/components/layout/page.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Table from '$lib/components/ui/table';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	let open = $state(false);
	$effect(() => {
		if (data.dialog.open) open = true;
	});
</script>

<Page title="Catalog" subtitle="Services, owners, and what depends on what">
	<div class="flex flex-col gap-3.5">
		<div class="flex items-center">
			<span class="text-subtle-foreground text-[13px]">
				{data.services.length}
				{data.services.length === 1 ? 'service' : 'services'}
			</span>
			<div class="flex-1"></div>
			<Button size="sm" href="/catalog?new">
				<PlusIcon data-icon="inline-start" />
				New service
			</Button>
		</div>

		{#if !data.anyService}
			<div
				class="text-muted-foreground flex flex-col items-center gap-2.5 rounded-xl border border-dashed px-5 py-14"
			>
				<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
					<BoxesIcon class="text-subtle-foreground size-5" />
				</span>
				<div class="text-[15px] font-medium">No services yet</div>
				<p class="text-subtle-foreground m-0 max-w-[420px] text-center text-[13px] leading-[1.55]">
					Services tie alerts, owners, runbooks, and status pages together. Start with the thing
					that pages you most.
				</p>
				<Button variant="secondary" size="sm" href="/catalog?new">
					<PlusIcon data-icon="inline-start" />
					Register your first service
				</Button>
			</div>
		{:else}
			<div class="bg-card overflow-hidden rounded-xl border">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head class="pl-[18px]">Service</Table.Head>
							<Table.Head>Team</Table.Head>
							<Table.Head class="w-[110px]">Open alerts</Table.Head>
							<Table.Head class="w-[120px]">Open incidents</Table.Head>
							<Table.Head class="w-[150px] pr-[18px]">Dependencies</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each data.services as service (service.id)}
							<Table.Row
								data-clickable="true"
								data-service={service.id}
								onclick={() => goto(`/catalog/${service.id}`)}
							>
								<Table.Cell class="max-w-[400px] py-2.5 pl-[18px] whitespace-normal">
									<!-- Real link keeps the clickable row keyboard reachable -->
									<a
										href="/catalog/{service.id}"
										class="text-foreground hover:text-brand-foreground font-mono text-[13px] font-medium"
										onclick={(event) => event.stopPropagation()}
									>
										{service.id}
									</a>
									<div class="text-subtle-foreground mt-0.5 truncate text-xs">
										{service.description}
									</div>
								</Table.Cell>
								<Table.Cell><Tag>{service.team}</Tag></Table.Cell>
								<Table.Cell>
									{#if service.openAlerts}
										<Badge tone="warning" size="sm">{service.openAlerts}</Badge>
									{:else}
										<span class="text-subtle-foreground font-mono text-xs">0</span>
									{/if}
								</Table.Cell>
								<Table.Cell>
									{#if service.openIncidents}
										<Badge tone="critical" size="sm">{service.openIncidents}</Badge>
									{:else}
										<span class="text-subtle-foreground font-mono text-xs">0</span>
									{/if}
								</Table.Cell>
								<Table.Cell class="text-subtle-foreground pr-[18px] font-mono text-xs">
									{service.dependsOn} up · {service.dependedOnBy} down
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</div>
		{/if}
	</div>
</Page>

<EditDialog bind:open service={data.dialog.service} names={data.names} error={form?.error} />
