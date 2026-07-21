<script lang="ts">
	import { untrack } from 'svelte';
	import InfoIcon from '@lucide/svelte/icons/info';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import AuthShell from '$lib/components/auth/auth-shell.svelte';
	import OrDivider from '$lib/components/auth/or-divider.svelte';
	import TextField from '$lib/components/text-field.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Form from '$lib/components/ui/form';
	import { Input } from '$lib/components/ui/input';
	import { loginSchema } from '$lib/schemas/auth';
	import { slugify } from '$lib/slug';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const form = superForm(untrack(() => data.form), { validators: zod4Client(loginSchema) });
	const { form: formData, message, submitting, enhance } = form;

	const wrong = $derived($message === 'invalid');
	const deactivated = $derived($message === 'deactivated');
	const ssoRequired = $derived($message === 'sso-required');

	let ssoWorkspace = $state('');
	const ssoSlug = $derived(ssoWorkspace.trim() ? slugify(ssoWorkspace) : '');
</script>

{#snippet ssoStart()}
	<div class="flex flex-col gap-2">
		<label class="text-muted-foreground text-[13px] font-medium" for="sso-ws">Workspace URL</label>
		<div class="flex gap-2">
			<Input
				id="sso-ws"
				bind:value={ssoWorkspace}
				placeholder="acme"
				autocapitalize="none"
				autocorrect="off"
				spellcheck={false}
				class="flex-1"
			/>
			<Button
				href={ssoSlug ? `/sso/${ssoSlug}/start` : undefined}
				data-sveltekit-reload
				variant="secondary"
				aria-disabled={!ssoSlug}
				class={ssoSlug ? '' : 'pointer-events-none opacity-50'}
			>
				<KeyRoundIcon data-icon="inline-start" />
				Continue with SSO
			</Button>
		</div>
	</div>
{/snippet}

<AuthShell title="Log in">
	{#snippet footer()}
		{#if data.deployment === 'cloud'}
			<span>
				New to Opsybot?
				<a href="/signup" class="text-brand-foreground hover:underline">Create a workspace</a>
			</span>
		{:else}
			<span>Accounts on this instance are created by invite. Ask your instance admin.</span>
		{/if}
	{/snippet}

	{#if wrong}
		<Alert.Root tone="critical" class="mb-5">
			<OctagonAlertIcon />
			<Alert.Content>
				<Alert.Description>
					Email or password is incorrect. Check both and try again.
				</Alert.Description>
			</Alert.Content>
		</Alert.Root>
	{/if}

	{#if deactivated}
		<Alert.Root tone="critical" class="mb-5">
			<OctagonAlertIcon />
			<Alert.Content>
				<Alert.Title>Account deactivated</Alert.Title>
				<Alert.Description>
					This account was deactivated by a workspace admin. Contact your admin to restore access.
				</Alert.Description>
			</Alert.Content>
		</Alert.Root>
	{/if}

	{#if ssoRequired}
		<div class="flex flex-col gap-5">
			<Alert.Root tone="info">
				<InfoIcon />
				<Alert.Content>
					<Alert.Title>Single sign-on required</Alert.Title>
					<Alert.Description>
						This account's workspace requires SSO. Password login is disabled. Enter your workspace
						URL to continue to your identity provider.
					</Alert.Description>
				</Alert.Content>
			</Alert.Root>

			{@render ssoStart()}
		</div>
	{:else}
		<form method="POST" use:enhance class="flex flex-col gap-4">
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
				placeholder="••••••••••"
				autocomplete="current-password"
			/>

			<div class="flex items-center justify-between">
				<Form.Field {form} name="remember" class="w-auto space-y-0">
					<Form.Control>
						{#snippet children({ props })}
							<div class="flex items-center gap-2.5">
								<Checkbox {...props} bind:checked={$formData.remember} />
								<Form.Label class="text-foreground text-sm font-normal">Remember me</Form.Label>
							</div>
						{/snippet}
					</Form.Control>
				</Form.Field>

				<a href="/forgot-password" class="text-brand-foreground text-[13px] hover:underline">
					Forgot password?
				</a>
			</div>

			<Button type="submit" class="w-full" disabled={deactivated || $submitting}>Log in</Button>

			<OrDivider />

			{@render ssoStart()}

			<p class="text-subtle-foreground m-0 text-xs leading-[1.5]">
				Enter your workspace URL to sign in with your identity provider (OIDC or SAML).
			</p>
		</form>
	{/if}
</AuthShell>
