<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import GitBranchIcon from '@lucide/svelte/icons/git-branch';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import FlowCanvas from '$lib/components/escalation/flow-canvas.svelte';
	import TracePanel from '$lib/components/escalation/trace-panel.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { untrack } from 'svelte';
	import { computeTrace } from '$lib/escalation';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const tree = $derived(data.tree);
	const repeats = $derived(!!tree.repeat && tree.repeat !== '0');

	let scenario = $state({ priority: 'high', hours: 'off' });
	let activeIndex = $state(-1);
	let running = $state(false);

	// Reset trace state on route param change; the page instance is reused across policies
	$effect(() => {
		data.id;
		untrack(() => {
			scenario = { priority: 'high', hours: 'off' };
			activeIndex = -1;
			running = false;
		});
	});

	const trace = $derived(computeTrace(tree, scenario));
	const activeId = $derived(activeIndex >= 0 ? (trace.steps[activeIndex]?.id ?? null) : null);
	const laneChoices = $derived(activeIndex >= 0 ? trace.laneChoices : null);
</script>

<div class="flex flex-col gap-3.5">
	<a
		href={ws('/escalation-policies')}
		class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[12.5px] transition-colors"
	>
		<ArrowLeftIcon class="size-3.5" />
		All policies
	</a>

	<div class="flex flex-wrap items-center gap-2.5">
		<h2 class="font-mono text-[18px] font-semibold">{tree.name}</h2>
		<Tag>{tree.team}</Tag>
		{#if data.branch}
			<Badge tone="info" size="sm" dot>branches by {data.branch}</Badge>
		{/if}
		<div class="flex-1"></div>
		<Button size="sm" variant="secondary" href={ws(`/escalation-policies/${data.id}/edit`)}>
			<PencilIcon data-icon="inline-start" />
			Edit path
		</Button>
	</div>

	<div class="grid items-start gap-3.5 min-[1000px]:[grid-template-columns:minmax(0,1fr)_360px]">
		<section class="bg-card min-w-0 overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[14px] font-semibold">Escalation path</span>
				<span class="text-subtle-foreground ml-auto text-[12px]">
					{repeats ? `repeats ×${tree.repeat} if never acknowledged` : "doesn't repeat once exhausted"}
				</span>
			</header>
			<div class="overflow-x-auto px-3 pt-[18px] pb-5">
				<FlowCanvas {tree} editable={false} {activeId} {laneChoices} />
			</div>
		</section>

		<div class="flex flex-col gap-3.5">
			<TracePanel {tree} {trace} bind:scenario bind:activeIndex bind:running />

			<div class="bg-card overflow-hidden rounded-xl border">
				<header class="flex items-center gap-2 border-b px-4 py-3">
					<span class="text-[13.5px] font-semibold">Linked routing rules</span>
					<Badge tone="neutral" size="sm">{data.routing.length}</Badge>
				</header>
				{#each data.routing as link (link.rule)}
					<div class="flex items-start gap-2.5 border-t px-[14px] py-[11px] first:border-t-0">
						<GitBranchIcon class="text-subtle-foreground mt-0.5 size-[13px] shrink-0" />
						<div class="min-w-0 flex-1">
							<div class="font-mono text-[11.5px] leading-[1.5]">{link.rule}</div>
							<div class="text-subtle-foreground mt-0.5 font-mono text-[10.5px]">{link.matched}</div>
						</div>
					</div>
				{:else}
					<div class="text-subtle-foreground px-[14px] py-4 text-center text-[12.5px]">
						No routing rules point here yet. Send alerts to this policy from the routing tab.
					</div>
				{/each}
			</div>
		</div>
	</div>

	<div class="bg-card overflow-hidden rounded-xl border">
		<header class="flex items-center gap-2 border-b px-4 py-3">
			<span class="text-[14px] font-semibold">Recent escalations</span>
		</header>
		<div class="overflow-x-auto">
			<table class="w-full border-collapse text-[13.5px]">
				<thead>
					<tr class="text-subtle-foreground text-left text-[11.5px] tracking-[0.05em] uppercase">
						<th class="py-2.5 pr-3 pl-[18px] font-semibold">Alert</th>
						<th class="px-3 py-2.5 font-semibold">Started</th>
						<th class="px-3 py-2.5 font-semibold">Outcome</th>
						<th class="py-2.5 pr-[18px] pl-3 font-semibold">Time to outcome</th>
					</tr>
				</thead>
				<tbody>
					{#each data.recent as escalation, index (index)}
						<tr class="border-t">
							<td class="py-3 pr-3 pl-[18px]">
								<span class="font-medium">{escalation.alert}</span>
								{#if escalation.id}
									<span class="text-subtle-foreground ml-2 font-mono text-[11px]">→ {escalation.id}</span>
								{/if}
							</td>
							<td class="text-subtle-foreground px-3 py-3 font-mono">{escalation.at}</td>
							<td class="px-3 py-3">
								<Badge tone={escalation.tone} size="sm">{escalation.outcome}</Badge>
								{#if escalation.by}
									<span class="text-muted-foreground ml-2 text-[12px]">{escalation.by}</span>
								{/if}
							</td>
							<td class="text-subtle-foreground py-3 pr-[18px] pl-3 font-mono">{escalation.duration}</td>
						</tr>
					{:else}
						<tr class="border-t">
							<td colspan="4" class="text-subtle-foreground px-[18px] py-8 text-center text-[13px]">
								No escalations yet. When an alert routes here, its path shows up in this list.
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
</div>
