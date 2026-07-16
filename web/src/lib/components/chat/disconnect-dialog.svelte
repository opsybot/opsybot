<script lang="ts">
	import { untrack } from 'svelte';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import type { Platform } from '$lib/chat';

	let { platform, onclose }: { platform: Platform | null; onclose: () => void } = $props();

	// Hold the last platform so the title survives the close animation
	let current = $state<Platform | null>(null);
	const open = $derived(!!platform);

	$effect(() => {
		if (platform) untrack(() => (current = platform));
	});
</script>

<Dialog.Root {open} onOpenChange={(value) => (value ? undefined : onclose())}>
	<Dialog.Content class="sm:max-w-[440px]">
		{#if current}
			<form
				method="POST"
				action="?/disconnect"
				use:enhance={() =>
					async ({ result, update }) => {
						await update({ reset: false });
						const label = current?.label;
						onclose();
						if (result.type === 'success' && label) toast(`${label} disconnected.`);
					}}
			>
				<input type="hidden" name="platform" value={current.id} />
				<div class="flex flex-col gap-3 p-6">
					<div class="flex items-start gap-3">
						<span
							class="bg-critical-wash text-critical-ink flex size-[38px] shrink-0 items-center justify-center rounded-lg"
						>
							<TriangleAlertIcon class="size-5" />
						</span>
						<div class="flex flex-1 flex-col gap-1">
							<Dialog.Title class="tracking-heading text-xl font-semibold"
								>Disconnect {current.label}?</Dialog.Title
							>
							<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
								Active incidents lose their chat rooms and members stop getting DMs on this platform.
								Captured timelines stay in Opsybot.
							</Dialog.Description>
						</div>
					</div>
				</div>
				<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
					<Button type="button" variant="ghost" onclick={onclose}>Cancel</Button>
					<Button type="submit" variant="destructive">Disconnect</Button>
				</div>
			</form>
		{/if}
	</Dialog.Content>
</Dialog.Root>
