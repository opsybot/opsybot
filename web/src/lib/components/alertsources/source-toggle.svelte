<script lang="ts">
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Switch } from '$lib/components/ui/switch';

	let { id, paused }: { id: string; paused: boolean } = $props();

	let checked = $state(untrack(() => !paused));
	$effect(() => {
		checked = !paused;
	});

	let form: HTMLFormElement;
</script>

<div class="flex items-center gap-[9px]">
	<span class="text-[12.5px] {paused ? 'text-warning-ink' : 'text-muted-foreground'}">
		{paused ? 'Paused — events are dropped' : 'Receiving events'}
	</span>
	<form
		method="POST"
		action="?/toggle"
		use:enhance={() =>
			async ({ result, update }) => {
				await update({ reset: false });
				if (result.type !== 'success') {
					checked = !paused;
					return;
				}
				if (paused) {
					toast.warning('Integration paused. Incoming events are dropped, not queued.');
				} else {
					toast.success('Integration resumed.');
				}
			}}
		bind:this={form}
	>
		<input type="hidden" name="id" value={id} />
		<Switch
			bind:checked
			aria-label={checked ? 'Pause this source' : 'Resume this source'}
			onCheckedChange={() => form.requestSubmit()}
		/>
	</form>
</div>
