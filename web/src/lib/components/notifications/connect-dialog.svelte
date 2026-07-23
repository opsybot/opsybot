<script lang="ts">
	import { untrack } from 'svelte';
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { CHANNEL_ICONS } from '$lib/components/notifications/icons';
	import { ws } from '$lib/navigation';
	import { channelMeta, type ChannelType } from '$lib/notifications';

	let { type, onclose }: { type: ChannelType | null; onclose: () => void } = $props();

	let current = $state<ChannelType | null>(null);
	let step = $state<'form' | 'code' | 'done'>('form');
	let channelId = $state('');
	let hint = $state('');
	let ntfyServer = $state('https://ntfy.sh');
	let ntfyTopic = $state('');
	let detail = $state('');
	let secret = $state('');
	let connectForm: HTMLFormElement;

	const open = $derived(!!type);
	const meta = $derived(current ? channelMeta(current) : null);
	const Icon = $derived(meta ? CHANNEL_ICONS[meta.icon] : null);
	const selfServe = $derived(current === 'email' || current === 'ntfy' || current === 'webhook');
	const composedDetail = $derived(
		current === 'ntfy' ? `${ntfyServer.replace(/\/+$/, '')}/${ntfyTopic.trim()}` : detail.trim()
	);

	$effect(() => {
		const next = type;
		untrack(() => {
			if (next) {
				current = next;
				step = 'form';
				channelId = '';
				hint = '';
				detail = '';
				secret = '';
				ntfyServer = 'https://ntfy.sh';
				ntfyTopic = '';
			}
		});
	});
</script>

<Dialog.Root {open} onOpenChange={(value) => (value ? undefined : onclose())}>
	<Dialog.Content class="sm:max-w-[460px]">
		{#if meta}
			<div class="flex flex-col gap-4 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						{#if Icon}<Icon class="size-5" />{/if}
					</span>
					<div class="flex flex-1 flex-col justify-center">
						<Dialog.Title class="tracking-heading text-xl font-semibold">Connect {meta.label}</Dialog.Title>
					</div>
				</div>

				<div class="flex flex-col gap-3.5" role="status" aria-live="polite">
					{#if step === 'done'}
						<Alert.Root tone="success">
							<CircleCheckIcon />
							<Alert.Content>
								<Alert.Title>Channel connected</Alert.Title>
								<Dialog.Description class="text-muted-foreground text-sm">
									This channel is ready for your notification rules.
								</Dialog.Description>
							</Alert.Content>
						</Alert.Root>
					{:else if step === 'code'}
						<Dialog.Description class="text-muted-foreground text-[13px] leading-[1.6]">
							{hint || 'Enter the code we just sent to confirm you own this channel.'}
						</Dialog.Description>
						<form
							method="POST"
							action="?/verify"
							use:enhance={() =>
								async ({ result, update }) => {
									if (result.type === 'failure') {
										toast.error(String(result.data?.error ?? 'That code did not work.'));
										return;
									}
									if (result.type === 'success') {
										step = 'done';
										await update({ reset: false });
									}
								}}
						>
							<input type="hidden" name="id" value={channelId} />
							<Field.Field class="gap-1.5 space-y-0">
								<Field.FieldLabel for="verify-code" class="text-muted-foreground text-[13px] font-medium">
									Confirmation code
								</Field.FieldLabel>
								<Input id="verify-code" name="code" class="font-mono tracking-[0.15em]" placeholder="000000" inputmode="numeric" />
							</Field.Field>
							<Button type="submit" class="mt-3 self-start">Confirm channel</Button>
						</form>
					{:else if !selfServe}
						<Dialog.Description class="text-muted-foreground text-[13px] leading-[1.6]">
							{meta.label} is delivered through your workspace's chat integration. Link your {meta.label}
							account on the Chat connections page, then it appears here.
						</Dialog.Description>
						<Button href={ws('/chat')} variant="secondary" class="self-start">Open Chat connections</Button>
					{:else}
						<form
							bind:this={connectForm}
							method="POST"
							action="?/connect"
							use:enhance={() =>
								async ({ result, update }) => {
									if (result.type === 'failure') {
										toast.error(String(result.data?.error ?? 'Could not connect that channel.'));
										return;
									}
									if (result.type === 'success') {
										channelId = String(result.data?.channelId ?? '');
										hint = String(result.data?.detail ?? '');
										await update({ reset: false });
										step = current === 'webhook' ? 'done' : 'code';
									}
								}}
						>
							<input type="hidden" name="type" value={current} />
							<input type="hidden" name="detail" value={composedDetail} />
							{#if current === 'ntfy'}
								<Dialog.Description class="text-muted-foreground mb-3.5 text-[13px] leading-[1.6]">
									Get a push on any ntfy topic you choose.
								</Dialog.Description>
								<Field.Field class="gap-1.5 space-y-0">
									<Field.FieldLabel for="ntfy-server" class="text-muted-foreground text-[13px] font-medium">
										Server URL
									</Field.FieldLabel>
									<Input id="ntfy-server" bind:value={ntfyServer} />
								</Field.Field>
								<Field.Field class="mt-3 gap-1.5 space-y-0">
									<Field.FieldLabel for="ntfy-topic" class="text-muted-foreground text-[13px] font-medium">
										Topic
									</Field.FieldLabel>
									<Input id="ntfy-topic" class="font-mono" placeholder="my-pages-x7k2" bind:value={ntfyTopic} />
									<Field.FieldDescription class="text-subtle-foreground text-xs">
										Pick something unguessable. Anyone who knows the topic can read it.
									</Field.FieldDescription>
								</Field.Field>
								<Input name="secret" type="hidden" bind:value={secret} />
							{:else if current === 'email'}
								<Dialog.Description class="text-muted-foreground mb-3.5 text-[13px] leading-[1.6]">
									Verify the address Opsybot should email.
								</Dialog.Description>
								<Field.Field class="gap-1.5 space-y-0">
									<Field.FieldLabel for="email-addr" class="text-muted-foreground text-[13px] font-medium">
										Email address
									</Field.FieldLabel>
									<Input id="email-addr" type="email" placeholder="you@company.com" bind:value={detail} />
								</Field.Field>
							{:else}
								<Dialog.Description class="text-muted-foreground mb-3.5 text-[13px] leading-[1.6]">
									Opsybot will POST each notification to your endpoint.
								</Dialog.Description>
								<Field.Field class="gap-1.5 space-y-0">
									<Field.FieldLabel for="hook-url" class="text-muted-foreground text-[13px] font-medium">
										Endpoint URL
									</Field.FieldLabel>
									<Input id="hook-url" class="font-mono" placeholder="https://hooks.example.com/page" bind:value={detail} />
								</Field.Field>
								<Field.Field class="mt-3 gap-1.5 space-y-0">
									<Field.FieldLabel for="hook-secret" class="text-muted-foreground text-[13px] font-medium">
										Signing secret (optional)
									</Field.FieldLabel>
									<Input id="hook-secret" name="secret" class="font-mono" placeholder="Signs the payload" bind:value={secret} />
								</Field.Field>
							{/if}
							<Button type="submit" class="mt-4 self-start">
								{current === 'webhook' ? 'Connect and send challenge' : 'Connect and verify'}
							</Button>
						</form>
					{/if}
				</div>
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button variant="ghost" onclick={onclose}>{step === 'done' ? 'Done' : 'Close'}</Button>
			</div>
		{/if}
	</Dialog.Content>
</Dialog.Root>
