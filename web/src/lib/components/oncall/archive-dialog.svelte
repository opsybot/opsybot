<script lang="ts">
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';

	let { open = $bindable(false), name }: { open?: boolean; name: string } = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[420px]">
		<form method="POST" action="?/archive" use:enhance>
			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-warning-wash text-warning-ink flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						<TriangleAlertIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">
							Archive {name}?
						</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							The schedule stops paging immediately. History stays readable. Escalation policies that
							reference it will flag an error.
						</Dialog.Description>
					</div>
				</div>
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (open = false)}>Cancel</Button>
				<Button type="submit">Archive schedule</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
