<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import CheckIcon from '@lucide/svelte/icons/check';
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import ChevronUpIcon from '@lucide/svelte/icons/chevron-up';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import XIcon from '@lucide/svelte/icons/x';
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import ActionConfig from '$lib/components/workflows/action-config.svelte';
	import WfSelect from '$lib/components/workflows/wf-select.svelte';
	import { ICON } from '$lib/components/workflows/icons';
	import * as Alert from '$lib/components/ui/alert';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { ws } from '$lib/navigation';
	import {
		ACTION_META,
		ACTION_TYPES,
		CONDITION_FIELDS,
		TRIGGERS,
		defaultConfig,
		loops,
		shortId,
		webhookMissingUrl,
		type ActionType,
		type Condition,
		type TriggerType,
		type WorkflowAction
	} from '$lib/workflows';

	type Initial = {
		name: string;
		trigger: TriggerType;
		conditions: Condition[];
		actions: WorkflowAction[];
	} | null;

	let {
		initial,
		editing,
		fromTemplate,
		roleNames
	}: { initial: Initial; editing: boolean; fromTemplate: boolean; roleNames: string[] } = $props();

	// Seeded once; an existing workflow loads verbatim, including an empty condition list
	let name = $state(untrack(() => initial?.name ?? ''));
	let trigger = $state<TriggerType>(untrack(() => initial?.trigger ?? 'declared'));
	let conditions = $state(
		untrack(() =>
			(initial ? initial.conditions : [{ field: 'severity', value: 'SEV1' }]).map((condition) => ({
				id: shortId('c'),
				field: condition.field,
				value: condition.value
			}))
		)
	);
	let actions = $state(
		untrack(() =>
			(initial
				? initial.actions
				: [{ id: shortId('a'), type: 'post' as ActionType, config: defaultConfig('post') }]
			).map((action) => ({
				id: action.id || shortId('a'),
				type: action.type,
				config: { ...defaultConfig(action.type), ...action.config }
			}))
		)
	);

	const loop = $derived(loops(trigger, actions));
	const noActions = $derived(actions.length === 0);
	const anyEmptyWebhook = $derived(actions.some(webhookMissingUrl));
	const canSave = $derived(name.trim().length > 0 && !loop && !noActions && !anyEmptyWebhook);

	const definition = $derived(
		JSON.stringify({
			name: name.trim(),
			trigger,
			conditions: conditions.map(({ field, value }) => ({ field, value })),
			actions: actions.map(({ id, type, config }) => ({ id, type, config }))
		})
	);

	function conditionPlaceholder(field: string): string {
		if (field === 'severity') return 'SEV1';
		if (field === 'service') return 'payments-api';
		return 'true';
	}

	function moveAction(index: number, by: -1 | 1) {
		const to = index + by;
		if (to < 0 || to >= actions.length) return;
		[actions[index], actions[to]] = [actions[to], actions[index]];
	}

	function addAction(type: ActionType) {
		actions.push({ id: shortId('a'), type, config: defaultConfig(type) });
	}
</script>

<form
	method="POST"
	action="?/save"
	use:enhance={() =>
		async ({ result }) => {
			if (result.type === 'failure') {
				toast.error(String(result.data?.error ?? 'Could not save the workflow.'));
				return;
			}
			if (result.type !== 'success') return;
			toast.success(
				editing
					? `“${name.trim()}” saved.`
					: `“${name.trim()}” saved${fromTemplate ? ' from template' : ''}. It starts disabled — flip it on when ready.`
			);
			await goto(ws('/workflows'));
		}}
	class="flex max-w-[760px] flex-col gap-3.5"
