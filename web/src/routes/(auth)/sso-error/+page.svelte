<script lang="ts">
	import RotateCwIcon from '@lucide/svelte/icons/rotate-cw';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { toast } from 'svelte-sonner';
	import AuthShell from '$lib/components/auth/auth-shell.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const MESSAGES: Record<string, { title: string; detail: string }> = {
		not_enabled: {
			title: 'SSO is off for this workspace',
			detail: 'An admin hasn’t turned on single sign-on here yet. Sign in with a password, or ask an admin to enable SSO.'
		},
		not_found: {
			title: 'Workspace not found',
			detail: 'That workspace URL doesn’t exist or has no SSO configured. Check the URL and try again.'
		},
		invalid_state: {
			title: 'Sign-in expired',
			detail: 'The sign-in took too long or was retried in another tab. Start again from the log-in page.'
		},
		exchange_failed: {
			title: 'Your identity provider rejected the sign-in',
			detail: 'The handshake with your provider failed. Try again — if it keeps failing, ask your admin to check the SSO settings.'
		},
		not_provisioned: {
			title: 'No account yet',
			detail: 'Your provider signed you in, but there’s no Opsybot account and automatic account creation is off. Ask an admin to invite you.'
		},
		domain_not_allowed: {
			title: 'Email domain not allowed',
			detail: 'Your email domain isn’t on this workspace’s allowed list. Ask an admin to add it.'
		},
		deactivated: {
			title: 'Account deactivated',
			detail: 'This account was deactivated by a workspace admin. Contact your admin to restore access.'
		},
		email_missing: {
			title: 'No email from your provider',
			detail: 'Your identity provider didn’t return an email address, which Opsybot needs to identify you.'
		},
		email_unverified: {
			title: 'Email not verified',
			detail: 'Your identity provider reported your email as unverified. Verify it there, then try again.'
		},
		idp_error: {
			title: 'Identity provider error',
			detail: 'Your identity provider returned an error. Try again — if it persists, ask your admin.'
		}
	};
	const FALLBACK = {
		title: 'Single sign-on failed',
		detail: 'We couldn’t complete the login. Try again — if it persists, contact your workspace admin.'
	};

	const message = $derived(MESSAGES[data.code] ?? FALLBACK);

	async function copy() {
		await navigator.clipboard.writeText(`error=${data.code}`);
		toast.success('Error code copied.');
	}
</script>

<AuthShell
	title="Single sign-on failed"
	subtitle="We couldn’t complete the login. Nothing is wrong with your account — this is between Opsybot and your identity provider."
>
	{#snippet footer()}
		<span>Back to <a href="/login" class="text-brand-foreground hover:underline">log in</a></span>
	{/snippet}

	<div class="flex flex-col gap-5">
		<Alert.Root tone="warning">
			<TriangleAlertIcon />
			<Alert.Content>
				<Alert.Title>{message.title}</Alert.Title>
				<Alert.Description>{message.detail}</Alert.Description>
			</Alert.Content>
		</Alert.Root>

		<div>
			<div class="text-subtle-foreground tracking-label mb-2 text-[11px] uppercase">Error code</div>
			<pre
				class="bg-inset text-muted-foreground m-0 overflow-x-auto rounded-md border px-4 py-3.5 font-mono text-xs leading-[1.7]">error={data.code}</pre>
		</div>

		<div class="flex gap-2.5">
			<Button href="/login">
				<RotateCwIcon data-icon="inline-start" />
				Back to log in
			</Button>
			<Button variant="ghost" onclick={copy}>Copy code</Button>
		</div>
	</div>
</AuthShell>
