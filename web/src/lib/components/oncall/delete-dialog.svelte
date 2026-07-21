<script lang="ts">
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';

	let { open = $bindable(false), name }: { open?: boolean; name: string } = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[440px]">
		<form
			method="POST"
			action="?/delete"
			use:enhance={() => async ({ result, update }) => {
				await update();
				if (result.type === 'failure') {
					open = false;
					toast.error(String(result.data?.error ?? 'Could not delete the schedule.'));
				}
			}}
		>
			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-critical-wash text-critical-ink flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						<OctagonAlertIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">
							Delete {name}?
						</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							Its layers, participants, and overrides go too. This cannot be undone. Restore it
							instead to keep the history.
						</Dialog.Description>
					</div>
				</div>
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (open = false)}>Cancel</Button>
				<Button type="submit" variant="destructive">Delete schedule</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
