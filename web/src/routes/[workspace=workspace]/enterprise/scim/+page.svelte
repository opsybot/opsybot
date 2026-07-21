<script lang="ts">
	import type { Component } from 'svelte';
	import type { LucideProps } from '@lucide/svelte';
	import ArrowRightLeftIcon from '@lucide/svelte/icons/arrow-right-left';
	import CopyIcon from '@lucide/svelte/icons/copy';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import EyeOffIcon from '@lucide/svelte/icons/eye-off';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import RotateCwIcon from '@lucide/svelte/icons/rotate-cw';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import UserPlusIcon from '@lucide/svelte/icons/user-plus';
	import UserXIcon from '@lucide/svelte/icons/user-x';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import UpgradeState from '$lib/components/enterprise/upgrade-state.svelte';
	import { ENT_PITCH, scimEventColor, type ScimEventKind } from '$lib/enterprise';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const EVENT_ICON: Record<ScimEventKind, Component<LucideProps>> = {
		sync: RotateCwIcon,
		create: UserPlusIcon,
		deprovision: UserXIcon,
		update: PencilIcon
	};

	let revealed = $state(false);
	let rotateOpen = $state(false);
</script>

{#if !data.licensed}
	<UpgradeState pitch={ENT_PITCH.scim} />
{:else if data.scim}
	{@const scim = data.scim}
	<div class="flex max-w-[760px] flex-col gap-3.5">
		<div class="bg-card flex flex-col gap-3.5 rounded-xl border p-4">
			<div>
				<div class="text-subtle-foreground mb-[7px] text-[11px] tracking-[0.08em] uppercase">SCIM 2.0 endpoint</div>
				<div class="flex items-center gap-2">
					<code class="bg-inset text-foreground flex-1 rounded-md border px-[11px] py-[9px] font-mono text-[12px] [overflow-wrap:anywhere]">{scim.endpoint}</code>
					<Button size="sm" variant="secondary" onclick={() => toast.success('SCIM endpoint copied.')}>
						<CopyIcon data-icon="inline-start" />
						Copy
					</Button>
				</div>
			</div>
			<div>
				<div class="text-subtle-foreground mb-[7px] text-[11px] tracking-[0.08em] uppercase">Bearer token</div>
				<div class="flex items-center gap-2">
					<code class="bg-inset text-foreground flex-1 rounded-md border px-[11px] py-[9px] font-mono text-[12px] [overflow-wrap:anywhere]">
						{revealed ? scim.token : scim.token.slice(0, 7) + '••••••••••••'}
					</code>
					<Button size="sm" variant="ghost" onclick={() => (revealed = !revealed)}>
						{#if revealed}<EyeOffIcon data-icon="inline-start" />Hide{:else}<EyeIcon data-icon="inline-start" />Reveal{/if}
					</Button>
					<Button size="sm" variant="secondary" onclick={() => (rotateOpen = true)}>
						<RotateCwIcon data-icon="inline-start" />
						Rotate
					</Button>
				</div>
			</div>
			<div class="flex items-center gap-2.5">
				<Badge tone="success" size="sm" dot>sync healthy</Badge>
				<span class="text-subtle-foreground font-mono text-[11.5px]">last sync {scim.lastSync}</span>
			</div>
		</div>

		<div class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Recent provisioning events</span>
			</header>
			<div>
				{#each scim.events as event (event.at)}
					{@const Icon = EVENT_ICON[event.kind]}
					<div class="flex items-start gap-3 border-t px-4 py-3 first:border-t-0" data-event={event.kind}>
						<Icon class="mt-0.5 size-3.5 shrink-0" style="color: {scimEventColor(event.tone)}" />
						<div class="min-w-0 flex-1">
							<div class="text-[13px] leading-[1.5]">{event.text}</div>
							<div class="text-subtle-foreground mt-0.5 font-mono text-[10.5px]">{event.at}</div>
							{#if event.wizard}
								<button
									type="button"
									class="text-brand-foreground mt-[5px] inline-flex items-center gap-1.5 text-[11.5px] hover:underline"
									onclick={() =>
										toast(
											'Opens the reassignment record: schedule layer, escalation step, and follow-up all moved to Sana Ito.'
										)}
								>
									<ArrowRightLeftIcon class="size-[11px]" />
									{event.wizard}: view outcome
								</button>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>
	</div>

	<Dialog.Root bind:open={rotateOpen}>
		<Dialog.Content class="sm:max-w-[440px]">
			<form
				method="POST"
				action="?/rotate"
				use:enhance={() => async ({ result, update }) => {
					if (result.type === 'failure') {
						toast.error(String(result.data?.error ?? 'Could not rotate the token.'));
						return;
					}
					if (result.type !== 'success') return;
					await update({ reset: false });
					rotateOpen = false;
					revealed = true;
					toast.success('Token rotated: update your IdP now.');
				}}
			>
				<div class="flex flex-col gap-3 p-6">
					<div class="flex items-start gap-3">
						<span class="bg-warning-wash text-warning-ink flex size-[38px] shrink-0 items-center justify-center rounded-lg">
							<TriangleAlertIcon class="size-5" />
						</span>
						<div class="flex flex-1 flex-col gap-1">
							<Dialog.Title class="tracking-heading text-xl font-semibold">Rotate the SCIM token?</Dialog.Title>
							<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
								The IdP stops syncing until you paste the new token there. The old token dies immediately.
							</Dialog.Description>
						</div>
					</div>
				</div>
				<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
					<Button type="button" variant="ghost" onclick={() => (rotateOpen = false)}>Cancel</Button>
					<Button type="submit">Rotate token</Button>
				</div>
			</form>
		</Dialog.Content>
	</Dialog.Root>
{/if}
