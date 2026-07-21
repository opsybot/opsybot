<script lang="ts">
	import { tick, untrack } from 'svelte';
	import ArrowUpRightIcon from '@lucide/svelte/icons/arrow-up-right';
	import CheckIcon from '@lucide/svelte/icons/check';
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import SendIcon from '@lucide/svelte/icons/send';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import { CHAT_ICONS } from '$lib/components/chat/icons';
	import ScopeList from '$lib/components/chat/scope-list.svelte';
	import type { InstallStep, Platform } from '$lib/chat';

	let { platform, onclose }: { platform: Platform | null; onclose: () => void } = $props();

	let current = $state<Platform | null>(null);
	let step = $state<InstallStep>('consent');
	const open = $derived(!!platform);

	let waitTimer: ReturnType<typeof setTimeout> | undefined;
	let testTimer: ReturnType<typeof setTimeout> | undefined;
	let connectForm: HTMLFormElement;
	let doneButton = $state<HTMLElement | null>(null);

	function clearTimers() {
		clearTimeout(waitTimer);
		clearTimeout(testTimer);
	}

	$effect(() => {
		const next = platform;
		untrack(() => {
			clearTimers();
			if (next) {
				current = next;
				step = 'consent';
			}
		});
	});

	const Icon = $derived(current ? CHAT_ICONS[current.icon] : null);
	const title = $derived(
		!current
			? ''
			: step === 'consent'
				? `Connect ${current.label}`
				: step === 'waiting'
					? `Waiting for ${current.label}`
					: `${current.label} connected`
	);

	function authorize() {
		step = 'waiting';
		clearTimeout(waitTimer);
		waitTimer = setTimeout(() => (step = 'done'), 1800);
	}

	async function runTest() {
		step = 'tested';
		clearTimeout(testTimer);
		testTimer = setTimeout(
			() => toast.success('/opsy test posted in #opsybot-sandbox: round trip 0.6 s.'),
			900
		);
		await tick();
		doneButton?.focus();
	}
</script>

<Dialog.Root {open} onOpenChange={(value) => (value ? undefined : onclose())}>
	<Dialog.Content class="sm:max-w-[500px]">
		{#if current}
			<div class="flex flex-col gap-4 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						{#if Icon}<Icon class="size-5" />{/if}
					</span>
					<div class="flex flex-1 flex-col justify-center">
						<Dialog.Title class="tracking-heading text-xl font-semibold">{title}</Dialog.Title>
					</div>
				</div>

				<div class="flex flex-col gap-3.5" role="status" aria-live="polite">
					{#if step === 'consent'}
						<Dialog.Description class="text-muted-foreground text-[13px] leading-[1.6]">
							Opsybot asks {current.label} for exactly these permissions, nothing broader:
						</Dialog.Description>
						<ScopeList scopes={current.scopes} />
						<p class="text-subtle-foreground text-[12px] leading-[1.55]">
							You can disconnect any time. Incident history already captured stays in Opsybot.
						</p>
					{:else if step === 'waiting'}
						<div class="text-muted-foreground flex items-center gap-2.5">
							<span
								class="border-border border-t-primary size-4 shrink-0 animate-spin rounded-full border-2 [animation-duration:0.8s] motion-reduce:animate-none"
								aria-hidden="true"
							></span>
							<Dialog.Description class="text-muted-foreground text-[13px]">
								Complete the authorization in the {current.label} window.
							</Dialog.Description>
						</div>
					{:else}
						<Alert.Root tone="success">
							<CircleCheckIcon />
							<Alert.Content>
								<Alert.Title>Workspace linked</Alert.Title>
								<Dialog.Description class="text-muted-foreground text-sm">
									Connected to <strong class="text-foreground font-semibold">Acme Corp</strong>. The bot
									joined #opsybot-sandbox for testing.
								</Dialog.Description>
							</Alert.Content>
						</Alert.Root>
						<div>
							<div class="text-subtle-foreground mb-2.5 text-[11px] tracking-[0.08em] uppercase">
								One last check. Run a test command
							</div>
							<div class="flex items-center gap-2">
								<code
									class="bg-inset text-foreground flex-1 rounded-md border px-[11px] py-[9px] font-mono text-[12.5px]"
									>/opsy test</code
								>
								<Button size="sm" variant="secondary" disabled={step === 'tested'} onclick={runTest}>
									<SendIcon data-icon="inline-start" />
									{step === 'tested' ? 'Sent' : 'Run it for me'}
								</Button>
							</div>
							{#if step === 'tested'}
								<div class="mt-2 flex items-center gap-[7px]">
									<CheckIcon class="text-primary size-[13px]" />
									<span class="text-muted-foreground text-[12.5px]"
										>Bot replied in #opsybot-sandbox. You're set.</span
									>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				{#if step === 'consent'}
					<Button variant="ghost" onclick={onclose}>Cancel</Button>
					<Button onclick={authorize}>
						<ArrowUpRightIcon data-icon="inline-start" />
						Continue to {current.label}
					</Button>
				{:else if step === 'waiting'}
					<Button variant="ghost" onclick={onclose}>Cancel</Button>
				{:else}
					<Button bind:ref={doneButton} onclick={() => connectForm.requestSubmit()}>Done</Button>
				{/if}
			</div>
		{/if}
	</Dialog.Content>
</Dialog.Root>

<form
	bind:this={connectForm}
	method="POST"
	action="?/connect"
	class="hidden"
	use:enhance={() =>
		async ({ result, update }) => {
			await update({ reset: false });
			if (result.type !== 'success') return;
			const label = current?.label;
			onclose();
			if (label) toast.success(`${label} is connected.`);
		}}
>
	<input type="hidden" name="platform" value={current?.id ?? ''} />
</form>
