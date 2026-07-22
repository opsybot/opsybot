<script lang="ts">
	import LayersIcon from '@lucide/svelte/icons/layers';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import WfSelect from '$lib/components/workflows/wf-select.svelte';
	import { Button } from '$lib/components/ui/button';
	import { GROUP_WINDOWS, RT_FIELDS, type GroupRule } from '$lib/alertsources';

	let { rules }: { rules: GroupRule[] } = $props();

	const cid = () => Math.random().toString(36).slice(2, 8);

	type Draft = { key: string; fields: string[]; windowSeconds: string };

	let drafts = $state<Draft[]>([]);

	$effect(() => {
		const saved = rules;
		untrack(() => {
			drafts = saved.map((rule) => ({
				key: cid(),
				fields: [...rule.fields],
				windowSeconds: String(rule.windowSeconds)
			}));
		});
	});

	const payload = $derived(
		JSON.stringify(
			drafts
				.filter((draft) => draft.fields.length)
				.map((draft) => ({ fields: draft.fields, windowSeconds: Number(draft.windowSeconds) }))
		)
	);

	function addRule() {
		drafts.push({ key: cid(), fields: ['service'], windowSeconds: '300' });
	}

	function toggleField(draft: Draft, field: string) {
		const index = draft.fields.indexOf(field);
		if (index === -1) draft.fields.push(field);
		else draft.fields.splice(index, 1);
	}
</script>

<div class="bg-card overflow-hidden rounded-xl border">
	<header class="flex items-center gap-2 border-b px-4 py-3">
		<LayersIcon class="text-subtle-foreground size-3.5" />
		<span class="text-[13.5px] font-semibold">Grouping</span>
		<div class="flex-1"></div>
		<Button size="sm" variant="secondary" onclick={addRule}>
			<PlusIcon data-icon="inline-start" />
			Add rule
		</Button>
	</header>

	<form
		method="POST"
		action="?/saveGroups"
		use:enhance={() => async ({ result, update }) => {
			await update({ reset: false, invalidateAll: true });
			if (result.type === 'failure') {
				toast.error(String(result.data?.error ?? 'Could not save those grouping rules.'));
				return;
			}
			if (result.type === 'success') toast.success('Grouping rules saved.');
		}}
		class="flex flex-col gap-3 p-[14px]"
	>
		<p class="text-subtle-foreground m-0 text-[12.5px] leading-[1.55]">
			Alerts that share every chosen field collapse into one parent inside the window. Only the
			parent routes, so a burst pages once.
		</p>

		{#each drafts as draft, index (draft.key)}
			<div class="bg-inset flex flex-col gap-2.5 rounded-lg border p-3">
				<div class="flex items-center gap-2">
					<span class="text-muted-foreground text-[12.5px] font-medium">Rule {index + 1}</span>
					<div class="flex-1"></div>
					<Button
						size="icon-sm"
						variant="ghost"
						onclick={() => drafts.splice(index, 1)}
						aria-label="Remove rule {index + 1}"
					>
						<Trash2Icon />
					</Button>
				</div>

				<fieldset class="flex flex-wrap items-center gap-1.5">
					<legend class="text-subtle-foreground mb-1 text-[12px]">Group when these match</legend>
					{#each RT_FIELDS as field (field)}
						<label
							class="border-input hover:border-brand-edge inline-flex cursor-pointer items-center gap-1.5 rounded-full border px-2.5 py-1 font-mono text-[11.5px] transition-colors {draft.fields.includes(
								field
							)
								? 'border-brand-edge bg-brand-wash text-brand-foreground'
								: 'text-muted-foreground'}"
						>
							<input
								type="checkbox"
								class="sr-only"
								checked={draft.fields.includes(field)}
								onchange={() => toggleField(draft, field)}
							/>
							{field}
						</label>
					{/each}
				</fieldset>

				<WfSelect
					label="Window"
					options={GROUP_WINDOWS}
					bind:value={draft.windowSeconds}
					size="sm"
					class="max-w-[200px]"
				/>
			</div>
		{:else}
			<p class="text-subtle-foreground m-0 text-[12.5px]">
				No grouping rules. Every alert stands on its own.
			</p>
		{/each}

		<input type="hidden" name="rules" value={payload} />
		<Button type="submit" size="sm" class="self-start">Save grouping</Button>
	</form>
</div>
