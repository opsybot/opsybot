<script lang="ts">
	import ChannelRow from '$lib/components/notifications/channel-row.svelte';
	import ConnectDialog from '$lib/components/notifications/connect-dialog.svelte';
	import { CHANNEL_ICONS } from '$lib/components/notifications/icons';
	import { CHANNEL_TYPES, type ChannelType } from '$lib/notifications';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let connecting = $state<ChannelType | null>(null);

	const connectable = $derived(
		CHANNEL_TYPES.filter((type) => !data.channels.some((channel) => channel.type === type.id))
	);
</script>

<div class="flex max-w-[720px] flex-col gap-3.5">
	<div class="bg-card overflow-hidden rounded-xl border">
		<header class="flex items-center gap-2 border-b px-4 py-3">
			<span class="text-[14px] font-semibold">Connected channels</span>
			<span class="text-subtle-foreground text-[12.5px]">where Opsybot can reach you</span>
		</header>
		{#if data.channels.length === 0}
			<div class="text-subtle-foreground px-4 py-8 text-center text-[13px]">
				No channels connected yet. Add one below so Opsybot can reach you.
			</div>
		{:else}
			<div>
				{#each data.channels as channel (channel.id)}
					<ChannelRow {channel} onverify={() => (connecting = channel.type)} />
				{/each}
			</div>
		{/if}
	</div>

	{#if connectable.length}
		<div class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Add a channel</span>
			</header>
			<div class="grid gap-2.5 p-3.5 [grid-template-columns:repeat(auto-fill,minmax(190px,1fr))]">
				{#each connectable as type (type.id)}
					{@const Icon = CHANNEL_ICONS[type.icon]}
					<button
						type="button"
						onclick={() => (connecting = type.id)}
						class="bg-inset hover:border-brand-edge flex flex-col items-start gap-1.5 rounded-md border p-3 text-left transition-[border-color,transform] hover:-translate-y-px"
					>
						<span
							class="bg-inset text-muted-foreground flex size-[30px] items-center justify-center rounded-sm border"
						>
							<Icon class="size-[15px]" />
						</span>
						<span class="text-[13px] font-semibold">{type.label}</span>
						<span class="text-subtle-foreground text-[11.5px] leading-[1.4]">{type.desc}</span>
					</button>
				{/each}
			</div>
		</div>
	{/if}
</div>

<ConnectDialog type={connecting} onclose={() => (connecting = null)} />
