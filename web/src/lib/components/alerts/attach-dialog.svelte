<script lang="ts">
	import LinkIcon from '@lucide/svelte/icons/link';
	import { enhance } from '$app/forms';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import { SEVERITY_TONE } from '$lib/dashboard';
	import type { Severity } from '$lib/dashboard';
	import { cn } from '$lib/utils';

	let {
		open = $bindable(false),
		incidents
	}: {
		open?: boolean;
		incidents: { id: string; name: string; severity: Severity }[];
	} = $props();

	let picked = $state('');
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[460px]">
		<form
			method="POST"
			action="?/attach"
			use:enhance={() => async ({ update }) => {
				await update();
				open = false;
				picked = '';
			}}
		>
			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						<LinkIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">
							Attach to an existing incident
						</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							The alert and its timeline are added to the incident's context.
						</Dialog.Description>
					</div>
				</div>

				<div class="mt-1 flex flex-col gap-2">
					{#each incidents as incident (incident.id)}
						<button
							type="button"
							onclick={() => (picked = incident.id)}
							class={cn(
								'bg-inset text-foreground flex w-full items-center gap-2.5 rounded-md border px-3 py-2.5 text-left text-[13px]',
								picked === incident.id
									? 'border-brand-edge bg-brand-wash'
									: 'hover:border-border-strong'
							)}
						>
							<Badge tone={SEVERITY_TONE[incident.severity]} size="sm">{incident.severity}</Badge>
							<span class="font-medium">{incident.name}</span>
							<span class="text-subtle-foreground ml-auto font-mono text-[11.5px]">
								{incident.id}
							</span>
						</button>
					{/each}
					<input type="hidden" name="incident" value={picked} />
				</div>
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (open = false)}>Cancel</Button>
				<Button type="submit" disabled={!picked}>Attach</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
