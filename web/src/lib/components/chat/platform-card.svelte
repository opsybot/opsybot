<script lang="ts">
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import LinkIcon from '@lucide/svelte/icons/link';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import SendIcon from '@lucide/svelte/icons/send';
	import UnplugIcon from '@lucide/svelte/icons/unplug';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
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
	const linked = $derived(connection?.linked ?? false);
	const linkLabel = $derived(
		linked ? (connection?.linkedHandle ? connection.linkedHandle : 'Linked') : 'Link my account'
	);

	let linkPoll: ReturnType<typeof setInterval> | undefined;
	function pollForLink() {
		clearInterval(linkPoll);
		let tries = 0;
		linkPoll = setInterval(async () => {
			tries += 1;
			await invalidateAll();
			if (linked) {
				clearInterval(linkPoll);
				toast.success(`${platform.label} linked!`);
			} else if (tries >= 13) {
				clearInterval(linkPoll);
			}
		}, 3000);
	}
	$effect(() => () => clearInterval(linkPoll));
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
					action="?/test"
					use:enhance={() =>
						async ({ result, update }) => {
							await update({ reset: false });
							if (result.type === 'success') {
								const detail = String(result.data?.detail ?? '');
								if (result.data?.delivered) {
									toast.success(detail || `Sent a test message on ${platform.label}.`);
								} else {
									toast.error(detail || `Could not send a test message on ${platform.label}.`);
								}
							} else if (result.type === 'failure') {
								toast.error(String(result.data?.error ?? 'Could not send a test message.'));
							}
						}}
				>
					<input type="hidden" name="platform" value={platform.id} />
					<Button type="submit" size="sm" variant="ghost">
						<SendIcon data-icon="inline-start" />
						Send test
					</Button>
				</form>
				{#if platform.authKind === 'telegram'}
					<form
						method="POST"
						action="?/linkTelegram"
						use:enhance={() =>
							async ({ result, update }) => {
								await update({ reset: false });
								if (result.type === 'success' && result.data?.telegramUrl) {
									window.open(String(result.data.telegramUrl), '_blank', 'noopener');
									toast.info('Opened Telegram — tap Start to link your account.');
									pollForLink();
								} else if (result.type === 'failure') {
									toast.error(String(result.data?.error ?? 'Could not start Telegram linking.'));
								}
							}}
					>
						<input type="hidden" name="platform" value={platform.id} />
						<Button
							type="submit"
							size="sm"
							variant="ghost"
							title={linked ? 'Re-link your account' : undefined}
						>
							{#if linked}<CircleCheckIcon data-icon="inline-start" class="text-success-ink" />{:else}<LinkIcon
									data-icon="inline-start"
								/>{/if}
							{linkLabel}
						</Button>
					</form>
				{:else if platform.authKind === 'oauth'}
					<form
						method="POST"
						action="?/linkOAuth"
						use:enhance={() =>
							async ({ result, update }) => {
								await update({ reset: false });
								if (result.type === 'success' && result.data?.oauthUrl) {
									window.location.href = String(result.data.oauthUrl);
								} else if (result.type === 'failure') {
									toast.error(String(result.data?.error ?? 'Could not start sign-in.'));
								}
							}}
					>
						<input type="hidden" name="platform" value={platform.id} />
						<Button
							type="submit"
							size="sm"
							variant="ghost"
							title={linked ? 'Re-link your account' : undefined}
						>
							{#if linked}<CircleCheckIcon data-icon="inline-start" class="text-success-ink" />{:else}<LinkIcon
									data-icon="inline-start"
								/>{/if}
							{linkLabel}
						</Button>
					</form>
				{:else}
					<form
						method="POST"
						action="?/link"
						use:enhance={() =>
							async ({ result, update }) => {
								await update({ reset: false });
								if (result.type === 'success') {
									const handle = result.data?.handle;
									toast.success(
										handle
											? `Linked your ${platform.label} account: ${handle}.`
											: `Linked your ${platform.label} account.`
									);
								} else if (result.type === 'failure') {
									toast.error(String(result.data?.error ?? 'Could not link your account.'));
								}
							}}
					>
						<input type="hidden" name="platform" value={platform.id} />
						<Button
							type="submit"
							size="sm"
							variant="ghost"
							title={linked ? 'Re-link your account' : undefined}
						>
							{#if linked}<CircleCheckIcon data-icon="inline-start" class="text-success-ink" />{:else}<LinkIcon
									data-icon="inline-start"
								/>{/if}
							{linkLabel}
						</Button>
					</form>
				{/if}
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
		{#if platform.authKind !== 'telegram'}
			<WorkspaceDefaults {platform} />
		{/if}
	{/if}
</section>
