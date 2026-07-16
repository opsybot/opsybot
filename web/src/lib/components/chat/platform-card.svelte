<script lang="ts">
	import PlusIcon from '@lucide/svelte/icons/plus';
	import RotateCwIcon from '@lucide/svelte/icons/rotate-cw';
	import UnplugIcon from '@lucide/svelte/icons/unplug';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { CHAT_ICONS } from '$lib/components/chat/icons';
	import ScopeList from '$lib/components/chat/scope-list.svelte';
	import WorkspaceDefaults from '$lib/components/chat/workspace-defaults.svelte';
	import { connectionBadge, type Platform } from '$lib/chat';

	let {
		platform,
		oninstall,
		ondisconnect
	}: { platform: Platform; oninstall: () => void; ondisconnect: () => void } = $props();

	const connection = $derived(platform.connection);
	const badge = $derived(connectionBadge(platform));
	const Icon = $derived(CHAT_ICONS[platform.icon]);
</script>

<section
	class="bg-card overflow-hidden rounded-xl border"
	data-platform={platform.id}
	data-connected={connection ? 'true' : 'false'}
>
	<header class="flex items-center gap-3 px-4 py-[14px]">
		<span
			class="bg-inset text-muted-foreground flex size-[34px] shrink-0 items-center justify-center rounded-sm border"
		>
			<Icon class="size-4" />
		</span>
		<div class="min-w-0 flex-1">
			<div class="flex items-center gap-2">
				<span class="text-[14px] font-semibold">{platform.label}</span>
				<Badge tone={badge.tone} size="sm" dot={badge.dot}>{badge.label}</Badge>
			</div>
			{#if connection}
				<div class="text-subtle-foreground mt-0.5 truncate font-mono text-[11px]">
					{connection.workspace} · {connection.healthNote}
				</div>
			{:else}
				<div class="text-subtle-foreground mt-0.5 text-[12px]">{platform.tagline}</div>
			{/if}
		</div>
		{#if connection}
			<div class="flex shrink-0 gap-2">
				<form
					method="POST"
					action="?/reconnect"
					use:enhance={() =>
						async ({ result, update }) => {
							await update({ reset: false });
							if (result.type === 'success') toast.success(`${platform.label} reconnected — scopes refreshed.`);
						}}
				>
					<input type="hidden" name="platform" value={platform.id} />
					<Button type="submit" size="sm" variant="ghost">
						<RotateCwIcon data-icon="inline-start" />
						Reconnect
					</Button>
				</form>
				<Button size="sm" variant="ghost" onclick={ondisconnect}>
					<UnplugIcon data-icon="inline-start" />
					Disconnect
				</Button>
			</div>
		{:else}
			<Button size="sm" class="shrink-0" onclick={oninstall}>
				<PlusIcon data-icon="inline-start" />
				Connect
			</Button>
		{/if}
	</header>

	{#if connection}
		<div class="border-t px-4 py-[14px]">
			<div class="text-subtle-foreground mb-2.5 text-[11px] tracking-[0.08em] uppercase">
				What Opsybot can do here
			</div>
			<ScopeList scopes={platform.scopes} />
		</div>
		<WorkspaceDefaults {platform} />
	{/if}
</section>
