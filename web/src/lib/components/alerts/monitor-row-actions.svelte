<script lang="ts">
	import CopyIcon from '@lucide/svelte/icons/copy';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import type { Heartbeat } from '$lib/alerts';

	let { monitor, onedit }: { monitor: Heartbeat; onedit: (monitor: Heartbeat) => void } = $props();

	async function copyUrl() {
		await navigator.clipboard.writeText(monitor.checkInUrl);
		toast.success('Check-in URL copied.');
	}
</script>

<div class="flex items-center justify-end gap-1">
	<Button size="icon-sm" variant="ghost" onclick={copyUrl} aria-label="Copy check-in URL for {monitor.name}">
		<CopyIcon />
	</Button>
	<Button
		size="icon-sm"
		variant="ghost"
		onclick={() => onedit(monitor)}
		aria-label="Edit {monitor.name}"
	>
		<PencilIcon />
	</Button>
	<form
		method="POST"
		action="?/delete"
		use:enhance={() => async ({ result, update }) => {
			await update({ invalidateAll: true });
			if (result.type === 'failure') {
				toast.error(String(result.data?.error ?? 'Could not delete that monitor.'));
				return;
			}
			if (result.type === 'success') toast.success(`${monitor.name} deleted.`);
		}}
	>
		<input type="hidden" name="id" value={monitor.id} />
		<Button type="submit" size="icon-sm" variant="ghost" aria-label="Delete {monitor.name}">
			<Trash2Icon />
		</Button>
	</form>
</div>
