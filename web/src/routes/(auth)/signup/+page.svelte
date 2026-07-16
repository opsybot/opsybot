<script lang="ts">
	import AuthShell from '$lib/components/auth/auth-shell.svelte';
	import StepDots from '$lib/components/auth/step-dots.svelte';
	import AccountStep from './account-step.svelte';
	import WorkspaceStep from './workspace-step.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const onWorkspace = $derived(data.step === 'workspace');
</script>

<AuthShell
	title={onWorkspace ? 'Create your workspace' : 'Create your account'}
	subtitle={onWorkspace
		? 'The workspace is where your team, schedules, and incidents live.'
		: 'No card required. You can invite your team once the workspace exists.'}
>
	{#snippet footer()}
		<span>
			Already have an account?
			<a href="/login" class="text-brand-foreground hover:underline">Log in</a>
		</span>
	{/snippet}

	<StepDots step={onWorkspace ? 2 : 1} total={2} />

	{#if data.step === 'workspace'}
		<WorkspaceStep data={data.form} />
	{:else}
		<AccountStep data={data.form} />
	{/if}
</AuthShell>
