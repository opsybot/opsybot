<script lang="ts">
	import Page from '$lib/components/layout/page.svelte';
	import DisconnectDialog from '$lib/components/chat/disconnect-dialog.svelte';
	import InstallDialog from '$lib/components/chat/install-dialog.svelte';
	import PlatformCard from '$lib/components/chat/platform-card.svelte';
	import type { Platform } from '$lib/chat';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let installing = $state<Platform | null>(null);
	let disconnecting = $state<Platform | null>(null);
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
