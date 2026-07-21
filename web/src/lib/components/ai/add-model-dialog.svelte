<script lang="ts">
	import { onDestroy, tick, untrack } from 'svelte';
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import SendIcon from '@lucide/svelte/icons/send';
	import SparklesIcon from '@lucide/svelte/icons/sparkles';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { CONTEXT_OPTIONS, TIMEOUT_OPTIONS } from '$lib/ai';

	let { open = $bindable(false) }: { open?: boolean } = $props();

	let name = $state('');
	let endpoint = $state('');
	let apiKey = $state('');
	let timeout = $state(TIMEOUT_OPTIONS[1]);
	let maxContext = $state(CONTEXT_OPTIONS[1]);
	let test = $state<'idle' | 'running' | 'ok'>('idle');

	let timer: ReturnType<typeof setTimeout>;
	let addForm: HTMLFormElement;
	let addButton = $state<HTMLElement | null>(null);

	$effect(() => {
		if (!open) return;
		untrack(() => {
			clearTimeout(timer);
			name = '';
			endpoint = '';
			apiKey = '';
			timeout = TIMEOUT_OPTIONS[1];
			maxContext = CONTEXT_OPTIONS[1];
			test = 'idle';
		});
	});
	onDestroy(() => clearTimeout(timer));

	$effect(() => {
		void [name, endpoint, apiKey, timeout, maxContext];
		untrack(() => {
			if (test !== 'idle') test = 'idle';
		});
	});

	function runTest() {
		test = 'running';
		clearTimeout(timer);
		timer = setTimeout(async () => {
			test = 'ok';
			await tick();
			addButton?.focus();
		}, 1800);
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[520px]">
		<div class="flex flex-col gap-4 p-6">
			<div class="flex items-start gap-3">
				<span
					class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
				>
					<SparklesIcon class="size-5" />
				</span>
				<div class="flex flex-1 flex-col gap-1">
					<Dialog.Title class="tracking-heading text-xl font-semibold">Add model connection</Dialog.Title>
					<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
						Self-hosted (Ollama, vLLM) or a hosted API. Incident data goes only to endpoints you add here.
					</Dialog.Description>
				</div>
			</div>

			<div class="flex flex-col gap-3.5">
				<div class="flex gap-2.5">
					<Field.Field class="flex-1 gap-1.5 space-y-0">
						<Field.FieldLabel for="model-name" class="text-muted-foreground text-[13px] font-medium">Name</Field.FieldLabel>
						<Input id="model-name" bind:value={name} placeholder="ollama-prod" />
					</Field.Field>
					<Field.Field class="w-[110px] gap-1.5 space-y-0">
						<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">Timeout</Field.FieldLabel>
						<Select.Root type="single" value={timeout} onValueChange={(value) => (timeout = value)}>
							<Select.Trigger size="sm" aria-label="Timeout">{timeout}</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each TIMEOUT_OPTIONS as option (option)}
										<Select.Item value={option} label={option}>{option}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</Field.Field>
				</div>

				<Field.Field class="gap-1.5 space-y-0">
					<Field.FieldLabel for="model-endpoint" class="text-muted-foreground text-[13px] font-medium">Endpoint URL</Field.FieldLabel>
					<Input id="model-endpoint" class="font-mono text-[12.5px]" bind:value={endpoint} placeholder="http://10.0.4.12:11434" />
				</Field.Field>

				<div class="flex gap-2.5">
					<Field.Field class="flex-1 gap-1.5 space-y-0">
						<Field.FieldLabel for="model-key" class="text-muted-foreground text-[13px] font-medium">API key</Field.FieldLabel>
						<Input id="model-key" type="password" bind:value={apiKey} placeholder="empty for local endpoints" />
					</Field.Field>
					<Field.Field class="w-[150px] gap-1.5 space-y-0">
						<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">Max context</Field.FieldLabel>
						<Select.Root type="single" value={maxContext} onValueChange={(value) => (maxContext = value)}>
							<Select.Trigger size="sm" aria-label="Max context">{maxContext}</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each CONTEXT_OPTIONS as option (option)}
										<Select.Item value={option} label={option}>{option}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</Field.Field>
				</div>

				{#if test === 'ok'}
					<Alert.Root tone="success">
						<CircleCheckIcon />
						<Alert.Content>
							<Alert.Title>Test request succeeded</Alert.Title>
							<Alert.Description>
								Round trip 1.9 s · 128k context confirmed · streaming works. No incident data was sent: the
								test uses a canned prompt.
							</Alert.Description>
						</Alert.Content>
					</Alert.Root>
				{:else}
					<div class="flex items-center gap-2.5" role="status" aria-live="polite">
						<Button
							size="sm"
							variant="secondary"
							disabled={!name.trim() || !endpoint.trim() || test === 'running'}
							onclick={runTest}
						>
							<SendIcon data-icon="inline-start" />
							{test === 'running' ? 'Testing…' : 'Run test request'}
						</Button>
						{#if test === 'running'}
							<span
								class="border-border border-t-primary size-4 shrink-0 animate-spin rounded-full border-2 [animation-duration:0.8s] motion-reduce:animate-none"
								aria-hidden="true"
							></span>
						{/if}
						<span class="text-subtle-foreground text-[11.5px]">Required before the model can be added.</span>
					</div>
				{/if}
			</div>
		</div>

		<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
			<Button variant="ghost" onclick={() => (open = false)}>Cancel</Button>
			<Button bind:ref={addButton} disabled={test !== 'ok'} onclick={() => addForm.requestSubmit()}>
				Add model
			</Button>
		</div>
	</Dialog.Content>
</Dialog.Root>

<form
	bind:this={addForm}
	method="POST"
	action="?/addModel"
	class="hidden"
	use:enhance={() => async ({ result, update }) => {
		if (result.type === 'failure') {
			toast.error(String(result.data?.error ?? 'Could not add the model.'));
			return;
		}
		if (result.type !== 'success') return;
		const added = name.trim();
		await update({ reset: false });
		open = false;
		toast.success(`${added} connected.`);
	}}
>
	<input type="hidden" name="name" value={name} />
	<input type="hidden" name="endpoint" value={endpoint} />
	<input type="hidden" name="timeout" value={timeout} />
	<input type="hidden" name="maxContext" value={maxContext} />
</form>
