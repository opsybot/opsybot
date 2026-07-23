<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import Page from '$lib/components/layout/page.svelte';
	import AdminPlatformCard from '$lib/components/chat/admin-platform-card.svelte';
	import DisconnectDialog from '$lib/components/chat/disconnect-dialog.svelte';
	import InstallDialog from '$lib/components/chat/install-dialog.svelte';
	import { ws } from '$lib/navigation';
	import { OAUTH_ERRORS, type Platform } from '$lib/chat';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let installing = $state<Platform | null>(null);
	let disconnecting = $state<Platform | null>(null);

	function labelFor(id: string): string {
		return data.platforms.find((platform) => platform.id === id)?.label ?? 'Chat';
	}

	$effect(() => {
		const params = page.url.searchParams;
		const connected = params.get('connected');
		const failed = params.get('chat_error');
		if (!connected && !failed && !params.get('linked')) return;
		if (connected) toast.success(`${labelFor(connected)} is connected.`);
		else if (failed) toast.error(OAUTH_ERRORS[failed] ?? OAUTH_ERRORS.error);
		goto(ws('/integrations'), { replaceState: true, noScroll: true, keepFocus: true, invalidateAll: true });
	});
</script>

<Page title="Integrations" subtitle="Connect your workspace to Slack, Teams, Discord, and Telegram">
	<div class="flex max-w-[760px] flex-col gap-3.5">
		{#each data.platforms as platform (platform.id)}
			<AdminPlatformCard
				{platform}
				oninstall={() => (installing = platform)}
				ondisconnect={() => (disconnecting = platform)}
			/>
		{/each}
	</div>

	<InstallDialog platform={installing} onclose={() => (installing = null)} />
	<DisconnectDialog platform={disconnecting} onclose={() => (disconnecting = null)} />
</Page>
