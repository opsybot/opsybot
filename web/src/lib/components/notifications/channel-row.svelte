<script lang="ts">
	import { onDestroy } from 'svelte';
	import SendIcon from '@lucide/svelte/icons/send';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { CHANNEL_ICONS } from '$lib/components/notifications/icons';
	import { channelMeta, verifyBadge, type Channel } from '$lib/notifications';

	let { channel, onverify }: { channel: Channel; onverify: () => void } = $props();

	const meta = $derived(channelMeta(channel.type));
	const Icon = $derived(CHANNEL_ICONS[meta.icon]);
	const badge = $derived(verifyBadge(channel));

	let testing = $state(false);
	let timer: ReturnType<typeof setTimeout>;
	function sendTest() {
		testing = true;
		timer = setTimeout(() => {
			testing = false;
			toast.success(`Test page sent to ${meta.label} — delivered in 0.8 s.`);
		}, 1200);
	}
	onDestroy(() => clearTimeout(timer));
</script>

<div
	class="flex items-center gap-3 border-t px-4 py-3 first:border-t-0"
	data-channel={channel.type}
	data-verified={channel.verified ? 'true' : 'false'}
>
	<span
		class="bg-inset text-muted-foreground flex size-[30px] shrink-0 items-center justify-center rounded-sm border"
	>
		<Icon class="size-[15px]" />
	</span>
	<div class="min-w-0 flex-1">
		<div class="flex items-center gap-2">
			<span class="text-[13.5px] font-medium">{meta.label}</span>
			<Badge tone={badge.tone} size="sm" dot>{badge.label}</Badge>
		</div>
		<div class="text-subtle-foreground mt-0.5 truncate font-mono text-[11.5px]">{channel.detail}</div>
	</div>
	{#if channel.verified}
		<Button size="sm" variant="secondary" disabled={testing} onclick={sendTest}>
			<SendIcon data-icon="inline-start" />
			{testing ? 'Sending…' : 'Send test'}
		</Button>
	{:else}
		<Button size="sm" variant="secondary" onclick={onverify}>Verify</Button>
	{/if}
	<form
		method="POST"
		action="?/remove"
		use:enhance={() => async ({ result, update }) => {
			await update({ reset: false });
			if (result.type === 'success') toast(`${meta.label} removed. Rules using it skip to the next step.`);
		}}
	>
		<input type="hidden" name="id" value={channel.id} />
		<Button type="submit" variant="ghost" size="icon-sm" aria-label="Remove {meta.label}">
			<Trash2Icon />
		</Button>
	</form>
</div>
