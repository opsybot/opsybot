<script lang="ts">
	import { untrack } from 'svelte';
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import AuthShell from '$lib/components/auth/auth-shell.svelte';
	import SlugField from '$lib/components/auth/slug-field.svelte';
	import TextField from '$lib/components/text-field.svelte';
	import TimezoneSelect from '$lib/components/timezone-select.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import { setupSchema } from '$lib/schemas/auth';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const form = superForm(untrack(() => data.form), { validators: zod4Client(setupSchema) });
	const { form: formData, message } = form;
</script>

<AuthShell
	title="Set up this instance"
	subtitle="Opsybot is running. Create the first account and workspace to finish setup."
>
	<div class="flex flex-col gap-5">
		<Alert.Root tone="brand">
			<ShieldCheckIcon />
			<Alert.Content>
				<Alert.Title>This account becomes the instance admin</Alert.Title>
				<Alert.Description>
					The instance admin manages users, SSO, and instance settings. You can add more admins
					later.
				</Alert.Description>
			</Alert.Content>
		</Alert.Root>

		{#if $message}
			<Alert.Root tone="critical">
				<OctagonAlertIcon />
				<Alert.Content>
					<Alert.Description>{$message}</Alert.Description>
				</Alert.Content>
			</Alert.Root>
		{/if}

		<form method="POST" use:form.enhance class="flex flex-col gap-4">
			<TextField {form} name="name" label="Name" placeholder="Maya Chen" autocomplete="name" />
			<TextField
				{form}
				name="email"
				label="Email"
				type="email"
				placeholder="you@company.com"
				autocomplete="email"
			/>
			<TextField
				{form}
				name="password"
				label="Password"
				type="password"
				placeholder="12+ characters"
				hint="12 characters minimum."
				autocomplete="new-password"
			/>
			<TextField
				{form}
				name="confirm"
				label="Confirm password"
				type="password"
				placeholder="Repeat it"
				autocomplete="new-password"
			/>
			<TextField {form} name="workspace" label="Workspace name" placeholder="Acme Corp" />
			<SlugField {form} name="slug" workspace={$formData.workspace} />
			<TimezoneSelect {form} name="timezone" label="Workspace timezone" />

			<Button type="submit" class="w-full">Create admin account</Button>
		</form>
	</div>
</AuthShell>
