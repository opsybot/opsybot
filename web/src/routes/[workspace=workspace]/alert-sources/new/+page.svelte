<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import CheckIcon from '@lucide/svelte/icons/check';
	import { onDestroy } from 'svelte';
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
	import { MAPPINGS, type Format, type Mapping, type Source } from '$lib/alertsources';
	import { ws } from '$lib/navigation';
	import type { DeliveryEvent } from '$lib/server/alertsources';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let step = $state(0);
	let format = $state<Format | null>(null);
	let name = $state('');
	let rows = $state<Mapping[]>([]);
	let source = $state<Source | null>(null);
	let events = $state<DeliveryEvent[]>([]);
	let polling = $state(false);

	let createForm: HTMLFormElement;
	let mappingForm: HTMLFormElement;
	let checkForm: HTMLFormElement;
	let timer: ReturnType<typeof setInterval> | null = null;

	function stopPolling() {
		if (timer) clearInterval(timer);
		timer = null;
		polling = false;
	}

	onDestroy(stopPolling);

	function startPolling() {
		if (timer) return;
		polling = true;
		timer = setInterval(() => checkForm.requestSubmit(), 3000);
	}

	function pickFormat(chosen: Format) {
		format = chosen;
		rows = MAPPINGS[chosen.id].map((mapping) => ({ ...mapping }));
	}

	function finish() {
		stopPolling();
		goto(ws(`/alert-sources/${source?.slug}`));
	}
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
			<p class="text-subtle-foreground m-0 text-[12.5px]">
				For a job that should check in on a schedule, create a
				<a href={ws('/alerts/heartbeats')} class="text-brand-foreground underline">heartbeat monitor</a>
				instead.
			</p>
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
			<Button
				class="self-start"
				disabled={!format || !name.trim()}
				onclick={() => createForm.requestSubmit()}
			>
				Continue
			</Button>
		</div>
	{:else if step === 1 && source}
		<div class="bg-card flex flex-col gap-4 rounded-xl border p-[18px]">
			<p class="text-subtle-foreground m-0 text-[12.5px] leading-[1.55]">
				Point {source.name} at this URL. Requests may be signed with an HMAC of the body in the
				<code class="font-mono">X-Opsy-Signature</code> header using the secret below. Both are shown
				again on the source page, and the secret can be rotated any time.
			</p>
			<CopyField label="Endpoint URL" value={source.ingestUrl} />
			<CopyField label="Signing secret" value={source.secret} secret />
			<div class="flex gap-2.5">
				<Button onclick={() => (step = 2)}>Continue</Button>
			</div>
		</div>
	{:else if step === 2}
		<div class="flex flex-col gap-4">
			<MappingTable
				bind:rows
				editable={format?.id === 'generic'}
				note={format?.id === 'generic'
					? undefined
					: 'Pre-filled for this format. Mapping stays editable after setup.'}
			/>
			<div class="flex gap-2.5">
				<Button variant="ghost" onclick={() => (step = 1)}>Back</Button>
				<Button
					onclick={() => (format?.id === 'generic' ? mappingForm.requestSubmit() : (step = 3))}
				>
					Continue
				</Button>
			</div>
		</div>
	{:else}
		<div class="bg-card flex flex-col gap-4 rounded-xl border p-[18px]">
			<p class="text-muted-foreground m-0 text-[13.5px] leading-[1.6]">
				Fire an alert from {name} at the endpoint and watch it arrive. It runs through your routing
				rules like any other alert.
			</p>

			{#if events.length}
				{@const latest = events[0]}
				<Alert.Root tone={latest.outcome === 'failed' ? 'critical' : 'success'}>
					<CheckIcon />
					<Alert.Content>
						<Alert.Title>
							{latest.outcome === 'failed' ? 'Event rejected' : 'Event received'}
						</Alert.Title>
						<Alert.Description>
							{latest.at} · {latest.outcome} · dedup key {latest.title}
						</Alert.Description>
					</Alert.Content>
				</Alert.Root>
			{:else if polling}
				<div class="text-muted-foreground flex items-center gap-2.5">
					<span
						class="border-border border-t-primary size-4 shrink-0 animate-spin rounded-full border-2 motion-reduce:animate-none"
						aria-hidden="true"
					></span>
					<span class="text-[13px]">Waiting for the first event…</span>
				</div>
			{/if}

			<div class="flex gap-2.5">
				{#if events.length}
					<Button onclick={finish}>
						<CheckIcon data-icon="inline-start" />
						Finish setup
					</Button>
				{:else if polling}
					<Button variant="ghost" onclick={stopPolling}>Stop waiting</Button>
				{:else}
					<Button onclick={startPolling}>Watch for an event</Button>
				{/if}
				<Button variant="ghost" onclick={() => (step = 2)}>Back</Button>
				<Button variant="ghost" onclick={finish}>Skip verification</Button>
			</div>
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
				source = result.data?.source as Source;
				step = 1;
			}}
	>
		<input type="hidden" name="name" value={name} />
		<input type="hidden" name="format" value={format?.id ?? ''} />
	</form>

	<form
		method="POST"
		action="?/mapping"
		bind:this={mappingForm}
		class="hidden"
		use:enhance={() =>
			async ({ result }) => {
				if (result.type === 'failure') {
					toast.error(String(result.data?.error ?? 'Could not save that mapping.'));
					return;
				}
				if (result.type === 'success') step = 3;
			}}
	>
		<input type="hidden" name="slug" value={source?.slug ?? ''} />
		<input type="hidden" name="mapping" value={JSON.stringify(rows)} />
	</form>

	<form
		method="POST"
		action="?/check"
		bind:this={checkForm}
		class="hidden"
		use:enhance={() =>
			async ({ result }) => {
				if (result.type !== 'success') return;
				const received = (result.data?.events ?? []) as DeliveryEvent[];
				if (!received.length) return;
				events = received;
				stopPolling();
			}}
	>
		<input type="hidden" name="slug" value={source?.slug ?? ''} />
	</form>
</div>
