<script lang="ts">
	import AuthShell from '$lib/components/auth/auth-shell.svelte';
	import CodeForm from '$lib/components/auth/code-form.svelte';
	import RecoveryForm from './recovery-form.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<AuthShell
	title="Two-factor check"
	subtitle={data.mode === 'recovery'
		? 'Enter one of your recovery codes. Each code works once.'
		: 'Enter the 6-digit code from your authenticator app.'}
>
	{#snippet footer()}
		<span>
			Signed in as {data.email} —
			<a href="/login" class="text-brand-foreground hover:underline">not you?</a>
		</span>
	{/snippet}

	{#if data.mode === 'recovery'}
		<RecoveryForm data={data.form} />
		<a
			href="/two-factor"
			class="text-muted-foreground hover:text-brand-foreground mx-auto mt-4 block text-[13px]"
		>
			Use an authenticator code instead
		</a>
	{:else}
		<CodeForm data={data.form} />
		<a
			href="/two-factor?mode=recovery"
			class="text-muted-foreground hover:text-brand-foreground mx-auto mt-4 block text-[13px]"
		>
			Use a recovery code
		</a>
	{/if}
</AuthShell>