>
	<input type="hidden" name="definition" value={definition} />

	<a
		href={ws('/workflows')}
		class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[12.5px] transition-colors"
	>
		<ArrowLeftIcon class="size-3.5" />
		Workflows
	</a>

	<h2 class="text-[18px] font-semibold tracking-[-0.01em]">
		{editing ? 'Edit workflow' : 'New workflow'}
		{#if fromTemplate}
			<Badge tone="brand" size="sm" class="ml-2.5 align-middle">from template</Badge>
		{/if}
	</h2>

	<div class="bg-card rounded-xl border p-4">
		<Field.Field class="max-w-[360px] gap-1.5 space-y-0">
			<Field.FieldLabel for="wf-name" class="text-muted-foreground text-[13px] font-medium">
				Name
			</Field.FieldLabel>
			<Input id="wf-name" bind:value={name} placeholder="SEV1 comms cadence" />
		</Field.Field>
	</div>

	<div class="bg-card flex flex-col gap-3 rounded-xl border p-4">
		<div class="text-subtle-foreground -mb-0.5 text-[11px] tracking-[0.08em] uppercase">When</div>
		<WfSelect options={TRIGGERS} bind:value={trigger} class="max-w-[280px]" aria-label="Trigger" />

		<div class="text-subtle-foreground -mb-0.5 text-[11px] tracking-[0.08em] uppercase">
			Only if — all conditions match
		</div>
		{#each conditions as condition, i (condition.id)}
			<div class="flex items-center gap-2">
				<WfSelect
					size="sm"
					options={CONDITION_FIELDS}
					bind:value={condition.field}
					class="w-[150px]"
					aria-label="Condition field"
				/>
				<span class="text-subtle-foreground font-mono text-[12px]">is</span>
				<Input
					bind:value={condition.value}
					placeholder={conditionPlaceholder(condition.field)}
					aria-label="Condition value"
					class="h-[34px] max-w-[220px] flex-1 text-[13px]"
				/>
				{#if conditions.length > 1}
					<Button
						variant="ghost"
						size="icon-sm"
						aria-label="Remove condition"
						onclick={() => (conditions = conditions.filter((_, j) => j !== i))}
					>
						<XIcon />
					</Button>
				{/if}
			</div>
		{/each}
		<button
			type="button"
			class="text-muted-foreground hover:text-brand-foreground self-start text-[12.5px] transition-colors"
			onclick={() => conditions.push({ id: shortId('c'), field: 'service', value: '' })}
		>
			+ Add condition
		</button>
	</div>

	<div class="text-subtle-foreground text-[11px] tracking-[0.08em] uppercase">Then, in order</div>

	{#if loop}
		<Alert.Root tone="critical">
			<Alert.Content>
				<Alert.Title>This workflow would loop</Alert.Title>
				<Alert.Description>
					It fires on “update overdue” and adds a timeline note — which counts as activity and re-arms
					the overdue timer, firing it again. Remove the note action or change the trigger.
				</Alert.Description>
			</Alert.Content>
		</Alert.Root>
	{/if}
	{#if noActions}
		<Alert.Root tone="critical">
			<Alert.Content>
				<Alert.Description>A workflow needs at least one action.</Alert.Description>
			</Alert.Content>
		</Alert.Root>
	{/if}

	{#each actions as action, i (action.id)}
		{@const Meta = ICON[ACTION_META[action.type].icon]}
		<div
			data-action={action.type}
			class="bg-card overflow-hidden rounded-xl border {action.type === 'note' && loop
				? 'border-critical-edge'
				: ''}"
		>
			<header class="bg-inset flex items-center gap-[9px] border-b px-3.5 py-2.5">
				<span
					class="bg-brand-wash border-brand-edge text-brand-foreground flex size-5 shrink-0 items-center justify-center rounded-full border font-mono text-[11px] font-semibold"
				>
					{i + 1}
				</span>
				<Meta class="text-muted-foreground size-3.5" />
				<span class="text-[13px] font-semibold">{ACTION_META[action.type].label}</span>
				{#if webhookMissingUrl(action)}
					<Badge tone="critical" size="sm">URL required</Badge>
				{/if}
				<div class="flex-1"></div>
				<Button
					variant="ghost"
					size="icon-sm"
					aria-label="Move up"
					disabled={i === 0}
					onclick={() => moveAction(i, -1)}
				>
					<ChevronUpIcon />
				</Button>
				<Button
					variant="ghost"
					size="icon-sm"
					aria-label="Move down"
					disabled={i === actions.length - 1}
					onclick={() => moveAction(i, 1)}
				>
					<ChevronDownIcon />
				</Button>
				<Button
					variant="ghost"
					size="icon-sm"
					aria-label="Remove action"
					onclick={() => (actions = actions.filter((_, j) => j !== i))}
				>
					<Trash2Icon />
				</Button>
			</header>
			<div class="p-3.5">
				<ActionConfig {action} {roleNames} />
			</div>
		</div>
	{/each}

	<div class="flex flex-wrap gap-1.5">
		{#each ACTION_TYPES as type (type.value)}
			{@const Chip = ICON[type.icon]}
			<button
				type="button"
				class="border-border-strong text-muted-foreground hover:border-brand-edge hover:text-brand-foreground inline-flex items-center gap-1.5 rounded-full border border-dashed px-3 py-1.5 text-[12px] transition-colors"
				onclick={() => addAction(type.value)}
			>
				<Chip class="size-3" />
				{type.label}
			</button>
		{/each}
	</div>

	<div class="flex gap-2.5 pt-1">
		<Button type="submit" disabled={!canSave}>
			<CheckIcon data-icon="inline-start" />
			Save workflow
		</Button>
		<Button variant="ghost" href={ws('/workflows')}>Cancel</Button>
	</div>
</form>
