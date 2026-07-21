<script lang="ts">
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Switch } from '$lib/components/ui/switch';

	let { id, name, enabled }: { id: string; name: string; enabled: boolean } = $props();

	let checked = $state(untrack(() => enabled));
	$effect(() => {
		checked = enabled;
	});

	let form: HTMLFormElement;
</script>

<form
	method="POST"
	action="?/toggle"
	use:enhance={() =>
		async ({ result, update }) => {
			await update({ reset: false });
			if (result.type !== 'success') {
				checked = enabled;
				return;
			}
			if (enabled) {
				toast.success(`${name} enabled.`);
			} else {
				toast(`${name} disabled. Nothing fires until it's back on.`);
			}
		}}
	bind:this={form}
>
	<input type="hidden" name="id" value={id} />
	<Switch
		bind:checked
		aria-label={checked ? `Disable ${name}` : `Enable ${name}`}
		onCheckedChange={() => form.requestSubmit()}
	/>
</form>
