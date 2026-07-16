<script lang="ts">
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import SlidersHorizontalIcon from '@lucide/svelte/icons/sliders-horizontal';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { Separator } from '$lib/components/ui/separator';
	import { REPEAT_OPTIONS, TEAMS, type Analysis, type Tree } from '$lib/escalation';

	let {
		tree,
		analysis,
		onsetpolicy
	}: { tree: Tree; analysis: Analysis; onsetpolicy: (patch: Partial<Tree>) => void } = $props();

	const warns = $derived(
		[
			analysis.emptyLevel && 'A level notifies no one.',
			analysis.unreachableLevel && 'A level targets only people who can’t be paged.',
			analysis.hasDeadEnd && 'A branch lane has no levels — alerts routed there page no one.',
			analysis.deactivated && 'A deactivated user is still targeted.'
		].filter(Boolean) as string[]
	);

	const repeatLabel = $derived(REPEAT_OPTIONS.find((r) => r.value === tree.repeat)?.label);
</script>

<div class="flex items-center gap-2 text-[14px] font-semibold">
	<span class="text-muted-foreground flex size-[22px] shrink-0 items-center justify-center rounded-sm border bg-[var(--ink-4)]">
		<SlidersHorizontalIcon class="size-[13px]" />
	</span>
	Policy
</div>

<div>
	<div class="text-subtle-foreground mb-2.5 text-[11px] tracking-[0.08em] uppercase">Details</div>
	<div class="flex flex-col gap-2.5">
		<div class="flex flex-col gap-1.5">
			<span class="text-muted-foreground text-[13px] font-medium">Name</span>
			<Input value={tree.name} oninput={(e) => onsetpolicy({ name: e.currentTarget.value })} class="font-mono" aria-label="Policy name" />
		</div>
		<div class="flex flex-col gap-1.5">
			<span class="text-muted-foreground text-[13px] font-medium">Team</span>
			<Select.Root type="single" value={tree.team} onValueChange={(v) => onsetpolicy({ team: v })}>
				<Select.Trigger size="sm" class="w-full" aria-label="Team">{tree.team}</Select.Trigger>
				<Select.Content>
					<Select.Group>
						{#each TEAMS as team (team)}
							<Select.Item value={team} label={team}>{team}</Select.Item>
						{/each}
					</Select.Group>
				</Select.Content>
			</Select.Root>
		</div>
		<div class="flex flex-col gap-1.5">
			<span class="text-muted-foreground text-[13px] font-medium">Repeat if never acknowledged</span>
			<Select.Root type="single" value={tree.repeat} onValueChange={(v) => onsetpolicy({ repeat: v })}>
				<Select.Trigger size="sm" class="w-full" aria-label="Repeat if never acknowledged">{repeatLabel}</Select.Trigger>
				<Select.Content>
					<Select.Group>
						{#each REPEAT_OPTIONS as option (option.value)}
							<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
						{/each}
					</Select.Group>
				</Select.Content>
			</Select.Root>
		</div>
	</div>
</div>

<Separator />

<div>
	<div class="text-subtle-foreground mb-2.5 text-[11px] tracking-[0.08em] uppercase">Coverage</div>
	<div class="grid grid-cols-2 gap-3">
		<div>
			<div class="font-mono text-[20px] leading-none font-semibold">{analysis.reach}</div>
			<div class="text-subtle-foreground mt-[3px] text-[10.5px]">targets reached</div>
		</div>
		<div>
			<div class="font-mono text-[20px] leading-none font-semibold">{analysis.maxLevels}</div>
			<div class="text-subtle-foreground mt-[3px] text-[10.5px]">levels, deepest path</div>
		</div>
		<div>
			<div class="font-mono text-[20px] leading-none font-semibold">{analysis.maxMin}m</div>
			<div class="text-subtle-foreground mt-[3px] text-[10.5px]">to fully escalate</div>
		</div>
		<div>
			<div class="font-mono text-[20px] leading-none font-semibold">{analysis.branches}</div>
			<div class="text-subtle-foreground mt-[3px] text-[10.5px]">branches</div>
		</div>
	</div>
</div>

{#if warns.length}
	<div class="border-critical-edge bg-critical-wash rounded-md border px-3 py-2.5 text-[12px] leading-[1.5]">
		<div class="text-critical-ink mb-1 flex items-center gap-1.5 font-semibold">
			<TriangleAlertIcon class="size-[13px]" />
			Fix before saving
		</div>
		<ul class="m-0 flex list-disc flex-col gap-[3px] pl-[18px]">
			{#each warns as warn (warn)}
				<li>{warn}</li>
			{/each}
		</ul>
	</div>
{:else}
	<div class="bg-inset text-muted-foreground flex items-center gap-[7px] rounded-md border px-3 py-[9px] text-[12.5px]">
		<ShieldCheckIcon class="size-[13px]" />
		Every path reaches someone.
	</div>
{/if}

<p class="text-subtle-foreground m-0 text-[11.5px] leading-[1.5]">
	Select any node on the canvas to edit who it pages, how, and how long it waits.
</p>
