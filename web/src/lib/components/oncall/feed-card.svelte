<script lang="ts">
	import CheckIcon from '@lucide/svelte/icons/check';
	import CopyIcon from '@lucide/svelte/icons/copy';
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';

	let { url }: { url: string } = $props();

	let copied = $state(false);

	async function copy() {
		try {
			await navigator.clipboard.writeText(url);
		} catch {
			toast.error('The browser would not let us copy. Select the URL and copy it by hand.');
			return;
		}

		copied = true;
		toast.success('Feed URL copied. It works in Google Calendar and Apple Calendar.');
		setTimeout(() => (copied = false), 2000);
	}
</script>

<section class="bg-card overflow-hidden rounded-xl border">
	<header class="flex items-center gap-2.5 border-b px-4 py-3">
		<span class="text-[13.5px] font-semibold">Calendar feed</span>
	</header>

	<div class="flex flex-col gap-2 px-3.5 py-3">
		<code
			class="bg-inset text-muted-foreground rounded-md border px-2.5 py-2 font-mono text-[11px] [overflow-wrap:anywhere]"
		>
			{url}
		</code>
		<p class="text-subtle-foreground m-0 text-xs">
			Anyone with this URL can read the schedule. Treat it as a secret.
		</p>
		<Button variant="secondary" size="sm" class="self-start" onclick={copy}>
			{#if copied}
				<CheckIcon data-icon="inline-start" />
				Copied
			{:else}
				<CopyIcon data-icon="inline-start" />
				Copy URL
			{/if}
		</Button>
	</div>
</section>
