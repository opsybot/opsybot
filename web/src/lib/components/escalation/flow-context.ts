import type { Tree } from '$lib/escalation';

export type FlowContext = {
	tree: Tree;
	editable: boolean;
	selectedId: string | null;
	activeId: string | null;
	laneChoices: Record<string, string> | null;
	onSelect?: (id: string) => void;
	onInsertAfter?: (afterId: string) => void;
	onAddLevel?: (ownerId: string) => void;
	onAddBranch?: (ownerId: string) => void;
	onDeleteNode?: (id: string) => void;
	onChangeWait?: (id: string, wait: string) => void;
};
