<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import Page from '$lib/components/layout/page.svelte';
	import DisconnectDialog from '$lib/components/chat/disconnect-dialog.svelte';
	import InstallDialog from '$lib/components/chat/install-dialog.svelte';
	import PlatformCard from '$lib/components/chat/platform-card.svelte';
	import { ws } from '$lib/navigation';
	import type { Platform } from '$lib/chat';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let installing = $state<Platform | null>(null);
	let disconnecting = $state<Platform | null>(null);

	const OAUTH_ERRORS: Record<string, string> = {
		invalid_state: 'That sign-in link expired. Start the connection again.',
		forbidden: 'Your permission to manage chat connections changed before the install finished.',
		exchange_failed: 'The provider rejected the install. Try connecting again.',
		not_configured: 'This provider is not configured on the server yet.',
		secret_unavailable: 'Secret storage is not configured, so the token could not be saved.',
		denied: 'The install was cancelled.',
		error: 'The connection could not be completed.'
	};

	function labelFor(id: string): string {
		return data.platforms.find((platform) => platform.id === id)?.label ?? 'Chat';
	}

	$effect(() => {
		const params = page.url.searchParams;
		const connected = params.get('connected');
		const linked = params.get('linked');
		const failed = params.get('chat_error');
		if (!connected && !linked && !failed) return;
		if (connected) toast.success(`${labelFor(connected)} is connected.`);
		else if (linked) toast.success(`Linked your ${labelFor(linked)} account.`);
		else if (failed) toast.error(OAUTH_ERRORS[failed] ?? OAUTH_ERRORS.error);
		goto(ws('/chat'), { replaceState: true, noScroll: true, keepFocus: true, invalidateAll: true });
	});
</script>

<Page title="Chat connections" subtitle="Incidents run where your team already talks">
	<div class="flex max-w-[760px] flex-col gap-3.5">
		{#each data.platforms as platform (platform.id)}
			<PlatformCard
				{platform}
				oninstall={() => (installing = platform)}
				ondisconnect={() => (disconnecting = platform)}
			/>
		{/each}
	</div>

	<InstallDialog platform={installing} onclose={() => (installing = null)} />
	<DisconnectDialog platform={disconnecting} onclose={() => (disconnecting = null)} />
</Page>
