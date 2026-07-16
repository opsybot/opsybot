<script lang="ts">
	import SparklesIcon from '@lucide/svelte/icons/sparkles';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import { tick, untrack } from 'svelte';
	import { enhance } from '$app/forms';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { ws } from '$lib/navigation';
	import type { Factor } from '$lib/postmortems';

	let {
		factor,
		incidentId,
		readonly = false
	}: {
		factor: Factor;
		incidentId: string;
		readonly?: boolean;
	} = $props();

	// One form per field so a blur save cannot post the other field's half-typed value
	let label = $state(untrack(() => factor.label));
	let text = $state(untrack(() => factor.text));
	let labelForm = $state<HTMLFormElement | null>(null);
	let textForm = $state<HTMLFormElement | null>(null);

	async function save(form: HTMLFormElement | null) {
		await tick();
		form?.requestSubmit();
	}
</script>

<div class="bg-inset rounded-md border p-3">
	<form
		method="POST"
		action="?/factor"
		bind:this={labelForm}
		use:enhance={() => async ({ update }) => update({ reset: false })}
	>
		<input type="hidden" name="id" value={factor.id} />
		<input type="hidden" name="field" value="label" />
		<Input
			name="value"
			bind:value={label}
			{readonly}
			placeholder="The condition, in a few words — no canary stage"
			aria-label="Condition"
			class="mb-2 h-[34px] text-[13px] font-medium"
			onblur={() => save(labelForm)}
		/>
	</form>

	<form
		method="POST"
		action="?/factor"
		bind:this={textForm}
		use:enhance={() => async ({ update }) => update({ reset: false })}
	>
		<input type="hidden" name="id" value={factor.id} />
		<input type="hidden" name="field" value="text" />
		<Textarea
			name="value"
			bind:value={text}
			rows={2}
			{readonly}
			placeholder="What condition made this possible?"
			aria-label="Contributing factor"
			onblur={() => save(textForm)}
		/>
	</form>

	<div class="mt-[7px] flex flex-wrap items-center gap-2">
		{#if factor.fromTimeline.length}
			<Badge tone="brand" size="sm">
				<SparklesIcon />
				suggested by Opsybot
			</Badge>
			<span class="text-subtle-foreground text-[11.5px]">derived from:</span>
			{#each factor.fromTimeline as entry (entry)}
				<a
					href={ws(`/incidents/${incidentId}/timeline`)}
					class="border-brand-edge text-brand-foreground hover:bg-brand-wash rounded-full border px-2 py-0.5 font-mono text-[10.5px]"
				>
					{entry}
				</a>
			{/each}
		{:else}
			<span class="text-subtle-foreground text-[11.5px]">added by hand</span>
		{/if}

		<div class="flex-1"></div>

		{#if !readonly}
			<form method="POST" action="?/removeFactor" use:enhance>
				<input type="hidden" name="id" value={factor.id} />
				<Button type="submit" variant="ghost" size="icon-sm" aria-label="Remove factor">
					<Trash2Icon />
				</Button>
			</form>
		{/if}
	</div>
</div>
