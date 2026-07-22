<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import CheckIcon from '@lucide/svelte/icons/check';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import MousePointerClickIcon from '@lucide/svelte/icons/mouse-pointer-click';
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import BranchInspector from '$lib/components/escalation/branch-inspector.svelte';
	import FlowCanvas from '$lib/components/escalation/flow-canvas.svelte';
	import LevelInspector from '$lib/components/escalation/level-inspector.svelte';
	import PolicyInspector from '$lib/components/escalation/policy-inspector.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import {
		analyzeTree,
		appendNode,
		changeCondition,
		findNode,
		insertLevelAfterDeep,
		mkBranch,
		mkLevel,
		moveNodeDeep,
		nodeSiblings,
		removeNodeDeep,
		saveBlocked,
		updateNodes,
		type BranchKind,
		type Directory,
		type Level,
		type Tree
	} from '$lib/escalation';
	import { ws } from '$lib/navigation';

	let {
		initial,
		directory,
		backHref
	}: { initial: Tree; directory: Directory; backHref: string } = $props();

	let tree = $state<Tree>(untrack(() => structuredClone(initial)));
	let selectedId = $state<string | null>(null);
	let pendingDelete = $state<string | null>(null);

	const selected = $derived(selectedId ? findNode(tree.nodes, selectedId) : null);
	$effect(() => {
		if (selectedId && !findNode(tree.nodes, selectedId)) selectedId = null;
	});

	const analysis = $derived(analyzeTree(tree));
	const blocked = $derived(saveBlocked(analysis) || !tree.name.trim());
	const treeJson = $derived(JSON.stringify(tree));

	function update(id: string, fn: (level: Level) => Level) {
		tree = { ...tree, nodes: updateNodes(tree.nodes, id, fn) };
	}
	function insertAfter(afterId: string) {
		const level = mkLevel();
		tree = { ...tree, nodes: insertLevelAfterDeep(tree.nodes, afterId, level) };
		selectedId = level.id;
	}
	function addLevel(ownerId: string) {
		const level = mkLevel();
		tree = appendNode(tree, ownerId, level);
		selectedId = level.id;
	}
	function addBranch(ownerId: string) {
		const branch = mkBranch('priority');
		tree = appendNode(tree, ownerId, branch);
		selectedId = branch.id;
		toast('Branch added. Fill in each lane, then pick the condition.');
	}
	function requestDelete(id: string) {
		const node = findNode(tree.nodes, id);
		if (node?.type === 'branch') {
			pendingDelete = id;
			return;
		}
		tree = { ...tree, nodes: removeNodeDeep(tree.nodes, id) };
	}
	function confirmDelete() {
		if (pendingDelete) tree = { ...tree, nodes: removeNodeDeep(tree.nodes, pendingDelete) };
		pendingDelete = null;
		toast('Branch removed.');
	}
</script>

<svelte:window
	onkeydown={(event) => {
		if (event.key === 'Escape' && selectedId) selectedId = null;
	}}
/>

