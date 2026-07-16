<script lang="ts">
	import type { Snippet } from 'svelte';
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';

	let {
		open = $bindable(false),
		tone = 'warning',
		title,
		action,
		confirmLabel,
		children
	}: {
		open?: boolean;
		tone?: 'warning' | 'critical';
		title: string;
		action: string;
		confirmLabel: string;
		children: Snippet;
	} = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[440px]">
		<form
			method="POST"
			{action}
			use:enhance={() =>
				async ({ update }) => {
					await update();
					open = false;
				}}
		>
			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span
						class="flex size-[38px] shrink-0 items-center justify-center rounded-lg {tone === 'critical'
							? 'bg-critical-wash text-critical-ink'
							: 'bg-warning-wash text-warning-ink'}"
					>
						{#if tone === 'critical'}
							<OctagonAlertIcon class="size-5" />
						{:else}
							<TriangleAlertIcon class="size-5" />
						{/if}
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">{title}</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							{@render children()}
						</Dialog.Description>
					</div>
				</div>
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (open = false)}>Cancel</Button>
				<Button type="submit" variant={tone === 'critical' ? 'destructive' : 'default'}>
					{confirmLabel}
				</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
