<script lang="ts">
	import { goto } from '$app/navigation';
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import StatusBadge from '$lib/components/status-badge.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { createSvelteTable } from '$lib/components/ui/data-table';
	import * as Table from '$lib/components/ui/table';
	import { SEVERITY_TONE } from '$lib/dashboard';
	import type { Incident } from '$lib/incidents';
	import { ws } from '$lib/navigation';
	import { formatAge, formatDue } from '$lib/time';
	import {
		getCoreRowModel,
		getFilteredRowModel,
		type ColumnDef,
		type ColumnFiltersState
	} from '@tanstack/table-core';

	let {
		incidents,
		filters,
		now
	}: {
		incidents: Incident[];
		filters: ColumnFiltersState;
		now: number;
	} = $props();

	const columns: ColumnDef<Incident>[] = [
		{
			id: 'search',
			accessorFn: (row) => `${row.name} ${row.id}`.toLowerCase(),
			filterFn: (row, id, value) => String(row.getValue(id)).includes(String(value).toLowerCase())
		},
		{ id: 'severity', accessorKey: 'severity' },
		{ id: 'status', accessorKey: 'status' },
		{
			id: 'service',
			accessorFn: (row) => row.services,
			filterFn: (row, id, value) => (row.getValue(id) as string[]).includes(String(value))
		},
		{ id: 'team', accessorKey: 'team' },
		{ id: 'lead', accessorKey: 'lead' },
		{
			id: 'preset',
			accessorFn: (row) => row.id,
			filterFn: (row, _id, value) => {
				if (value === 'mine') return row.original.mine;
				if (value === 'active') return row.original.status !== 'resolved';
				return true;
			}
		},
		{
			id: 'range',
			accessorFn: (row) => row.declaredAt,
			filterFn: (row, id, value) => {
				const days = Number(String(value).replace('d', ''));
				return Date.parse(row.getValue(id) as string) >= now - days * 86_400_000;
			}
		}
	];

	const table = createSvelteTable({
		get data() {
			return incidents;
		},
		columns,
		state: {
			get columnFilters() {
				return filters;
			}
		},
		getCoreRowModel: getCoreRowModel(),
		getFilteredRowModel: getFilteredRowModel()
	});

	const rows = $derived(table.getRowModel().rows.map((row) => row.original));
</script>

<Table.Root class="text-[13.5px]">
	<Table.Header>
		<Table.Row class="hover:bg-transparent">
			<Table.Head class="py-2.5 pr-3 pl-[18px] text-[11.5px]">Incident</Table.Head>
			<Table.Head class="px-3 py-2.5 text-[11.5px]">Status</Table.Head>
			<Table.Head class="px-3 py-2.5 text-[11.5px]">Lead</Table.Head>
			<Table.Head class="px-3 py-2.5 text-[11.5px]">Age</Table.Head>
			<Table.Head class="py-2.5 pr-[18px] pl-3 text-[11.5px]">Next update</Table.Head>
		</Table.Row>
	</Table.Header>

	<Table.Body>
		{#each rows as incident (incident.id)}
			{@const overdue = incident.nextUpdateAt && Date.parse(incident.nextUpdateAt) < now}
			<Table.Row data-clickable="true" onclick={() => goto(ws(`/incidents/${incident.id}`))}>
				<Table.Cell class="p-3 pl-[18px]">
					<div class="flex items-center gap-2.5">
						<Badge tone={SEVERITY_TONE[incident.severity]} size="sm">{incident.severity}</Badge>
						<div>
							<div class="text-foreground font-medium">{incident.name}</div>
							<span class="text-subtle-foreground font-mono text-[11px]">
								{incident.id} · {incident.services.join(', ')}
							</span>
						</div>
					</div>
				</Table.Cell>

				<Table.Cell class="p-3">
					<StatusBadge status={incident.status} size="sm" />
				</Table.Cell>

				<Table.Cell class="p-3">
					<div class="flex items-center gap-2">
						<UserAvatar name={incident.lead} size="xs" onCall={incident.status !== 'resolved'} />
						<span class="text-muted-foreground">{incident.lead}</span>
					</div>
				</Table.Cell>

				<Table.Cell class="text-subtle-foreground p-3 font-mono">
					{formatAge(now - Date.parse(incident.declaredAt))}
				</Table.Cell>

				<Table.Cell
					class="p-3 pr-[18px] font-mono {overdue ? 'text-critical-ink' : 'text-subtle-foreground'}"
				>
					{incident.nextUpdateAt ? formatDue(incident.nextUpdateAt, now) : '–'}
				</Table.Cell>
			</Table.Row>
		{/each}
	</Table.Body>
</Table.Root>

{#if rows.length === 0}
	<div class="flex flex-col items-center gap-2.5 px-5 py-9">
		<div class="text-sm font-medium">Nothing matches these filters</div>
		<a href={ws('/incidents')} class="text-muted-foreground hover:text-brand-foreground text-[12.5px]">
			Reset all filters
		</a>
	</div>
{/if}
