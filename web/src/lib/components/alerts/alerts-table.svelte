<script lang="ts">
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import ChevronUpIcon from '@lucide/svelte/icons/chevron-up';
	import { SvelteSet } from 'svelte/reactivity';
	import { goto } from '$app/navigation';
	import Tag from '$lib/components/tag.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Table from '$lib/components/ui/table';
	import AlertStatus from './alert-status.svelte';
	import { SEVERITY_SHORT, SEVERITY_TONE, type Alert } from '$lib/alerts';
	import { ws } from '$lib/navigation';
	import { formatSince } from '$lib/time';

	let {
		alerts,
		now,
		selected
	}: {
		alerts: Alert[];
		now: number;
		selected: SvelteSet<string>;
	} = $props();

	const expanded = new SvelteSet<string>();
	const rows = $derived(alerts);

	function toggle(set: SvelteSet<string>, value: string) {
		if (set.has(value)) set.delete(value);
		else set.add(value);
	}
</script>

<Table.Root class="text-[13.5px]">
	<Table.Header>
		<Table.Row class="hover:bg-transparent">
			<Table.Head class="w-9 py-2.5 pl-3.5"></Table.Head>
			<Table.Head class="px-3 py-2.5 text-[11.5px]">Alert</Table.Head>
			<Table.Head class="px-3 py-2.5 text-[11.5px]">Status</Table.Head>
			<Table.Head class="px-3 py-2.5 text-[11.5px]">Source</Table.Head>
			<Table.Head class="px-3 py-2.5 text-[11.5px]">Service</Table.Head>
			<Table.Head class="py-2.5 pr-[18px] pl-3 text-[11.5px]">Last seen</Table.Head>
		</Table.Row>
	</Table.Header>

	<Table.Body>
		{#each rows as alert (alert.id)}
			<Table.Row
				data-clickable="true"
				onclick={() => goto(ws(`/alerts/${alert.id}`))}
				style="box-shadow: inset 3px 0 0 var(--{alert.severity})"
			>
				<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
				<Table.Cell class="p-3 pl-3.5" onclick={(event: MouseEvent) => event.stopPropagation()}>
					<Checkbox
						checked={selected.has(alert.id)}
						onCheckedChange={() => toggle(selected, alert.id)}
						aria-label={alert.title}
					/>
				</Table.Cell>

				<Table.Cell class="p-3 whitespace-normal">
					<div class="flex items-center gap-[9px]">
						<Badge tone={SEVERITY_TONE[alert.severity]} size="sm">
							{SEVERITY_SHORT[alert.severity]}
						</Badge>
						<span class="text-foreground font-medium">{alert.title}</span>

						{#if alert.count > 1}
							<button
								type="button"
								onclick={(event) => {
									event.stopPropagation();
									if (alert.children.length) toggle(expanded, alert.id);
								}}
								class="bg-inset border-input text-muted-foreground hover:text-brand-foreground hover:border-brand-edge inline-flex items-center gap-[3px] rounded-full border px-2 py-0.5 font-mono text-[11px]"
							>
								×{alert.count}
								{#if alert.children.length}
									{#if expanded.has(alert.id)}
										<ChevronUpIcon class="size-3" />
									{:else}
										<ChevronDownIcon class="size-3" />
									{/if}
								{/if}
							</button>
						{/if}
					</div>
					<div class="text-subtle-foreground mt-0.5 font-mono text-[11px]">
						{alert.labels.join('  ')}
					</div>
				</Table.Cell>

				<Table.Cell class="p-3"><AlertStatus {alert} /></Table.Cell>
				<Table.Cell class="text-subtle-foreground p-3 font-mono">{alert.source}</Table.Cell>
				<Table.Cell class="p-3"><Tag>{alert.service}</Tag></Table.Cell>
				<Table.Cell class="text-subtle-foreground p-3 pr-[18px] font-mono">
					{formatSince(now - Date.parse(alert.lastSeenAt))}
				</Table.Cell>
			</Table.Row>

			{#if expanded.has(alert.id)}
				{#each alert.children as child (child.id)}
					<Table.Row class="hover:bg-transparent">
						<Table.Cell class="bg-inset px-3 py-2"></Table.Cell>
						<Table.Cell class="bg-inset px-3 py-2" colspan={2}>
							<span class="bg-border-strong mr-2.5 inline-block h-px w-3.5 align-middle"></span>
							<span class="text-muted-foreground text-[12.5px]">{child.title}</span>
							{#if child.count > 1}
								<span class="text-subtle-foreground ml-2 font-mono text-[11px]">×{child.count}</span>
							{/if}
						</Table.Cell>
						<Table.Cell class="bg-inset px-3 py-2 capitalize">
							<span class="text-subtle-foreground text-[12px]">{child.status}</span>
						</Table.Cell>
						<Table.Cell class="bg-inset px-3 py-2"></Table.Cell>
						<Table.Cell class="bg-inset px-3 py-2"></Table.Cell>
						<Table.Cell class="bg-inset text-subtle-foreground px-3 py-2 pr-[18px] font-mono">
							{formatSince(now - Date.parse(child.lastSeenAt))}
						</Table.Cell>
					</Table.Row>
				{/each}
			{/if}
		{/each}
	</Table.Body>
</Table.Root>
