<script lang="ts">
	import { untrack } from 'svelte';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import AuthShell from '$lib/components/auth/auth-shell.svelte';
	import TextField from '$lib/components/text-field.svelte';
	import { Button } from '$lib/components/ui/button';
	import { resetPasswordSchema } from '$lib/schemas/auth';
	import { formatUtc } from '$lib/time';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const form = superForm(untrack(() => data.form), { validators: zod4Client(resetPasswordSchema) });
</script>

{#if data.state === 'expired'}
	<AuthShell
		title="This link has expired"
		subtitle="Reset links are valid for 60 minutes. This one was issued {formatUtc(data.issuedAt)}."
	>
		{#snippet footer()}
			<span>Back to <a href="/login" class="text-brand-foreground hover:underline">log in</a></span>
		{/snippet}

		<Button href="/forgot-password" class="w-full">Request a new link</Button>
	</AuthShell>
{:else}
	<AuthShell title="Choose a new password" subtitle="You're resetting the password for {data.email}.">
		{#snippet footer()}
			<span>Back to <a href="/login" class="text-brand-foreground hover:underline">log in</a></span>
		{/snippet}

		<form method="POST" use:form.enhance class="flex flex-col gap-4">
			<TextField
				{form}
				name="password"
				label="New password"
				type="password"
				placeholder="12+ characters"
				hint="12 characters minimum. Your other sessions stay signed in."
				autocomplete="new-password"
			/>
			<TextField
				{form}
				name="confirm"
				label="Confirm new password"
				type="password"
				placeholder="Repeat it"
				autocomplete="new-password"
			/>
			<Button type="submit" class="w-full">Update password</Button>
		</form>
	</AuthShell>
{/if}
