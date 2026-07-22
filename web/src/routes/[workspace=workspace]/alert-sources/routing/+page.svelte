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
	import { onDestroy, onMount, untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import CondChips from '$lib/components/alertsources/cond-chips.svelte';
	import GroupRules from '$lib/components/alertsources/group-rules.svelte';
	import PolicyField from '$lib/components/alertsources/policy-field.svelte';
	import RuleDialog from '$lib/components/alertsources/rule-dialog.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Textarea } from '$lib/components/ui/textarea';
	import { type RoutingRule } from '$lib/alertsources';
	import type { RoutePreview } from '$lib/server/alertsources';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let sample = $state(untrack(() => data.sample));
	let preview = $state<RoutePreview | null>(null);
	let previewError = $state('');
	let previewForm: HTMLFormElement;
	let debounce: ReturnType<typeof setTimeout> | null = null;

	const matchedIndex = $derived(
		preview?.matchedRouteId ? data.rules.findIndex((rule) => rule.id === preview?.matchedRouteId) : -1
	);

	function evaluate() {
		if (debounce) clearTimeout(debounce);
		debounce = setTimeout(() => previewForm.requestSubmit(), 400);
	}

	onMount(evaluate);
	onDestroy(() => {
		if (debounce) clearTimeout(debounce);
	});

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
			{@const hit = matchedIndex === index}
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
			class="bg-card flex items-start gap-3 rounded-xl border border-dashed px-[14px] py-[13px] transition-colors {!!preview &&
			matchedIndex === -1
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
					{#if preview && matchedIndex === -1}
						<Badge tone="brand" size="sm" dot>matches test alert</Badge>
					{/if}
				</div>
				<div class="flex items-center gap-2">
					<ArrowRightIcon class="text-subtle-foreground size-3 shrink-0" />
					<form
						method="POST"
						action="?/setDefault"
						class="flex items-end gap-2"
						use:enhance={() =>
							async ({ result, update }) => {
								await update({ reset: false });
								if (result.type === 'success') toast.success(`Default route now targets ${defaultPolicy}.`);
							}}
					>
						<PolicyField
							id="default-policy"
							label=""
							known={data.knownPolicies}
							bind:value={defaultPolicy}
							class="w-[210px]"
						/>
						<input type="hidden" name="policy" value={defaultPolicy} />
						<Button type="submit" size="sm" variant="secondary" disabled={defaultPolicy === data.defaultPolicy}>
							Save
						</Button>
					</form>
				</div>
			</div>
		</div>
	</div>

	<div class="flex flex-col gap-3.5">
	<GroupRules rules={data.groupRules} />

	<div class="bg-card overflow-hidden rounded-xl border">
		<header class="flex items-center gap-2 border-b px-4 py-3">
			<FlaskConicalIcon class="text-subtle-foreground size-3.5" />
			<span class="text-[13.5px] font-semibold">Test an alert</span>
		</header>
		<div class="flex flex-col gap-2.5 p-[14px]">
			<Textarea bind:value={sample} oninput={evaluate} rows={11} class="font-mono text-[11.5px] leading-[1.6]" aria-label="Sample alert JSON" />
			<div role="status" aria-live="polite">
				{#if preview}
					<div
						class="bg-brand-wash border-brand-edge text-muted-foreground flex flex-col gap-1 rounded-md border px-[11px] py-[9px]"
					>
						<span class="flex items-center gap-2 text-[12.5px]">
							<CheckIcon class="text-primary size-[13px] shrink-0" />
							{matchedIndex === -1 ? 'No rule matches: default route' : `Matches rule ${matchedIndex + 1}`}
							→ <span class="font-mono text-[12px]">{preview.policyRef}</span>
						</span>
						{#if preview.groupFields.length}
							<span class="text-subtle-foreground pl-[21px] text-[11.5px]">
								Groups by <span class="font-mono">{preview.groupFields.join(', ')}</span>
							</span>
						{/if}
					</div>
				{:else if previewError}
					<div
						class="bg-warning-wash border-warning-edge text-muted-foreground flex items-center gap-2 rounded-md border px-[11px] py-[9px]"
					>
						<TriangleAlertIcon class="text-warning-ink size-[13px] shrink-0" />
						<span class="text-[12.5px]">{previewError}</span>
					</div>
				{/if}
			</div>
			<p class="text-subtle-foreground m-0 text-[11.5px] leading-[1.5]">
				Edit the JSON and Opsybot evaluates it with the same engine that routes real alerts. Nothing
				is paged.
			</p>
			<form
				method="POST"
				action="?/preview"
				bind:this={previewForm}
				class="hidden"
				use:enhance={() => async ({ result }) => {
					if (result.type === 'failure') {
						preview = null;
						previewError = String(result.data?.error ?? 'Could not evaluate that sample.');
						return;
					}
					if (result.type !== 'success') return;
					previewError = '';
					preview = result.data?.preview as RoutePreview;
				}}
			>
				<input type="hidden" name="payload" value={sample} />
			</form>
		</div>
	</div>
	</div>
</div>

<RuleDialog
	bind:open={dialogOpen}
	initial={editing}
	rulesCount={data.rules.length}
	knownPolicies={data.knownPolicies}
	defaultPolicy={data.defaultPolicy}
/>
