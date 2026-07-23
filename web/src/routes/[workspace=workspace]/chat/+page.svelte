<script lang="ts">
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import Page from '$lib/components/layout/page.svelte';
	import MemberConnectionCard from '$lib/components/chat/member-connection-card.svelte';
	import { ws } from '$lib/navigation';
	import { OAUTH_ERRORS } from '$lib/chat';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const connected = $derived(data.platforms.filter((platform) => platform.connection));

	function labelFor(id: string): string {
		return data.platforms.find((platform) => platform.id === id)?.label ?? 'Chat';
	}

	$effect(() => {
		const params = page.url.searchParams;
		const linked = params.get('linked');
		const failed = params.get('chat_error');
		if (!linked && !failed && !params.get('connected')) return;
		if (linked) toast.success(`Linked your ${labelFor(linked)} account.`);
		else if (failed) toast.error(OAUTH_ERRORS[failed] ?? OAUTH_ERRORS.error);
		goto(ws('/chat'), { replaceState: true, noScroll: true, keepFocus: true, invalidateAll: true });
	});
</script>

<Page title="Chat connections" subtitle="Link your account so Opsybot can reach you where your team works">
	<div class="flex max-w-[720px] flex-col gap-3.5">
		{#if connected.length === 0}
			<div class="bg-card text-subtle-foreground rounded-xl border px-4 py-8 text-center text-[13px]">
				No chat providers are connected for this workspace yet. Ask an admin to connect one on the
				Integrations page.
			</div>
		{:else}
			{#each connected as platform (platform.id)}
				<MemberConnectionCard {platform} />
			{/each}
		{/if}
	</div>
</Page>
