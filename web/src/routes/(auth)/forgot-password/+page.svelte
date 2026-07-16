<script lang="ts">
	import { untrack } from 'svelte';
	import MailIcon from '@lucide/svelte/icons/mail';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import AuthShell from '$lib/components/auth/auth-shell.svelte';
	import TextField from '$lib/components/text-field.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import { forgotPasswordSchema } from '$lib/schemas/auth';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const form = superForm(untrack(() => data.form), { validators: zod4Client(forgotPasswordSchema) });
	const { form: formData, message } = form;

	let reopened = $state(false);
	const showConfirmation = $derived($message === 'sent' && !reopened);
</script>

{#if showConfirmation}
	<AuthShell title="Check your email">
		{#snippet footer()}
			<span>Back to <a href="/login" class="text-brand-foreground hover:underline">log in</a></span>
		{/snippet}

		<div class="flex flex-col gap-5">
			<Alert.Root tone="neutral">
				<MailIcon />
				<Alert.Content>
					<Alert.Description>
						If an account exists for
						<strong class="text-foreground font-semibold">{$formData.email}</strong>, a reset link is
						on its way. It's valid for 60 minutes. Check spam before requesting another.
					</Alert.Description>
				</Alert.Content>
			</Alert.Root>

			<Button variant="secondary" class="w-full" onclick={() => (reopened = true)}>
				Use a different email
			</Button>
		</div>
	</AuthShell>
{:else}
	<AuthShell
		title="Reset your password"
		subtitle="Enter the email you log in with and we'll send a reset link."
	>
		{#snippet footer()}
			<span>Back to <a href="/login" class="text-brand-foreground hover:underline">log in</a></span>
		{/snippet}

		<form method="POST" use:form.enhance class="flex flex-col gap-4">
			<TextField
				{form}
				name="email"
				label="Email"
				type="email"
				placeholder="you@company.com"
				autocomplete="email"
			/>
			<Button type="submit" class="w-full">Send reset link</Button>
		</form>
	</AuthShell>
{/if}
