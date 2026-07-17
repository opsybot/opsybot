<script lang="ts">
	import type { Component } from 'svelte';
	import type { LucideProps } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';

	let {
		open,
		title,
		description,
		confirmLabel = 'Confirm',
		cancelLabel = 'Cancel',
		tone = 'warning',
		icon,
		onConfirm,
		onCancel
	}: {
		open: boolean;
		title: string;
		description?: string;
		confirmLabel?: string;
		cancelLabel?: string;
		tone?: 'warning' | 'critical';
		icon?: Component<LucideProps>;
		onConfirm: () => void;
		onCancel: () => void;
	} = $props();

	const Icon = $derived(icon);
	const square = $derived(
		tone === 'critical' ? 'bg-critical-wash text-critical-ink' : 'bg-warning-wash text-warning-ink'
	);
</script>

<Dialog.Root {open} onOpenChange={(value) => (value ? undefined : onCancel())}>
	<Dialog.Content class="sm:max-w-[460px]">
		<div class="flex flex-col gap-4 p-6">
			<div class="flex items-start gap-3">
				<span class="flex size-[38px] shrink-0 items-center justify-center rounded-lg {square}">
					{#if Icon}<Icon class="size-5" />{/if}
				</span>
				<div class="flex flex-1 flex-col gap-1">
					<Dialog.Title class="tracking-heading text-lg font-semibold">{title}</Dialog.Title>
					{#if description}
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							{description}
						</Dialog.Description>
					{/if}
				</div>
			</div>
		</div>
		<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
			<Button variant="ghost" onclick={onCancel}>{cancelLabel}</Button>
			<Button variant="destructive" onclick={onConfirm}>{confirmLabel}</Button>
		</div>
	</Dialog.Content>
</Dialog.Root>
