<script lang="ts">
	import { untrack } from 'svelte';
	import ArrowUpRightIcon from '@lucide/svelte/icons/arrow-up-right';
	import CircleAlertIcon from '@lucide/svelte/icons/circle-alert';
	import PlugIcon from '@lucide/svelte/icons/plug';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { CHAT_ICONS } from '$lib/components/chat/icons';
	import ScopeList from '$lib/components/chat/scope-list.svelte';
	import type { Platform } from '$lib/chat';

	let { platform, onclose }: { platform: Platform | null; onclose: () => void } = $props();

	let current = $state<Platform | null>(null);
	let externalId = $state('');
	let error = $state('');
	let pending = $state(false);
	const open = $derived(!!platform);

	$effect(() => {
		const next = platform;
		untrack(() => {
			if (next) {
				current = next;
				externalId = '';
				error = '';
				pending = false;
			}
		});
	});

	const Icon = $derived(current ? CHAT_ICONS[current.icon] : null);
	const oauth = $derived(current?.authKind === 'oauth');
	const field = $derived(!oauth ? (current?.externalIdField ?? null) : null);
</script>

<Dialog.Root {open} onOpenChange={(value) => (value ? undefined : onclose())}>
	<Dialog.Content class="sm:max-w-[500px]">
		{#if current}
			<form
				method="POST"
				action={oauth ? '?/oauthStart' : '?/connect'}
				use:enhance={() => {
					pending = true;
					error = '';
					return async ({ result, update }) => {
						pending = false;
						await update({ reset: false });
						if (result.type === 'success') {
							if (oauth && result.data?.oauthUrl) {
								window.location.href = String(result.data.oauthUrl);
								return;
							}
							const label = current?.label;
							onclose();
							if (label) toast.success(`${label} is connected.`);
							return;
						}
						if (result.type === 'failure') {
							error = String(result.data?.error ?? 'Could not connect that provider.');
						}
					};
				}}
			>
				<input type="hidden" name="platform" value={current.id} />
				<div class="flex flex-col gap-4 p-6">
					<div class="flex items-start gap-3">
						<span
							class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
						>
							{#if Icon}<Icon class="size-5" />{/if}
						</span>
						<div class="flex flex-1 flex-col justify-center">
							<Dialog.Title class="tracking-heading text-xl font-semibold">
								Connect {current.label}
							</Dialog.Title>
						</div>
					</div>

					<Dialog.Description class="text-muted-foreground text-[13px] leading-[1.6]">
						{#if oauth}
							Opsybot will send you to {current.label} to approve the install for this workspace, then
							bring you back here. It asks for exactly these permissions, nothing broader:
						{:else}
							Opsybot uses the {current.label} bot your admin configured — no token to paste here. Once
							connected, each person links their own account. It can:
						{/if}
					</Dialog.Description>
					<ScopeList scopes={current.scopes} />

					{#if field}
						<Field.Field class="gap-1.5 space-y-0">
							<Field.FieldLabel for="external-id" class="text-muted-foreground text-[13px] font-medium">
								{field.label}
							</Field.FieldLabel>
							<Input
								id="external-id"
								name="externalId"
								class="font-mono text-[12.5px]"
								placeholder={field.placeholder}
								bind:value={externalId}
								autocomplete="off"
							/>
							<Field.FieldDescription class="text-subtle-foreground text-xs">
								{field.hint}
							</Field.FieldDescription>
						</Field.Field>
					{/if}

					{#if error}
						<Alert.Root tone="critical">
							<CircleAlertIcon />
							<Alert.Content>
								<Alert.Title>Could not connect</Alert.Title>
								<Alert.Description class="text-sm">{error}</Alert.Description>
							</Alert.Content>
						</Alert.Root>
					{/if}

					<p class="text-subtle-foreground text-[12px] leading-[1.55]">
						You can disconnect any time. Incident history already captured stays in Opsybot.
					</p>
				</div>

				<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
					<Button type="button" variant="ghost" onclick={onclose}>Cancel</Button>
					<Button type="submit" disabled={pending}>
						{#if oauth}
							<ArrowUpRightIcon data-icon="inline-start" />
							{pending ? 'Redirecting…' : `Continue to ${current.label}`}
						{:else}
							<PlugIcon data-icon="inline-start" />
							{pending ? 'Connecting…' : `Connect ${current.label}`}
						{/if}
					</Button>
				</div>
			</form>
		{/if}
	</Dialog.Content>
</Dialog.Root>
