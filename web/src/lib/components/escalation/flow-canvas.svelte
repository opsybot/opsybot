<script lang="ts">
	import FlowSeq from '$lib/components/escalation/flow-seq.svelte';
	import PlainConnector from '$lib/components/escalation/plain-connector.svelte';
	import TriggerNode from '$lib/components/escalation/trigger-node.svelte';
	import type { FlowContext } from '$lib/components/escalation/flow-context';
	import type { Tree } from '$lib/escalation';

	let {
		tree,
		editable = false,
		selectedId = null,
		activeId = null,
		laneChoices = null,
		onselect,
		oninsertafter,
		onaddlevel,
		onaddbranch,
		ondeletenode,
		onchangewait
	}: {
		tree: Tree;
		editable?: boolean;
		selectedId?: string | null;
		activeId?: string | null;
		laneChoices?: Record<string, string> | null;
		onselect?: (id: string) => void;
		oninsertafter?: (afterId: string) => void;
		onaddlevel?: (ownerId: string) => void;
		onaddbranch?: (ownerId: string) => void;
		ondeletenode?: (id: string) => void;
		onchangewait?: (id: string, wait: string) => void;
	} = $props();

	const ctx = $derived<FlowContext>({
		tree,
		editable,
		selectedId,
		activeId,
		laneChoices,
		onSelect: onselect,
		onInsertAfter: oninsertafter,
		onAddLevel: onaddlevel,
		onAddBranch: onaddbranch,
		onDeleteNode: ondeletenode,
		onChangeWait: onchangewait
	});
</script>

<div class="flex flex-col items-center px-1 pt-1 pb-1.5 [min-width:min-content]">
	<TriggerNode />
	<PlainConnector />
	<FlowSeq nodes={tree.nodes} ownerId="root" {ctx} />
</div>
