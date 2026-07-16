<script lang="ts">
	import HeartPulseIcon from '@lucide/svelte/icons/heart-pulse';
	import AlertsTabs from '$lib/components/alerts/alerts-tabs.svelte';
	import NewMonitorDialog from '$lib/components/alerts/new-monitor-dialog.svelte';
	import Page from '$lib/components/layout/page.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Empty from '$lib/components/ui/empty';
	import * as Table from '$lib/components/ui/table';
	import { formatSince, formatUtcTime } from '$lib/time';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let creating = $state(false);
</script>

<Page title="Alerts" subtitle="Deduplicated signals from every connected source">
	<AlertsTabs current="heartbeats" />

	<section class="bg-card overflow-hidden rounded-xl border">
		<header class="flex items-center gap-3 border-b px-4 py-3">
			<div>
				<div class="text-sm font-semibold">Heartbeat monitors</div>
				<p class="text-subtle-foreground m-0 mt-0.5 text-[12.5px]">
					A job that stops checking in pages the on-call, the same as any other alert.
				</p>
			</div>
			<div class="flex-1"></div>
			<Button size="sm" onclick={() => (creating = true)}>
				<HeartPulseIcon data-icon="inline-start" />
				New monitor
			</Button>
		</header>

		{#if data.heartbeats.length}
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Monitor</Table.Head>
						<Table.Head class="w-[110px]">State</Table.Head>
						<Table.Head class="w-[130px]">Interval</Table.Head>
						<Table.Head class="w-[100px]">Grace</Table.Head>
						<Table.Head class="w-[190px]">Last check-in</Table.Head>
						<Table.Head class="w-[170px]">Routing</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each data.heartbeats as heartbeat (heartbeat.id)}
						<Table.Row
							style={heartbeat.state === 'missed'
								? 'box-shadow: inset 3px 0 0 var(--critical)'
								: undefined}
						>
							<Table.Cell class="font-mono text-[12.5px] font-medium">{heartbeat.name}</Table.Cell>
							<Table.Cell>
								<Badge tone={heartbeat.state === 'missed' ? 'critical' : 'success'} size="sm">
									{heartbeat.state}
								</Badge>
							</Table.Cell>
							<Table.Cell class="text-muted-foreground">{heartbeat.interval}</Table.Cell>
							<Table.Cell class="text-muted-foreground">{heartbeat.grace}</Table.Cell>
							<Table.Cell class="text-subtle-foreground font-mono text-[11.5px]">
								{#if heartbeat.lastSeenAt}
									{formatSince(data.now - Date.parse(heartbeat.lastSeenAt))} · {formatUtcTime(
										heartbeat.lastSeenAt
									)}
								{:else}
									never
								{/if}
							</Table.Cell>
							<Table.Cell class="text-muted-foreground font-mono text-[11.5px]">
								{heartbeat.policy}
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		{:else}
			<Empty.Root class="py-10">
				<Empty.Header>
					<Empty.Media variant="icon">
						<HeartPulseIcon />
					</Empty.Media>
					<Empty.Title>No heartbeat monitors</Empty.Title>
					<Empty.Description>
						Point a cron job or worker at a check-in URL and Opsybot pages when it goes quiet.
					</Empty.Description>
				</Empty.Header>
				<Empty.Content>
					<Button size="sm" onclick={() => (creating = true)}>New monitor</Button>
				</Empty.Content>
			</Empty.Root>
		{/if}
	</section>
</Page>

<NewMonitorDialog bind:open={creating} />
