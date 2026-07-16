<script lang="ts">
	import CheckCircle2Icon from '@lucide/svelte/icons/circle-check-big';
	import FlaskConicalIcon from '@lucide/svelte/icons/flask-conical';
	import GitBranchIcon from '@lucide/svelte/icons/git-branch';
	import LoaderIcon from '@lucide/svelte/icons/loader';
	import PlayIcon from '@lucide/svelte/icons/play';
	import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';
	import { Button } from '$lib/components/ui/button';
	import * as Select from '$lib/components/ui/select';
	import { BRANCH_DEFS, targetInvalid, usedConditions, type Trace, type Tree } from '$lib/escalation';

	let {
		tree,
		trace,
		scenario = $bindable(),
		activeIndex = $bindable(),
		running = $bindable()
	}: {
		tree: Tree;
		trace: Trace;
		scenario: { priority: string; hours: string };
		activeIndex: number;
		running: boolean;
	} = $props();

	const conds = $derived(usedConditions(tree));
	const steps = $derived(trace.steps);
	const hasTrace = $derived(activeIndex >= 0);
	const done = $derived(activeIndex >= steps.length - 1);

	const people = $derived.by(() => {
		const set = new Set<string>();
		for (const step of steps) if (step.kind === 'level') for (const t of step.targets) set.add(`${t.type}:${t.value}`);
		return set.size;
	});

	const reduceMotion = () => {
		try {
			return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
		} catch {
			return false;
		}
	};

	$effect(() => {
		if (!running) return;
		if (reduceMotion()) {
			activeIndex = steps.length - 1;
			running = false;
			return;
		}
		const timer = setInterval(() => {
			if (activeIndex >= steps.length - 1) {
				clearInterval(timer);
				running = false;
				return;
			}
			activeIndex += 1;
		}, 950);
		return () => clearInterval(timer);
	});

	function run() {
		activeIndex = 0;
		running = true;
	}
	function reset() {
		running = false;
		activeIndex = -1;
	}
	function setCondition(id: 'priority' | 'hours', value: string) {
		scenario = { ...scenario, [id]: value };
		reset();
	}

	const fmtT = (t: number) => `${String(Math.floor(t)).padStart(2, '0')}:00`;
</script>

