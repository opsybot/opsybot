<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import CheckIcon from '@lucide/svelte/icons/check';
	import SendIcon from '@lucide/svelte/icons/send';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import AddStepper from '$lib/components/alertsources/add-stepper.svelte';
	import CopyField from '$lib/components/alertsources/copy-field.svelte';
	import MappingTable from '$lib/components/alertsources/mapping-table.svelte';
	import { ICON } from '$lib/components/alertsources/icons';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { MAPPINGS, endpointUrl, slugify, type Format, type Mapping } from '$lib/alertsources';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let step = $state(0);
	let format = $state<Format | null>(null);
	let name = $state('');
	let rows = $state<Mapping[]>([]);
	let testState = $state<'idle' | 'waiting' | 'ok'>('idle');

	const slug = $derived(slugify(name) || format?.id || 'source');
	const url = $derived(endpointUrl(slug));

	function pickFormat(chosen: Format) {
		format = chosen;
		rows = MAPPINGS[chosen.id].map((mapping) => ({ ...mapping }));
	}

	function runTest() {
		testState = 'waiting';
		setTimeout(() => (testState = 'ok'), 1600);
	}

	let createForm: HTMLFormElement;
</script>

<div class="flex max-w-[720px] flex-col gap-4">
	<a
		href={ws('/alert-sources')}
		class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[12.5px] transition-colors"
	>
		<ArrowLeftIcon class="size-3.5" />
		Alert sources
	</a>
	<h2 class="text-[18px] font-semibold tracking-[-0.01em]">Add integration</h2>
	<AddStepper {step} />

	{#if step === 0}
		<div class="flex flex-col gap-4">
			<div class="grid gap-2.5 [grid-template-columns:repeat(auto-fill,minmax(200px,1fr))]">
				{#each data.formats as option (option.id)}
					{@const Icon = ICON[option.icon]}
					<button
						type="button"
						onclick={() => pickFormat(option)}
						class="bg-card flex flex-col items-start gap-2 rounded-xl border p-3.5 text-left transition-[border-color,transform] hover:-translate-y-px hover:border-border-strong {format?.id ===
						option.id
							? 'border-brand-edge bg-brand-wash'
							: ''}"
					>
						<span
							class="bg-inset text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-sm border"
						>
							<Icon class="size-4" />
						</span>
						<span class="text-[13.5px] font-semibold">{option.label}</span>
						<span class="text-subtle-foreground text-[12px] leading-[1.45]">{option.desc}</span>
					</button>
				{/each}
			</div>
			<Field.Field class="max-w-[320px] gap-1.5 space-y-0">
				<Field.FieldLabel for="src-name" class="text-muted-foreground text-[13px] font-medium">
					Name
				</Field.FieldLabel>
				<Input
					id="src-name"
					bind:value={name}
					placeholder={format ? `${format.id}-prod` : 'prometheus-prod'}
					class="font-mono"
				/>
				<Field.FieldDescription class="text-subtle-foreground text-xs">
					Shows on alerts from this source.
				</Field.FieldDescription>
			</Field.Field>
			<Button class="self-start" disabled={!format || !name.trim()} onclick={() => (step = 1)}>
				Continue
			</Button>
		</div>
	{:else if step === 1}
		<div class="bg-card flex flex-col gap-4 rounded-xl border p-[18px]">
			<p class="text-subtle-foreground m-0 text-[12.5px] leading-[1.55]">
				The endpoint URL and signing secret are generated when you create the source, and shown on
				its page. Requests are verified with an HMAC of the body using that secret ({format?.id ===
				'heartbeat'
					? 'optional for heartbeats'
					: 'header X-Opsy-Signature'}). Rotate it any time.
			</p>
			<div class="flex gap-2.5">
				<Button variant="ghost" onclick={() => (step = 0)}>Back</Button>
				<Button onclick={() => (step = 2)}>Continue</Button>
			</div>
		</div>
	{:else if step === 2}
		<div class="flex flex-col gap-4">
			<MappingTable
				bind:rows
				editable={format?.id === 'generic'}
				note={format?.id === 'generic' ? undefined : 'Pre-filled for this format. Mapping stays editable after setup.'}
			/>
			<div class="flex gap-2.5">
				<Button variant="ghost" onclick={() => (step = 1)}>Back</Button>
				<Button onclick={() => (step = 3)}>Continue</Button>
			</div>
		</div>
	{:else}
		<div class="bg-card flex flex-col gap-4 rounded-xl border p-[18px]">
			<p class="text-muted-foreground m-0 text-[13.5px] leading-[1.6]">
				Fire a test alert from {name} at the endpoint, or let Opsybot send one for you, and watch it
				arrive.
			</p>
			{#if testState === 'idle'}
				<Button class="self-start" onclick={runTest}>
					<SendIcon data-icon="inline-start" />
					Send a test event
				</Button>
			{:else if testState === 'waiting'}
				<div class="text-muted-foreground flex items-center gap-2.5">
					<span
						class="border-border border-t-primary size-4 shrink-0 animate-spin rounded-full border-2 motion-reduce:animate-none"
						aria-hidden="true"
					></span>
					<span class="text-[13px]">Waiting for the event…</span>
				</div>
			{:else}
				<Alert.Root tone="success">
					<CheckIcon />
					<Alert.Content>
						<Alert.Title>Test event received</Alert.Title>
						<Alert.Description>
							2026-07-11 09:58:12 UTC · signature valid · parsed OK: title, severity, and service all
							mapped. It would route through your rules; no one was paged.
						</Alert.Description>
					</Alert.Content>
				</Alert.Root>
				<div class="flex gap-2.5">
					<Button onclick={() => createForm.requestSubmit()}>
						<CheckIcon data-icon="inline-start" />
						Finish setup
					</Button>
					<Button variant="ghost" onclick={() => (testState = 'idle')}>Send another</Button>
				</div>
			{/if}
			{#if testState !== 'ok'}
				<div class="flex gap-2.5">
					<Button variant="ghost" onclick={() => (step = 2)}>Back</Button>
					<Button variant="ghost" onclick={() => createForm.requestSubmit()}>Skip verification</Button>
				</div>
			{/if}
		</div>
	{/if}

	<form
		method="POST"
		action="?/create"
		bind:this={createForm}
		class="hidden"
		use:enhance={() =>
			async ({ result }) => {
				if (result.type === 'failure') {
					toast.error(String(result.data?.error ?? 'Could not connect the source.'));
					return;
				}
				if (result.type !== 'success') return;
				toast.success(`${name} is connected. Alerts route through your routing rules.`);
				await goto(ws('/alert-sources'));
			}}
	>
		<input type="hidden" name="name" value={name} />
		<input type="hidden" name="format" value={format?.id ?? ''} />
		<input type="hidden" name="mapping" value={JSON.stringify(rows)} />
	</form>
</div>
