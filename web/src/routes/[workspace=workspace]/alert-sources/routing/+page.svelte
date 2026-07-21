<script lang="ts">
	import AnchorIcon from '@lucide/svelte/icons/anchor';
	import ArrowRightIcon from '@lucide/svelte/icons/arrow-right';
	import CheckIcon from '@lucide/svelte/icons/check';
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import ChevronUpIcon from '@lucide/svelte/icons/chevron-up';
	import FlaskConicalIcon from '@lucide/svelte/icons/flask-conical';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { tick, untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import CondChips from '$lib/components/alertsources/cond-chips.svelte';
	import RuleDialog from '$lib/components/alertsources/rule-dialog.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Select from '$lib/components/ui/select';
	import { Textarea } from '$lib/components/ui/textarea';
	import { RT_POLICIES, evaluateSample, type RoutingRule } from '$lib/alertsources';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let sample = $state(untrack(() => data.sample));
	const test = $derived(evaluateSample(sample, data.rules, data.defaultPolicy));

	let dialogOpen = $state(false);
	let editing = $state<RoutingRule | null>(null);

	function openNew() {
		editing = null;
		dialogOpen = true;
	}
	function openEdit(rule: RoutingRule) {
		editing = rule;
		dialogOpen = true;
	}

	let defaultPolicy = $state(untrack(() => data.defaultPolicy));
	$effect(() => {
		defaultPolicy = data.defaultPolicy;
	});
	let defaultForm: HTMLFormElement;
	async function changeDefault(policy: string) {
		defaultPolicy = policy;
		await tick();
		defaultForm.requestSubmit();
	}
</script>

<div class="grid items-start gap-3.5 min-[1100px]:[grid-template-columns:minmax(0,1fr)_340px]">
	<div class="flex min-w-0 flex-col gap-2.5">
		<div class="flex items-center">
			<span class="text-subtle-foreground text-[13px]">
				Evaluated top to bottom: first match wins
			</span>
			<div class="flex-1"></div>
			<Button size="sm" onclick={openNew}>
				<PlusIcon data-icon="inline-start" />
				New rule
			</Button>
		</div>

		{#each data.rules as rule, index (rule.id)}
			{@const hit = test.ok && test.index === index}
			<div
				data-rule={rule.id}
				class="bg-card flex items-start gap-3 rounded-xl border px-[14px] py-[13px] transition-colors {hit
					? 'border-brand-edge shadow-[var(--glow-brand)]'
					: ''}"
			>
				<span
					class="bg-inset text-muted-foreground mt-px flex size-[22px] shrink-0 items-center justify-center rounded-full border font-mono text-[11px] font-semibold"
				>
					{index + 1}
				</span>
				<div class="flex min-w-0 flex-1 flex-col gap-[7px]">
					<CondChips conditions={rule.conditions} />
					<div class="flex items-center gap-[7px] text-[12px]">
						<ArrowRightIcon class="text-subtle-foreground size-3" />
						<span class="font-mono text-[12px]">{rule.policy}</span>
						{#if hit}
							<Badge tone="brand" size="sm" dot>matches test alert</Badge>
						{/if}
					</div>
				</div>
				<div class="flex shrink-0 gap-0.5">
					<form method="POST" action="?/moveRule" use:enhance={() => async ({ update }) => update({ reset: false })}>
						<input type="hidden" name="id" value={rule.id} />
						<input type="hidden" name="dir" value="up" />
						<Button type="submit" variant="ghost" size="icon-sm" aria-label="Move up" disabled={index === 0}>
							<ChevronUpIcon />
						</Button>
					</form>
					<form method="POST" action="?/moveRule" use:enhance={() => async ({ update }) => update({ reset: false })}>
						<input type="hidden" name="id" value={rule.id} />
						<input type="hidden" name="dir" value="down" />
						<Button type="submit" variant="ghost" size="icon-sm" aria-label="Move down" disabled={index === data.rules.length - 1}>
							<ChevronDownIcon />
						</Button>
					</form>
					<Button variant="ghost" size="icon-sm" aria-label="Edit rule" onclick={() => openEdit(rule)}>
						<PencilIcon />
					</Button>
					<form
						method="POST"
						action="?/deleteRule"
						use:enhance={() =>
							async ({ result, update }) => {
								await update({ reset: false });
								if (result.type === 'success') toast.success('Rule removed.');
							}}
					>
						<input type="hidden" name="id" value={rule.id} />
						<Button type="submit" variant="ghost" size="icon-sm" aria-label="Delete rule">
							<Trash2Icon />
						</Button>
					</form>
				</div>
			</div>
		{/each}

		<div
			class="bg-card flex items-start gap-3 rounded-xl border border-dashed px-[14px] py-[13px] transition-colors {test.ok &&
			test.index === -1
				? 'border-brand-edge shadow-[var(--glow-brand)]'
				: ''}"
		>
			<span
				class="bg-inset text-muted-foreground mt-px flex size-[22px] shrink-0 items-center justify-center rounded-full border"
			>
				<AnchorIcon class="size-3" />
			</span>
			<div class="flex min-w-0 flex-1 flex-col gap-1.5">
				<div class="flex flex-wrap items-center gap-2">
					<span class="text-[13px] font-semibold">Default route</span>
					<span class="text-subtle-foreground text-[11.5px]">
						always last: catches everything unmatched
					</span>
					{#if test.ok && test.index === -1}
						<Badge tone="brand" size="sm" dot>matches test alert</Badge>
					{/if}
				</div>
				<div class="flex items-center gap-2">
					<ArrowRightIcon class="text-subtle-foreground size-3 shrink-0" />
					<form method="POST" action="?/setDefault" bind:this={defaultForm} use:enhance={() =>
						async ({ result, update }) => {
							await update({ reset: false });
							if (result.type === 'success') toast.success(`Default route now targets ${defaultPolicy}.`);
						}}
					>
						<input type="hidden" name="policy" value={defaultPolicy} />
						<Select.Root type="single" value={defaultPolicy} onValueChange={changeDefault}>
							<Select.Trigger size="sm" class="w-[210px]" aria-label="Default route policy">
								{defaultPolicy}
							</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each RT_POLICIES as policy (policy)}
										<Select.Item value={policy} label={policy}>{policy}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</form>
				</div>
			</div>
		</div>
	</div>

	<div class="bg-card overflow-hidden rounded-xl border">
		<header class="flex items-center gap-2 border-b px-4 py-3">
			<FlaskConicalIcon class="text-subtle-foreground size-3.5" />
			<span class="text-[13.5px] font-semibold">Test an alert</span>
		</header>
		<div class="flex flex-col gap-2.5 p-[14px]">
			<Textarea bind:value={sample} rows={11} class="font-mono text-[11.5px] leading-[1.6]" aria-label="Sample alert JSON" />
			<div role="status" aria-live="polite">
				{#if test.ok}
					<div
						class="bg-brand-wash border-brand-edge text-muted-foreground flex items-center gap-2 rounded-md border px-[11px] py-[9px]"
					>
						<CheckIcon class="text-primary size-[13px] shrink-0" />
						<span class="text-[12.5px]">
							{test.index === -1 ? 'No rule matches: default route' : `Matches rule ${test.index + 1}`}
							→ <span class="font-mono text-[12px]">{test.policy}</span>
						</span>
					</div>
				{:else}
					<div
						class="bg-warning-wash border-warning-edge text-muted-foreground flex items-center gap-2 rounded-md border px-[11px] py-[9px]"
					>
						<TriangleAlertIcon class="text-warning-ink size-[13px] shrink-0" />
						<span class="text-[12.5px]">{test.error}</span>
					</div>
				{/if}
			</div>
			<p class="text-subtle-foreground m-0 text-[11.5px] leading-[1.5]">
				Edit the JSON: the result updates as you type. Nothing is paged.
			</p>
		</div>
	</div>
</div>

<RuleDialog bind:open={dialogOpen} initial={editing} rulesCount={data.rules.length} />