{#snippet traceRow(step: Trace['steps'][number], isActive: boolean)}
	{#if step.kind === 'branch'}
		{@const def = BRANCH_DEFS[step.on]}
		{@const lane = def.lanes.find((l) => l.key === step.laneKey) ?? def.lanes[0]}
		<div
			class="bg-inset motion-safe:animate-in motion-safe:fade-in-0 motion-safe:slide-in-from-top-1 flex items-start gap-2.5 border-t px-[11px] py-2 text-[12.5px] leading-[1.45] first:border-t-0 {isActive
				? '!bg-brand-wash'
				: ''}"
		>
			<span class="text-subtle-foreground flex w-[42px] shrink-0 items-center gap-1 pt-px font-mono text-[11px]">
				<GitBranchIcon class="size-3" />
			</span>
			<div>
				<span class="text-muted-foreground">No ack — routed by {def.title}: </span>
				<strong>{lane.label}</strong>
			</div>
		</div>
	{:else if step.kind === 'end'}
		<div
			class="motion-safe:animate-in motion-safe:fade-in-0 motion-safe:slide-in-from-top-1 flex items-start gap-2.5 border-t px-[11px] py-2 text-[12.5px] leading-[1.45] first:border-t-0 {isActive
				? 'bg-brand-wash'
				: ''}"
		>
			<span class="text-subtle-foreground w-[42px] shrink-0 pt-px font-mono text-[11px]">{fmtT(step.t)}</span>
			<div>
				<strong class="text-warning-ink">Exhausted.</strong>
				{step.repeat && step.repeat !== '0' ? `Repeated ×${step.repeat}, then no` : 'No'} one else is paged.
			</div>
		</div>
	{:else}
		{@const names = step.targets.map((t) => t.value)}
		{@const modeNote = step.targets.length > 1 ? (step.mode === 'rr' ? ' · round-robin' : ' · all at once') : ''}
		{@const bad = step.targets.length > 0 && step.targets.every(targetInvalid)}
		<div
			class="motion-safe:animate-in motion-safe:fade-in-0 motion-safe:slide-in-from-top-1 flex items-start gap-2.5 border-t px-[11px] py-2 text-[12.5px] leading-[1.45] first:border-t-0 {isActive
				? 'bg-brand-wash'
				: ''}"
		>
			<span class="text-subtle-foreground w-[42px] shrink-0 pt-px font-mono text-[11px]">{fmtT(step.t)}</span>
			<div>
				<span class="text-muted-foreground">Paged </span>
				<strong>{names.join(' + ')}</strong>
				<span class="text-subtle-foreground">{modeNote}</span>
				{#if bad}<span class="text-critical-ink"> · can't be paged</span>{/if}
			</div>
		</div>
	{/if}
{/snippet}

<div class="bg-card overflow-hidden rounded-xl border">
	<header class="flex items-center gap-2.5 border-b px-4 py-3">
		<span
			class="text-brand-foreground flex size-[22px] shrink-0 items-center justify-center rounded-sm border bg-[var(--ink-4)]"
		>
			<FlaskConicalIcon class="size-[13px]" />
		</span>
		<span class="text-[13.5px] font-semibold">Test this path</span>
	</header>
	<div class="flex flex-col gap-3 px-[15px] py-[13px]">
		<p class="text-subtle-foreground m-0 text-[12px] leading-[1.5]">
			Pick a scenario and watch exactly who gets paged, in order, assuming no one acknowledges.
		</p>

		{#if conds.length === 0}
			<div class="text-muted-foreground text-[12.5px]">
				This policy has no branches — every alert follows the same path.
			</div>
		{:else}
			<div class="flex flex-col gap-2.5">
				{#each conds as cond (cond)}
					{@const ctrl = BRANCH_DEFS[cond].control}
					<div class="flex flex-col gap-1.5">
						<span class="text-muted-foreground text-[13px] font-medium">{ctrl.label}</span>
						<Select.Root
							type="single"
							value={scenario[ctrl.id]}
							onValueChange={(value) => setCondition(ctrl.id, value)}
						>
							<Select.Trigger size="sm" class="w-full" aria-label={ctrl.label}>
								{ctrl.options.find((o) => o.value === scenario[ctrl.id])?.label}
							</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each ctrl.options as option (option.value)}
										<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</div>
				{/each}
			</div>
		{/if}

		<div class="flex gap-2">
			<Button size="sm" class="flex-1" onclick={run} disabled={running}>
				{#if running}
					<LoaderIcon data-icon="inline-start" class="animate-spin motion-reduce:animate-none" />
					Running…
				{:else}
					<PlayIcon data-icon="inline-start" />
					{hasTrace ? 'Run again' : 'Run trace'}
				{/if}
			</Button>
			{#if hasTrace}
				<Button size="sm" variant="ghost" onclick={reset}>
					<RotateCcwIcon data-icon="inline-start" />
					Reset
				</Button>
			{/if}
		</div>

		<div class="contents" role="status" aria-live="polite">
			{#if hasTrace}
				<div class="flex flex-col overflow-hidden rounded-md border">
					{#each steps.slice(0, activeIndex + 1) as step, index (index)}
						{@render traceRow(step, index === activeIndex && (running || !done))}
					{/each}
				</div>
			{/if}

			{#if hasTrace && done}
				<div class="text-muted-foreground flex items-center gap-2 text-[12.5px]">
					<CheckCircle2Icon class="text-primary size-[13px]" />
					Reaches <strong>{people}</strong>
					{people === 1 ? 'target' : 'targets'} · full escalation in <strong>{trace.totalMin}m</strong>
				</div>
			{/if}
		</div>
	</div>
</div>
