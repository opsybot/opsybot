<script lang="ts">
	import { untrack } from 'svelte';
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';
	import { superForm, type Infer, type SuperValidated } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import TextField from '$lib/components/text-field.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import { recoveryCodeSchema } from '$lib/schemas/auth';

	let { data }: { data: SuperValidated<Infer<typeof recoveryCodeSchema>> } = $props();

	const form = superForm(untrack(() => data), { validators: zod4Client(recoveryCodeSchema) });
	const { message } = form;
</script>

{#if $message === 'wrong'}
	<Alert.Root tone="critical" class="mb-4">
		<OctagonAlertIcon />
		<Alert.Content>
			<Alert.Description>
				That recovery code didn't match. Each code works once. Try another.
			</Alert.Description>
		</Alert.Content>
	</Alert.Root>
{/if}

<form method="POST" action="?/recovery" use:form.enhance class="flex flex-col gap-4">
	<TextField
		{form}
		name="code"
		label="Recovery code"
		placeholder="xxxx-xxxx"
		autocomplete="off"
		class="font-mono"
	/>
	<Button type="submit" class="w-full">Verify</Button>
</form>
