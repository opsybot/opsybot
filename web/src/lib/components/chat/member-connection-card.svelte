<script lang="ts">
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import LinkIcon from '@lucide/svelte/icons/link';
	import SendIcon from '@lucide/svelte/icons/send';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { CHAT_ICONS } from '$lib/components/chat/icons';
	import { linkBadge, type Platform } from '$lib/chat';

	let { platform }: { platform: Platform } = $props();

	const connection = $derived(platform.connection);
	const badge = $derived(linkBadge(connection));
	const Icon = $derived(CHAT_ICONS[platform.icon]);
	const linked = $derived(connection?.linked ?? false);
	const linkLabel = $derived(
		linked ? (connection?.linkedHandle ? connection.linkedHandle : 'Linked') : 'Link my account'
	);
	const detail = $derived.by(() => {
		if (!linked) return platform.tagline;
		const who = connection?.linkedHandle ? `Linked as ${connection.linkedHandle}` : 'Linked';
		return connection?.linkMethod === 'email' ? `${who} · matched by your email` : who;
	});

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
	data-linked={linked ? 'true' : 'false'}
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
			<div class="text-subtle-foreground mt-0.5 truncate text-[12px]">{detail}</div>
		</div>
		<div class="flex shrink-0 gap-2">
			<form
				method="POST"
				action="?/test"
				use:enhance={() =>
					async ({ result, update }) => {
						await update({ reset: false });
						if (result.type === 'success') {
							const message = String(result.data?.detail ?? '');
							if (result.data?.delivered) {
								toast.success(message || `Sent a test message on ${platform.label}.`);
							} else {
								toast.error(message || `Could not send a test message on ${platform.label}.`);
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
					<Button type="submit" size="sm" variant="ghost" title={linked ? 'Re-link your account' : undefined}>
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
					<Button type="submit" size="sm" variant="ghost" title={linked ? 'Re-link your account' : undefined}>
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
					<Button type="submit" size="sm" variant="ghost" title={linked ? 'Re-link your account' : undefined}>
						{#if linked}<CircleCheckIcon data-icon="inline-start" class="text-success-ink" />{:else}<LinkIcon
								data-icon="inline-start"
							/>{/if}
						{linkLabel}
					</Button>
				</form>
			{/if}
		</div>
	</header>
</section>
