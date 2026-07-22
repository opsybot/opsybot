<script lang="ts">
	import ArrowDownIcon from '@lucide/svelte/icons/arrow-down';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import ArrowUpIcon from '@lucide/svelte/icons/arrow-up';
	import BellIcon from '@lucide/svelte/icons/bell';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { Button } from '$lib/components/ui/button';
	import * as Select from '$lib/components/ui/select';
	import { Separator } from '$lib/components/ui/separator';
	import TargetRow from '$lib/components/escalation/target-row.svelte';
	import {
		TARGET_TYPES,
		WAIT_OPTIONS,
		targetOptions,
		type Directory,
		type Level,
		type NotifyMode,
		type TargetType
	} from '$lib/escalation';

	let {
		node,
		sib,
		directory,
		onupdate,
		onmove,
		onremove,
		ondeselect
	}: {
		node: Level;
		sib: { canUp: boolean; canDown: boolean } | null;
		directory: Directory;
		onupdate: (id: string, fn: (level: Level) => Level) => void;
		onmove: (id: string, dir: -1 | 1) => void;
		onremove: (id: string) => void;
		ondeselect: () => void;
	} = $props();

	const options = $derived.by(() => {
		const already = new Set(node.targets.map((t) => `${t.type}:${t.ref}`));
		return targetOptions(directory, node.addType).filter(
			(option) => !option.invalid && !already.has(`${node.addType}:${option.ref}`)
		);
	});
	const waitLabel = $derived(WAIT_OPTIONS.find((w) => w.value === node.wait)?.label);

	const MODES: { value: NotifyMode; label: string; hint: string }[] = [
		{ value: 'all', label: 'All at once', hint: 'Everyone in this level is paged together' },
		{ value: 'rr', label: 'Round-robin', hint: 'One target per escalation, rotating' }
	];

	function onModeKey(event: KeyboardEvent, index: number) {
		if (!['ArrowUp', 'ArrowLeft', 'ArrowDown', 'ArrowRight'].includes(event.key)) return;
		event.preventDefault();
		const dir = event.key === 'ArrowUp' || event.key === 'ArrowLeft' ? -1 : 1;
		const nextIndex = (index + dir + MODES.length) % MODES.length;
		onupdate(node.id, (level) => ({ ...level, mode: MODES[nextIndex].value }));
		const group = (event.currentTarget as HTMLElement).parentElement;
		(group?.children[nextIndex] as HTMLElement | undefined)?.focus();
	}
</script>

<button
	type="button"
	class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[11.5px] transition-colors"
	onclick={ondeselect}
>
	<ArrowLeftIcon class="size-3" />
	Policy settings
</button>

<div class="flex items-center gap-2">
	<div class="flex flex-1 items-center gap-2 text-[14px] font-semibold">
		<span class="text-muted-foreground flex size-[22px] shrink-0 items-center justify-center rounded-sm border bg-[var(--ink-4)]">
			<BellIcon class="size-[13px]" />
		</span>
		Notify level
	</div>
	<Button variant="ghost" size="icon-sm" aria-label="Move up" disabled={!sib?.canUp} onclick={() => onmove(node.id, -1)}>
		<ArrowUpIcon />
	</Button>
	<Button variant="ghost" size="icon-sm" aria-label="Move down" disabled={!sib?.canDown} onclick={() => onmove(node.id, 1)}>
		<ArrowDownIcon />
	</Button>
</div>

<div>
	<div class="text-subtle-foreground mb-2.5 text-[11px] tracking-[0.08em] uppercase">Who to page</div>
	{#if node.targets.length === 0}
		<div class="text-critical-ink mb-2.5 flex items-center gap-1.5 text-[12px]">
			<TriangleAlertIcon class="size-[13px]" />
			This level notifies no one.
		</div>
	{:else}
		<div class="mb-2.5 flex flex-col gap-1.5">
			{#each node.targets as target, index (index)}
				<TargetRow
					{target}
					onremove={() => onupdate(node.id, (level) => ({ ...level, targets: level.targets.filter((_, other) => other !== index) }))}
				/>
			{/each}
		</div>
	{/if}
	<div class="flex gap-2">
		<Select.Root
			type="single"
			value={node.addType}
			onValueChange={(value) => onupdate(node.id, (level) => ({ ...level, addType: value as TargetType }))}
		>
			<Select.Trigger size="sm" class="w-[116px]" aria-label="Target type">
				{TARGET_TYPES.find((t) => t.value === node.addType)?.label}
			</Select.Trigger>
			<Select.Content>
				<Select.Group>
					{#each TARGET_TYPES as type (type.value)}
						<Select.Item value={type.value} label={type.label}>{type.label}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
		<Select.Root
			type="single"
			value=""
			onValueChange={(ref) => {
				const option = options.find((entry) => entry.ref === ref);
				if (option) {
					onupdate(node.id, (level) => ({
						...level,
						targets: [...level.targets, { type: level.addType, ref: option.ref, value: option.label }]
					}));
				}
			}}
		>
			<Select.Trigger size="sm" class="flex-1" aria-label="Add {node.addType}">Add {node.addType}…</Select.Trigger>
			<Select.Content>
				<Select.Group>
					{#each options as option (option.ref)}
						<Select.Item value={option.ref} label={option.label}>{option.label}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
	</div>
</div>

<Separator />

<div>
	<div class="text-subtle-foreground mb-2.5 text-[11px] tracking-[0.08em] uppercase">Notify mode</div>
	<div role="radiogroup" aria-label="Notify mode" class="flex flex-col gap-2.5">
		{#each MODES as option, index (option.value)}
			<button
				type="button"
				role="radio"
				aria-checked={node.mode === option.value}
				tabindex={node.mode === option.value ? 0 : -1}
				class="flex items-start gap-2.5 text-left"
				onclick={() => onupdate(node.id, (level) => ({ ...level, mode: option.value }))}
				onkeydown={(event) => onModeKey(event, index)}
			>
				<span
					class="mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-full border {node.mode ===
					option.value
						? 'border-primary'
						: 'border-border-strong'}"
				>
					{#if node.mode === option.value}
						<span class="bg-primary size-2 rounded-full"></span>
					{/if}
				</span>
				<div>
					<div class="text-[13px] font-medium">{option.label}</div>
					<div class="text-subtle-foreground text-[12px]">{option.hint}</div>
				</div>
			</button>
		{/each}
	</div>
</div>

<Separator />

<div>
	<div class="text-subtle-foreground mb-2.5 text-[11px] tracking-[0.08em] uppercase">
		Then wait for acknowledgement
	</div>
	<Select.Root type="single" value={node.wait} onValueChange={(value) => onupdate(node.id, (level) => ({ ...level, wait: value }))}>
		<Select.Trigger size="sm" class="w-full" aria-label="Wait for acknowledgement">{waitLabel}</Select.Trigger>
		<Select.Content>
			<Select.Group>
				{#each WAIT_OPTIONS as option (option.value)}
					<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
				{/each}
			</Select.Group>
		</Select.Content>
	</Select.Root>
	<p class="text-subtle-foreground mt-2 text-[11.5px] leading-[1.5]">
		If no one acknowledges within this window, the escalation continues to the next step.
	</p>
</div>

<Separator />

<Button
	variant="ghost"
	size="sm"
	class="text-critical-ink hover:text-critical-ink self-start"
	onclick={() => onremove(node.id)}
>
	<Trash2Icon data-icon="inline-start" />
	Delete this level
</Button>
