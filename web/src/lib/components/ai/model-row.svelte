<script lang="ts">
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';
	import SendIcon from '@lucide/svelte/icons/send';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import type { Model } from '$lib/ai';

	let { model, isDefault }: { model: Model; isDefault: boolean } = $props();

	const ok = $derived(model.health === 'ok');

	function test() {
		if (ok) toast.success(`Test OK — ${model.latency}.`);
		else toast.error('Test failed — connection timed out after 30 s.');
	}
</script>

<div class="flex items-start gap-3 border-t px-4 py-3 first:border-t-0" data-model={model.id}>
	{#if ok}
		<CircleCheckIcon class="mt-0.5 size-3.5 shrink-0 text-[var(--success)]" />
		<span class="sr-only">Reachable.</span>
	{:else}
		<OctagonAlertIcon class="mt-0.5 size-3.5 shrink-0 text-[var(--critical)]" />
	{/if}
	<div class="min-w-0 flex-1">
		<div class="flex flex-wrap items-center gap-2">
			<span class="text-[13.5px] font-semibold">{model.name}</span>
			{#if isDefault}<Badge tone="brand" size="sm">default</Badge>{/if}
			{#if !ok}<Badge tone="critical" size="sm" dot>unreachable</Badge>{/if}
		</div>
		<div class="text-subtle-foreground mt-0.5 truncate font-mono text-[11px]">
			{model.endpoint} · last test {model.latency}
		</div>
	</div>
	{#if !isDefault && ok}
		<form method="POST" action="?/makeDefault" use:enhance={() => async ({ result, update }) => {
			await update({ reset: false });
			if (result.type === 'success') toast.success(`${model.name} is now the default model.`);
		}}>
			<input type="hidden" name="id" value={model.id} />
			<Button type="submit" size="sm" variant="ghost">Make default</Button>
		</form>
	{/if}
	<Button size="sm" variant="ghost" onclick={test}>
		<SendIcon data-icon="inline-start" />
		Test
	</Button>
	<form method="POST" action="?/remove" use:enhance={() => async ({ result, update }) => {
		await update({ reset: false });
		if (result.type === 'success') toast(`${model.name} removed.`);
	}}>
		<input type="hidden" name="id" value={model.id} />
		<Button type="submit" variant="ghost" size="icon-sm" aria-label="Remove {model.name}">
			<Trash2Icon />
		</Button>
	</form>
</div>
