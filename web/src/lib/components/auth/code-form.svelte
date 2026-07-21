<script lang="ts">
	import { untrack } from 'svelte';
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';
	import { superForm, type Infer, type SuperValidated } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Form from '$lib/components/ui/form';
	import * as InputOTP from '$lib/components/ui/input-otp';
	import { totpSchema } from '$lib/schemas/auth';

	let {
		data,
		action = '?/code',
		submitLabel = 'Verify',
		wrongMessage = "That code didn't match. Codes rotate every 30 s. Wait for a fresh one and try again."
	}: {
		data: SuperValidated<Infer<typeof totpSchema>>;
		action?: string;
		submitLabel?: string;
		wrongMessage?: string;
	} = $props();

	const form = superForm(untrack(() => data), { validators: zod4Client(totpSchema) });
	const { form: formData, message } = form;
</script>

{#if $message === 'wrong'}
	<Alert.Root tone="critical" class="mb-4">
		<OctagonAlertIcon />
		<Alert.Content>
			<Alert.Description>{wrongMessage}</Alert.Description>
		</Alert.Content>
	</Alert.Root>
{/if}

<form method="POST" {action} use:form.enhance class="flex flex-col gap-4">
	<Form.Field {form} name="code" class="flex flex-col gap-1.5 space-y-0">
		<Form.Control>
			{#snippet children({ props })}
				<input type="hidden" {...props} value={$formData.code} />
				<InputOTP.Root maxlength={6} bind:value={$formData.code}>
					{#snippet children({ cells })}
						<InputOTP.Group>
							{#each cells.slice(0, 3) as cell, index (index)}
								<InputOTP.Slot {cell} class="mr-2 last:mr-0" />
							{/each}
						</InputOTP.Group>
						<InputOTP.Separator />
						<InputOTP.Group>
							{#each cells.slice(3, 6) as cell, index (index)}
								<InputOTP.Slot {cell} class="mr-2 last:mr-0" />
							{/each}
						</InputOTP.Group>
					{/snippet}
				</InputOTP.Root>
			{/snippet}
		</Form.Control>
		<Form.FieldErrors class="text-critical-ink text-xs font-normal" />
	</Form.Field>

	<Button type="submit" class="w-full">{submitLabel}</Button>
</form>
