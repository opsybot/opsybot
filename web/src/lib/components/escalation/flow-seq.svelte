<script lang="ts">
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import AddAffordance from '$lib/components/escalation/add-affordance.svelte';
	import BranchNode from '$lib/components/escalation/branch-node.svelte';
	import EndNode from '$lib/components/escalation/end-node.svelte';
	import Self from '$lib/components/escalation/flow-seq.svelte';
	import ForkSvg from '$lib/components/escalation/fork-svg.svelte';
	import LaneHead from '$lib/components/escalation/lane-head.svelte';
	import LevelNode from '$lib/components/escalation/level-node.svelte';
	import WaitConnector from '$lib/components/escalation/wait-connector.svelte';
	import type { FlowContext } from '$lib/components/escalation/flow-context';
	import { laneMeta, type EscNode } from '$lib/escalation';

	let { nodes, ownerId, ctx }: { nodes: EscNode[]; ownerId: string; ctx: FlowContext } = $props();

	const endsInBranch = $derived(nodes.length > 0 && nodes[nodes.length - 1].type === 'branch');

	function laneTone(branchId: string, laneId: string): string {
		if (!ctx.laneChoices || !ctx.laneChoices[branchId]) return 'strong';
		return ctx.laneChoices[branchId] === laneId ? 'mint' : 'faint';
	}
</script>

<div class="flex flex-col items-center">
	{#each nodes as node (node.id)}
		{#if node.type === 'level'}
			<LevelNode
				{node}
				editable={ctx.editable}
				selected={ctx.selectedId === node.id}
				active={ctx.activeId === node.id}
				onselect={ctx.onSelect}
				ondelete={ctx.onDeleteNode}
			/>
			<WaitConnector
				wait={node.wait}
				aboveId={node.id}
				editable={ctx.editable}
				oninsert={ctx.onInsertAfter}
				onchangewait={ctx.onChangeWait}
			/>
		{:else}
			<BranchNode
				{node}
				editable={ctx.editable}
				selected={ctx.selectedId === node.id}
				active={ctx.activeId === node.id}
				onselect={ctx.onSelect}
			/>
			<div class="inline-flex flex-col items-stretch">
				<ForkSvg tones={node.lanes.map((lane) => laneTone(node.id, lane.id))} />
				<div class="flex items-start justify-center gap-8">
					{#each node.lanes as lane (lane.id)}
						{@const dim = !!ctx.laneChoices && !!ctx.laneChoices[node.id] && ctx.laneChoices[node.id] !== lane.id}
						<div class="flex flex-col items-center transition-opacity {dim ? 'opacity-30' : ''}">
							<LaneHead meta={laneMeta(node, lane)} />
							<div class="flex flex-col items-center">
								<span class="bg-border-strong h-[9px] w-[1.5px]"></span>
								<ChevronDownIcon class="text-subtle-foreground -mt-1 size-3.5" />
							</div>
							<Self nodes={lane.nodes} ownerId={lane.id} {ctx} />
						</div>
					{/each}
				</div>
			</div>
		{/if}
	{/each}

	{#if !endsInBranch}
		{#if ctx.editable}
			<AddAffordance
				onaddlevel={() => ctx.onAddLevel?.(ownerId)}
				onaddbranch={() => ctx.onAddBranch?.(ownerId)}
			/>
		{/if}
		<EndNode repeat={ctx.tree.repeat} active={ctx.activeId === `end-${ownerId}`} />
	{/if}
</div>
