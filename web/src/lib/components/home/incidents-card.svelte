<script lang="ts">
	import { goto } from '$app/navigation';
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import StatusBadge from '$lib/components/status-badge.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import * as Table from '$lib/components/ui/table';
	import { SEVERITY_TONE, type Incident } from '$lib/dashboard';
	import { ws } from '$lib/navigation';
	import Card from './card.svelte';
	import LiveAge from './live-age.svelte';
	import QuietRow from './quiet-row.svelte';

	let { incidents, now }: { incidents: Incident[]; now: number } = $props();
</script>

<Card
	title="Active incidents"
	count={incidents.length}
	countTone={incidents.length ? 'critical' : 'neutral'}
	accent={incidents.length ? 'var(--critical)' : undefined}
	live={incidents.length > 0}
>
	{#snippet aside()}
		<a href={ws('/incidents')} class="text-brand-foreground text-[12.5px] hover:underline">View all</a>
	{/snippet}

	{#if incidents.length === 0}
		<QuietRow text="No active incidents." />
	{:else}
		<Table.Root class="text-[13.5px]">
			<Table.Header>
				<Table.Row class="hover:bg-transparent">
					<Table.Head class="py-[9px] pr-3 pl-[18px] text-[11.5px]">Incident</Table.Head>
					<Table.Head class="px-3 py-[9px] text-[11.5px]">Status</Table.Head>
					<Table.Head class="px-3 py-[9px] text-[11.5px]">Lead</Table.Head>
					<Table.Head class="px-3 py-[9px] text-[11.5px]">Age</Table.Head>
				</Table.Row>
			</Table.Header>

			<Table.Body>
				{#each incidents as incident (incident.id)}
					<Table.Row data-clickable="true" onclick={() => goto(ws('/incidents'))}>
						<Table.Cell class="p-3 pl-[18px]">
							<div class="flex items-center gap-2.5">
								<Badge tone={SEVERITY_TONE[incident.severity]} size="sm">
									{incident.severity}
								</Badge>
								<div>
									<div class="text-foreground font-medium">{incident.title}</div>
									<span class="text-subtle-foreground font-mono text-[11.5px]">{incident.id}</span>
								</div>
							</div>
						</Table.Cell>

						<Table.Cell class="p-3">
							<StatusBadge status={incident.status} size="sm" />
						</Table.Cell>

						<Table.Cell class="p-3">
							<div class="flex items-center gap-2">
								<UserAvatar name={incident.lead} onCall size="xs" />
								<span class="text-muted-foreground">{incident.lead}</span>
							</div>
						</Table.Cell>

						<Table.Cell class="text-subtle-foreground p-3 font-mono">
							<LiveAge since={incident.declaredAt} {now} />
						</Table.Cell>
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	{/if}
</Card>
