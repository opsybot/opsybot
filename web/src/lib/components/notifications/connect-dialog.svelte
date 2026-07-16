<script lang="ts">
	import { onDestroy, untrack } from 'svelte';
	import ArrowUpRightIcon from '@lucide/svelte/icons/arrow-up-right';
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import { enhance } from '$app/forms';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { CHANNEL_ICONS } from '$lib/components/notifications/icons';
	import { channelMeta, type ChannelType } from '$lib/notifications';

	let { type, onclose }: { type: ChannelType | null; onclose: () => void } = $props();

	// current keeps the last non-null type so labels survive the close animation
	let current = $state<ChannelType | null>(null);
	let step = $state<'form' | 'waiting' | 'done'>('form');
	const open = $derived(!!type);

	let timer: ReturnType<typeof setTimeout> | undefined;
	let connectForm: HTMLFormElement;

	$effect(() => {
		const next = type;
		untrack(() => {
			clearTimeout(timer);
			if (next) {
				current = next;
				step = 'form';
			}
		});
	});

	const meta = $derived(current ? channelMeta(current) : null);
	const Icon = $derived(meta ? CHANNEL_ICONS[meta.icon] : null);
	const waiting = $derived(
		meta?.connect === 'oauth'
			? 'Waiting for authorization…'
			: meta?.connect === 'telegram'
				? 'Waiting for the code…'
				: meta?.connect === 'ntfy'
					? 'Sending a test push…'
					: meta?.connect === 'email'
						? 'Waiting for you to click the link…'
						: 'POSTing a test event…'
	);

	function submit() {
		step = 'waiting';
		clearTimeout(timer);
		timer = setTimeout(() => {
			step = 'done';
			connectForm.requestSubmit();
		}, 1800);
	}

	onDestroy(() => clearTimeout(timer));
</script>

{#snippet actionRow(label: string, icon = false)}
	{#if step === 'waiting'}
		<div class="text-muted-foreground flex items-center gap-2.5">
			<span
				class="border-border border-t-primary size-4 shrink-0 animate-spin rounded-full border-2 [animation-duration:0.8s] motion-reduce:animate-none"
				aria-hidden="true"
			></span>
			<span class="text-[13px]">{waiting}</span>
		</div>
	{:else}
		<Button class="self-start" onclick={submit}>
			{#if icon}<ArrowUpRightIcon data-icon="inline-start" />{/if}
			{label}
		</Button>
	{/if}
{/snippet}

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
								<Alert.Title>Channel verified</Alert.Title>
								<Dialog.Description class="text-muted-foreground text-sm">
									A test message was delivered. This channel is ready for your notification rules.
								</Dialog.Description>
							</Alert.Content>
						</Alert.Root>
					{:else if meta.connect === 'oauth'}
						<Dialog.Description class="text-muted-foreground text-[13px] leading-[1.6]">
							You'll be sent to {meta.label} to authorize Opsybot. It only gets permission to DM you — nothing
							else.
						</Dialog.Description>
						{@render actionRow(`Continue to ${meta.label}`, true)}
					{:else if meta.connect === 'telegram'}
						<Dialog.Description class="text-muted-foreground text-[13px] leading-[1.6]">
							Two quick steps to link Telegram:
						</Dialog.Description>
						<ol class="text-muted-foreground m-0 list-decimal pl-[18px] text-[13px] leading-[1.8]">
							<li>
								Open
								<a href="https://t.me/opsybot_bot" target="_blank" rel="noopener noreferrer">t.me/opsybot_bot</a>
								and press <strong class="text-foreground font-semibold">Start</strong>.
							</li>
							<li>Send the bot this code:</li>
						</ol>
						<code
							class="bg-inset text-foreground self-start rounded-md border px-3.5 py-[9px] font-mono text-[16px] tracking-[0.15em]"
							>824 913</code
						>
						{@render actionRow('I sent the code')}
					{:else if meta.connect === 'ntfy'}
						<Dialog.Description class="text-muted-foreground text-[13px] leading-[1.6]">
							Get a push on any ntfy topic you choose.
						</Dialog.Description>
						<Field.Field class="gap-1.5 space-y-0">
							<Field.FieldLabel for="ntfy-server" class="text-muted-foreground text-[13px] font-medium">
								Server URL
							</Field.FieldLabel>
							<Input id="ntfy-server" value="https://ntfy.sh" />
							<Field.FieldDescription class="text-subtle-foreground text-xs">
								Self-hosted servers work too.
							</Field.FieldDescription>
						</Field.Field>
						<Field.Field class="gap-1.5 space-y-0">
							<Field.FieldLabel for="ntfy-topic" class="text-muted-foreground text-[13px] font-medium">
								Topic
							</Field.FieldLabel>
							<Input id="ntfy-topic" class="font-mono" placeholder="maya-pages-x7k2" />
							<Field.FieldDescription class="text-subtle-foreground text-xs">
								Pick something unguessable — anyone who knows the topic can read it.
							</Field.FieldDescription>
						</Field.Field>
						{@render actionRow('Connect and send test')}
					{:else if meta.connect === 'email'}
						<Dialog.Description class="text-muted-foreground text-[13px] leading-[1.6]">
							Verify the address Opsybot should email.
						</Dialog.Description>
						<Field.Field class="gap-1.5 space-y-0">
							<Field.FieldLabel for="email-addr" class="text-muted-foreground text-[13px] font-medium">
								Email address
							</Field.FieldLabel>
							<Input id="email-addr" type="email" value="maya@acme.dev" />
						</Field.Field>
						{@render actionRow('Send verification link')}
					{:else}
						<Dialog.Description class="text-muted-foreground text-[13px] leading-[1.6]">
							Opsybot will POST each notification to your endpoint.
						</Dialog.Description>
						<Field.Field class="gap-1.5 space-y-0">
							<Field.FieldLabel for="hook-url" class="text-muted-foreground text-[13px] font-medium">
								Endpoint URL
							</Field.FieldLabel>
							<Input id="hook-url" class="font-mono" placeholder="https://hooks.example.dev/page" />
						</Field.Field>
						<Field.Field class="gap-1.5 space-y-0">
							<Field.FieldLabel for="hook-secret" class="text-muted-foreground text-[13px] font-medium">
								Secret (optional)
							</Field.FieldLabel>
							<Input id="hook-secret" class="font-mono" placeholder="Used to sign the payload" />
						</Field.Field>
						{@render actionRow('Connect and send test')}
					{/if}
				</div>
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button variant="ghost" onclick={onclose}>Close</Button>
			</div>
		{/if}
	</Dialog.Content>
</Dialog.Root>

<form
	bind:this={connectForm}
	method="POST"
	action="?/connect"
	class="hidden"
	use:enhance={() => async ({ result, update }) => {
		if (result.type === 'success') await update({ reset: false });
	}}
>
	<input type="hidden" name="type" value={current ?? ''} />
</form>
