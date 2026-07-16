<script lang="ts">
	import RotateCwIcon from '@lucide/svelte/icons/rotate-cw';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { toast } from 'svelte-sonner';
	import AuthShell from '$lib/components/auth/auth-shell.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const details = $derived(
		[
			`error=${data.failure.code}`,
			`entity_id=${data.failure.entityId}`,
			`request_id=${data.failure.requestId}`,
			`timestamp=${data.failure.at}`
		].join('\n')
	);

	async function copy() {
		await navigator.clipboard.writeText(details);
		toast.success('Details copied.');
	}
</script>

<AuthShell
	title="Single sign-on failed"
	subtitle="Your identity provider returned an error, so we couldn't complete the login. Nothing is wrong with your account."
>
	{#snippet footer()}
		<span>Back to <a href="/login" class="text-brand-foreground hover:underline">log in</a></span>
	{/snippet}

	<div class="flex flex-col gap-5">
		<Alert.Root tone="warning">
			<TriangleAlertIcon />
			<Alert.Content>
				<Alert.Title>SAML response rejected</Alert.Title>
				<Alert.Description>
					The signature on the SAML response didn't match the certificate configured for this
					workspace. This usually means the certificate was rotated on the identity provider but not
					updated in Opsybot.
				</Alert.Description>
			</Alert.Content>
		</Alert.Root>

		<div>
			<div class="text-subtle-foreground tracking-label mb-2 text-[11px] uppercase">
				What to tell your admin
			</div>
			<pre
				class="bg-inset text-muted-foreground m-0 overflow-x-auto rounded-md border px-4 py-3.5 font-mono text-xs leading-[1.7]">{details}</pre>
		</div>

		<div class="flex gap-2.5">
			<Button href="/login">
				<RotateCwIcon data-icon="inline-start" />
				Retry SSO
			</Button>
			<Button variant="ghost" onclick={copy}>Copy details</Button>
		</div>
	</div>
</AuthShell>
