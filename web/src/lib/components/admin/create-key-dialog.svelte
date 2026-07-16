<script lang="ts">
	import { tick, untrack } from 'svelte';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import CopyField from '$lib/components/alertsources/copy-field.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { SCOPES, type KeyKind } from '$lib/admin';

	let { open = $bindable(false), kind }: { open?: boolean; kind: KeyKind } = $props();

	let name = $state('');
	let picked = $state(new Set<string>(['incidents:read']));
	let step = $state<'form' | 'revealed'>('form');
	let secret = $state('');
	let submitting = $state(false);

	let createForm: HTMLFormElement;
	let savedButton = $state<HTMLElement | null>(null);

	$effect(() => {
		if (!open) return;
		untrack(() => {
			name = '';
			picked = new Set(['incidents:read']);
			step = 'form';
			secret = '';
			submitting = false;
		});
	});

	const scopesJson = $derived(JSON.stringify([...picked]));

	function toggle(scope: string, on: boolean | 'indeterminate') {
		const next = new Set(picked);
		if (on) next.add(scope);
		else next.delete(scope);
		picked = next;
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[500px]">
		<div class="flex flex-col gap-4 p-6">
			<div class="flex items-start gap-3">
				<span class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg">
					<KeyRoundIcon class="size-5" />
				</span>
				<div class="flex flex-1 flex-col gap-1">
					<Dialog.Title class="tracking-heading text-xl font-semibold">
						{step === 'revealed' ? 'Copy your key now' : `Create ${kind} key`}
					</Dialog.Title>
					{#if step === 'revealed'}
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							This is the only time it's shown. Store it in a secrets manager.
						</Dialog.Description>
					{/if}
				</div>
			</div>

			{#if step === 'revealed'}
				<div class="flex flex-col gap-3.5" role="status" aria-live="polite">
					<CopyField label="API key" value={secret} secret />
				</div>
			{:else}
				<div class="flex flex-col gap-3.5">
					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel for="key-name" class="text-muted-foreground text-[13px] font-medium">Name</Field.FieldLabel>
						<Input id="key-name" class="font-mono text-[12.5px]" bind:value={name} placeholder="terraform-provider" />
					</Field.Field>
					<div>
						<div class="text-subtle-foreground mb-2.5 text-[11px] tracking-[0.08em] uppercase">Scopes</div>
						<div class="grid grid-cols-2 gap-2.5">
							{#each SCOPES as scope (scope)}
								<label class="flex items-center gap-2.5 text-[13px]">
									<Checkbox checked={picked.has(scope)} onCheckedChange={(value) => toggle(scope, value)} />
									<span class="font-mono text-[12px]">{scope}</span>
								</label>
							{/each}
						</div>
					</div>
				</div>
			{/if}
		</div>

		<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
			{#if step === 'revealed'}
				<Button bind:ref={savedButton} onclick={() => (open = false)}>I saved it</Button>
			{:else}
				<Button variant="ghost" onclick={() => (open = false)}>Cancel</Button>
				<Button
					disabled={!name.trim() || picked.size === 0 || submitting}
					onclick={() => {
						submitting = true;
						createForm.requestSubmit();
					}}
				>
					Create key
				</Button>
			{/if}
		</div>
	</Dialog.Content>
</Dialog.Root>

<form
	bind:this={createForm}
	method="POST"
	action="?/create"
	class="hidden"
	use:enhance={() => async ({ result, update }) => {
		if (result.type === 'failure') {
			submitting = false;
			toast.error(String(result.data?.error ?? 'Could not create the key.'));
			return;
		}
		if (result.type !== 'success') {
			submitting = false;
			return;
		}
		secret = String(result.data?.secret ?? '');
		step = 'revealed';
		await update({ reset: false });
		await tick();
		savedButton?.focus();
	}}
>
	<input type="hidden" name="name" value={name} />
	<input type="hidden" name="scopes" value={scopesJson} />
	<input type="hidden" name="kind" value={kind} />
</form>
