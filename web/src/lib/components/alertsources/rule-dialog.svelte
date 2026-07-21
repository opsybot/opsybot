<script lang="ts">
	import GitBranchIcon from '@lucide/svelte/icons/git-branch';
	import XIcon from '@lucide/svelte/icons/x';
	import { toast } from 'svelte-sonner';
	import { untrack } from 'svelte';
	import { enhance } from '$app/forms';
	import WfSelect from '$lib/components/workflows/wf-select.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Input } from '$lib/components/ui/input';
	import { RT_FIELDS, RT_OPS, RT_POLICIES, type RoutingRule } from '$lib/alertsources';

	let {
		open = $bindable(false),
		initial,
		rulesCount
	}: { open?: boolean; initial: RoutingRule | null; rulesCount: number } = $props();

	const cid = () => Math.random().toString(36).slice(2, 8);

	let conditions = $state<{ id: string; field: string; op: string; value: string }[]>([]);
	let policy = $state('payments-primary');
	let position = $state('end');

	$effect(() => {
		if (!open) return;
		const rule = initial;
		untrack(() => {
			conditions = rule
				? rule.conditions.map((condition) => ({ id: cid(), ...condition }))
				: [{ id: cid(), field: 'service', op: 'is', value: '' }];
			policy = rule ? rule.policy : 'payments-primary';
			position = 'end';
		});
	});

	const valid = $derived(conditions.some((condition) => condition.value.trim()));

	const positionOptions = $derived([
		{ value: 'start', label: 'First, before rule 1' },
		...Array.from({ length: Math.max(0, rulesCount - 1) }, (_, index) => ({
			value: String(index + 1),
			label: `After rule ${index + 1}`
		})),
		{ value: 'end', label: 'Last, just above the default route' }
	]);

	const definition = $derived(
		JSON.stringify({
			id: initial?.id ?? null,
			conditions: conditions
				.filter((condition) => condition.value.trim())
				.map(({ field, op, value }) => ({ field, op, value })),
			policy,
			position
		})
	);
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[560px]">
		<form
			method="POST"
			action="?/saveRule"
			use:enhance={() =>
				async ({ result, update }) => {
					await update({ reset: false });
					if (result.type !== 'success') return;
					toast.success(initial ? 'Rule updated.' : 'Rule added.');
					open = false;
				}}
		>
			<input type="hidden" name="definition" value={definition} />
			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						<GitBranchIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">
							{initial ? 'Edit rule' : 'New routing rule'}
						</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							First matching rule wins. Unmatched alerts take the default route.
						</Dialog.Description>
					</div>
				</div>

				<div class="mt-1 flex flex-col gap-4">
					<div class="flex flex-col gap-2">
						<div class="text-subtle-foreground text-[11px] tracking-[0.08em] uppercase">
							Conditions: all must match
						</div>
						{#each conditions as condition, index (condition.id)}
							<div class="flex items-center gap-2">
								<WfSelect
									size="sm"
									options={RT_FIELDS}
									bind:value={condition.field}
									class="w-[150px]"
									aria-label="Condition field"
								/>
								<WfSelect
									size="sm"
									options={RT_OPS}
									bind:value={condition.op}
									class="w-[110px]"
									aria-label="Condition operator"
								/>
								<Input
									bind:value={condition.value}
									placeholder={condition.op === 'matches' ? 'eu-*' : 'payments-api'}
									aria-label="Condition value"
									class="h-[34px] flex-1 font-mono text-[12.5px]"
								/>
								{#if conditions.length > 1}
									<Button
										variant="ghost"
										size="icon-sm"
										aria-label="Remove condition"
										onclick={() => (conditions = conditions.filter((_, other) => other !== index))}
									>
										<XIcon />
									</Button>
								{/if}
							</div>
						{/each}
						<button
							type="button"
							class="text-muted-foreground hover:text-brand-foreground self-start text-[12.5px] transition-colors"
							onclick={() => conditions.push({ id: cid(), field: 'labels.env', op: 'is', value: '' })}
						>
							+ Add condition
						</button>
					</div>

					<div class="flex flex-wrap gap-2.5">
						<WfSelect
							label="Route to policy"
							options={RT_POLICIES}
							bind:value={policy}
							class="min-w-[180px] flex-1"
						/>
						{#if !initial}
							<WfSelect
								label="Position"
								options={positionOptions}
								bind:value={position}
								class="min-w-[180px] flex-1"
							/>
						{/if}
					</div>
				</div>
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (open = false)}>Cancel</Button>
				<Button type="submit" disabled={!valid}>{initial ? 'Save rule' : 'Add rule'}</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