<div class="flex flex-col gap-3.5">
	<a
		href={backHref}
		class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[12.5px] transition-colors"
	>
		<ArrowLeftIcon class="size-3.5" />
		{backHref.endsWith('/escalation-policies') ? 'All policies' : 'Back to policy'}
	</a>

	<div class="flex items-center gap-2.5">
		<div class="min-w-0">
			<h2 class="text-[18px] font-semibold tracking-[-0.01em]">Escalation path</h2>
			<p class="text-subtle-foreground mt-0.5 text-[12.5px]">
				Build who gets paged, in what order, and where it stops.
			</p>
		</div>
		<div class="flex-1"></div>
		<Button variant="ghost" size="sm" href={backHref}>Cancel</Button>
		<form
			method="POST"
			action="?/save"
			use:enhance={() =>
				async ({ result }) => {
					if (result.type === 'failure') {
						toast.error(String(result.data?.error ?? 'Could not save the policy.'));
						return;
					}
					if (result.type !== 'success') return;
					toast.success('Policy saved. Applies to the next escalation.');
					await goto(ws(`/escalation-policies/${result.data?.id}`));
				}}
		>
			<input type="hidden" name="tree" value={treeJson} />
			<Button type="submit" size="sm" disabled={blocked}>
				<CheckIcon data-icon="inline-start" />
				Save policy
			</Button>
		</form>
	</div>

	{#if analysis.reachValid === 0}
		<Alert.Root tone="critical">
			<Alert.Content>
				<Alert.Title>This policy can never notify anyone</Alert.Title>
				<Alert.Description>
					Every path is empty or targets only people who can't be paged. Add at least one reachable
					target.
				</Alert.Description>
			</Alert.Content>
		</Alert.Root>
	{/if}

	<div class="grid items-start gap-3.5 min-[1000px]:[grid-template-columns:minmax(0,1fr)_340px]">
		<div class="bg-card min-h-[320px] overflow-x-auto rounded-xl border px-3 py-4">
			<div class="mb-1 flex flex-wrap gap-x-[18px] gap-y-1.5 border-b px-1.5 pt-0.5 pb-3.5">
				<span class="text-subtle-foreground inline-flex items-center gap-1.5 text-[11.5px]">
					<MousePointerClickIcon class="size-3" /> Click a step to edit who it pages
				</span>
				<span class="text-subtle-foreground inline-flex items-center gap-1.5 text-[11.5px]">
					<ClockIcon class="size-3" /> Click a wait to change its timing
				</span>
				<span class="text-subtle-foreground inline-flex items-center gap-1.5 text-[11.5px]">
					<PlusIcon class="size-3" /> Hover a connector to insert a step
				</span>
			</div>
			<FlowCanvas
				{tree}
				editable
				{selectedId}
				onselect={(id) => (selectedId = id)}
				oninsertafter={insertAfter}
				onaddlevel={addLevel}
				onaddbranch={addBranch}
				ondeletenode={requestDelete}
				onchangewait={(id, wait) => update(id, (level) => ({ ...level, wait }))}
			/>
		</div>

		<aside
			class="bg-card sticky top-2 flex max-h-[calc(100vh-96px)] flex-col gap-3 self-start overflow-y-auto rounded-xl border p-[15px]"
		>
			{#if !selected}
				<PolicyInspector {tree} {analysis} {directory} onsetpolicy={(patch) => (tree = { ...tree, ...patch })} />
			{:else if selected.type === 'branch'}
				<BranchInspector
					node={selected}
					onchangecondition={(id, on) => (tree = { ...tree, nodes: changeCondition(tree.nodes, id, on) })}
					onsethours={(id, hours) =>
						(tree = {
							...tree,
							nodes: tree.nodes.map(function patch(node): typeof node {
								if (node.type === 'branch') {
									if (node.id === id) return { ...node, hours };
									return {
										...node,
										lanes: node.lanes.map((lane) => ({ ...lane, nodes: lane.nodes.map(patch) }))
									};
								}
								return node;
							})
						})}
					onremove={requestDelete}
					ondeselect={() => (selectedId = null)}
				/>
			{:else}
				<LevelInspector
					{directory}
					node={selected}
					sib={nodeSiblings(tree.nodes, selected.id)}
					onupdate={update}
					onmove={(id, dir) => (tree = { ...tree, nodes: moveNodeDeep(tree.nodes, id, dir) })}
					onremove={requestDelete}
					ondeselect={() => (selectedId = null)}
				/>
			{/if}
		</aside>
	</div>
</div>

<Dialog.Root open={!!pendingDelete} onOpenChange={(open) => (open ? null : (pendingDelete = null))}>
	<Dialog.Content class="sm:max-w-[440px]">
		<div class="flex flex-col gap-3 p-6">
			<div class="flex items-start gap-3">
				<span class="bg-critical-wash text-critical-ink flex size-[38px] shrink-0 items-center justify-center rounded-lg">
					<OctagonAlertIcon class="size-5" />
				</span>
				<div class="flex flex-1 flex-col gap-1">
					<Dialog.Title class="tracking-heading text-xl font-semibold">Delete this branch?</Dialog.Title>
					<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
						Both lanes and every level inside them will be removed. This can't be undone.
					</Dialog.Description>
				</div>
			</div>
		</div>
		<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
			<Button type="button" variant="ghost" onclick={() => (pendingDelete = null)}>Keep it</Button>
			<Button type="button" variant="destructive" onclick={confirmDelete}>Delete branch</Button>
		</div>
	</Dialog.Content>
</Dialog.Root>
