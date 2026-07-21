<script lang="ts">
	import { untrack } from 'svelte';
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import AuthShell from '$lib/components/auth/auth-shell.svelte';
	import SlugField from '$lib/components/auth/slug-field.svelte';
	import StepDots from '$lib/components/auth/step-dots.svelte';
	import TextField from '$lib/components/text-field.svelte';
	import TimezoneSelect from '$lib/components/timezone-select.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import { signupSchema } from '$lib/schemas/auth';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const accountFields = ['name', 'email', 'password', 'confirm'] as const;
	const hasAccountError = (e: Record<string, string[] | undefined>) =>
		accountFields.some((field) => e[field]?.length);

	let step = $state(1);

	// dataType 'json' so step one still posts once its inputs leave the DOM
	const form = superForm(untrack(() => data.form), {
		dataType: 'json',
		validators: zod4Client(signupSchema),
		onUpdated: ({ form: result }) => {
			if (hasAccountError(result.errors as Record<string, string[] | undefined>)) step = 1;
		}
	});
	const { form: formData, message, errors, validateForm, submitting } = form;

	async function next() {
		const result = await validateForm({ update: false });
		const e = result.errors as Record<string, string[] | undefined>;
		$errors = {
			...$errors,
			name: e.name,
			email: e.email,
			password: e.password,
			confirm: e.confirm
		};
		if (!hasAccountError(e)) step = 2;
	}
</script>

<AuthShell
	title={step === 2 ? 'Create your workspace' : 'Create your account'}
	subtitle={step === 2
		? 'The workspace is where your team, schedules, and incidents live.'
		: 'No card required. You can invite your team once the workspace exists.'}
>
	{#snippet footer()}
		<span>
			Already have an account?
			<a href="/login" class="text-brand-foreground hover:underline">Log in</a>
		</span>
	{/snippet}

	<StepDots step={step === 2 ? 2 : 1} total={2} />

	{#if $message}
		<Alert.Root tone="critical" class="mb-5">
			<OctagonAlertIcon />
			<Alert.Content>
				<Alert.Description>{$message}</Alert.Description>
			</Alert.Content>
		</Alert.Root>
	{/if}

	<form method="POST" use:form.enhance class="flex flex-col gap-4">
		{#if step === 1}
			<TextField {form} name="name" label="Name" placeholder="Maya Chen" autocomplete="name" />
			<TextField
				{form}
				name="email"
				label="Work email"
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
			<Button type="button" class="w-full" onclick={next}>Continue</Button>
		{:else}
			<TextField
				{form}
				name="workspace"
				label="Workspace name"
				placeholder="Acme Corp"
				hint="Usually your company or team name."
			/>
			<SlugField {form} name="slug" workspace={$formData.workspace} />
			<TimezoneSelect {form} name="timezone" label="Workspace timezone" />
			<Button type="submit" class="w-full" disabled={$submitting}>Create workspace</Button>
			<button
				type="button"
				onclick={() => (step = 1)}
				class="text-muted-foreground hover:text-brand-foreground mx-auto block text-[13px]"
			>
				Back to account details
			</button>
		{/if}
	</form>
</AuthShell>
