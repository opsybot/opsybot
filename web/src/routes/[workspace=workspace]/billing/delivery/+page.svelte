<script lang="ts">
	import type { Component } from 'svelte';
	import { tick, untrack } from 'svelte';
	import type { LucideProps } from '@lucide/svelte';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import RadioIcon from '@lucide/svelte/icons/radio';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import SmartphoneIcon from '@lucide/svelte/icons/smartphone';
	import UnplugIcon from '@lucide/svelte/icons/unplug';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import { Switch } from '$lib/components/ui/switch';
	import { formatCount } from '$lib/billing';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const USAGE_ICON: Record<string, Component<LucideProps>> = {
		'message-square': MessageSquareIcon,
		phone: PhoneIcon,
		smartphone: SmartphoneIcon
	};

	let channels = $state(untrack(() => data.channels.map((channel) => ({ ...channel }))));
	let forms: Record<string, HTMLFormElement> = {};

	async function toggle(index: number, on: boolean) {
		channels[index].on = on;
		await tick();
		forms[channels[index].id]?.requestSubmit();
	}
</script>

<div class="flex max-w-[720px] flex-col gap-3.5">
	<div class="bg-card flex items-center gap-3 rounded-xl border p-4">
		<span class="bg-inset text-muted-foreground flex size-[34px] shrink-0 items-center justify-center rounded-sm border">
			<RadioIcon class="size-4" />
		</span>
		<div class="flex-1">
			<div class="text-[13.5px] font-semibold">Managed delivery {data.linked ? 'connected' : 'not connected'}</div>
			<div class="text-subtle-foreground mt-px text-[12px]">
				{data.linked
					? "Account acme-corp · SMS and voice route through Opsybot's carriers."
					: 'Self-hosted push works standalone. Link an account for SMS and voice without running your own gateways.'}
			</div>
		</div>
		{#if data.linked}
			<form
				method="POST"
				action="?/disconnect"
				use:enhance={() => async ({ result, update }) => {
					await update({ reset: false });
					if (result.type !== 'success') {
						toast.error('Could not disconnect the delivery bridge.');
						return;
					}
					toast('Delivery bridge disconnected. Push still works locally.');
				}}
			>
				<Button size="sm" variant="ghost" type="submit">
					<UnplugIcon data-icon="inline-start" />
					Disconnect
				</Button>
			</form>
		{:else}
			<form
				method="POST"
				action="?/connect"
				use:enhance={() => async ({ result, update }) => {
					await update({ reset: false });
					if (result.type !== 'success') {
						toast.error('Could not link the delivery account.');
						return;
					}
					toast.success('Delivery account linked.');
				}}
			>
				<Button size="sm" type="submit">
					<PlusIcon data-icon="inline-start" />
					Connect account
				</Button>
			</form>
		{/if}
	</div>

	{#if data.linked}
		<div class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Channels</span>
			</header>
			<div>
				{#each channels as channel, index (channel.id)}
					<div class="flex items-center gap-3 border-t px-4 py-[11px] first:border-t-0" data-channel={channel.id}>
						<div class="flex-1">
							<div class="text-[13px] font-medium">{channel.label}</div>
							<div class="text-subtle-foreground mt-px font-mono text-[10.5px]">transits: {channel.transit}</div>
						</div>
						<form
							method="POST"
							action="?/toggle"
							bind:this={forms[channel.id]}
							use:enhance={() => async ({ result, update }) => {
								await update({ reset: false });
								if (result.type !== 'success') {
									channels[index].on = data.channels[index].on;
									toast.error('Could not update the channel.');
								}
							}}
						>
							<input type="hidden" name="id" value={channel.id} />
							<input type="hidden" name="on" value={channels[index].on} />
							<Switch checked={channels[index].on} onCheckedChange={(value) => toggle(index, value)} aria-label="{channel.label} delivery" />
						</form>
					</div>
				{/each}
			</div>
		</div>

		<Alert.Root tone="info">
			<ShieldCheckIcon />
			<Alert.Content>
				<Alert.Title>What leaves your infrastructure</Alert.Title>
				<Alert.Description>
					Only the minimum to deliver a page: destination (phone number), a short message, and an incident ID for
					delivery receipts. No timelines, no alert payloads, no customer data. Push tokens never leave your instance.
				</Alert.Description>
			</Alert.Content>
		</Alert.Root>

		<div class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Usage this period</span>
			</header>
			<div>
				{#each data.usage as meter (meter.kind)}
					{@const Icon = USAGE_ICON[meter.icon]}
					<div class="flex items-center gap-3 border-t px-4 py-[11px] first:border-t-0" data-usage={meter.kind}>
						<Icon class="text-subtle-foreground size-[14px] shrink-0" />
						<span class="flex-1 text-[13px]">{meter.kind}</span>
						<span class="font-mono text-[12.5px]">{formatCount(meter.used)} sent</span>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>
