<script lang="ts">
	import { tick, untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { Switch } from '$lib/components/ui/switch';
	import { DEFAULT_DEFAULTS, previewChannelName, type Platform } from '$lib/chat';

	let { platform }: { platform: Platform } = $props();

	const initial = untrack(() => platform.connection?.defaults ?? DEFAULT_DEFAULTS);
	let naming = $state(initial.namingPattern);
	let announce = $state(initial.announceChannel);
	let archive = $state(initial.archiveOnResolve);

	const preview = $derived(previewChannelName(naming));

	let form: HTMLFormElement;

	async function save() {
		await tick();
		form.requestSubmit();
	}
</script>

<form
	method="POST"
	action="?/saveDefaults"
	bind:this={form}
	class="flex flex-col gap-3.5 border-t px-4 py-[14px]"
	use:enhance={() =>
		async ({ result, update }) => {
			await update({ reset: false });
			if (result.type === 'failure') toast.error(String(result.data?.error ?? 'Could not save those defaults.'));
		}}
>
	<input type="hidden" name="platform" value={platform.id} />
	<input type="hidden" name="announceChannel" value={announce} />
	<input type="hidden" name="archiveOnResolve" value={String(archive)} />

	<div class="text-subtle-foreground -mb-1 text-[11px] tracking-[0.08em] uppercase">
		Defaults for this workspace
	</div>

	<div class="flex flex-wrap items-start gap-3">
		<Field.Field class="w-[230px] gap-1.5 space-y-0">
			<Field.FieldLabel for="chan-{platform.id}" class="text-muted-foreground text-[13px] font-medium">
				Incident channel naming
			</Field.FieldLabel>
			<Input
				id="chan-{platform.id}"
				name="namingPattern"
				class="font-mono text-[12.5px]"
				bind:value={naming}
				onchange={save}
			/>
			<Field.FieldDescription class="text-subtle-foreground text-xs">
				Preview: {preview}
			</Field.FieldDescription>
		</Field.Field>

		<Field.Field class="w-[230px] gap-1.5 space-y-0">
			<Field.FieldLabel for="announce-{platform.id}" class="text-muted-foreground text-[13px] font-medium">
				Default announcement channel
			</Field.FieldLabel>
			<Input
				id="announce-{platform.id}"
				class="font-mono text-[12.5px]"
				placeholder="#incidents"
				bind:value={announce}
				onchange={save}
			/>
			<Field.FieldDescription class="text-subtle-foreground text-xs">
				Where new incidents are announced in {platform.label}.
			</Field.FieldDescription>
		</Field.Field>
	</div>

	<div class="flex items-center gap-2.5">
		<Switch
			bind:checked={archive}
			aria-label="Archive the incident channel when the incident resolves"
			onCheckedChange={save}
		/>
		<span class="text-muted-foreground text-[13px]">
			Archive the incident channel when the incident resolves
		</span>
	</div>
</form>
