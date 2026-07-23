<script lang="ts">
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const MESSAGES: Record<string, { title: string; detail: string }> = {
		invalid_state: {
			title: 'That connection link expired',
			detail: 'The install took too long or was retried in another tab. Open Integrations and click Connect again.'
		},
		denied: {
			title: 'Install cancelled',
			detail: 'You cancelled on the provider’s consent screen. Nothing changed. You can try again any time.'
		},
		forbidden: {
			title: 'You no longer have access',
			detail: 'Your permission to manage chat connections in that workspace changed before the install finished. Ask an admin.'
		},
		exchange_failed: {
			title: 'The provider rejected the install',
			detail: 'The handshake failed. Check the app’s redirect URL and credentials, then connect again.'
		},
		not_configured: {
			title: 'Provider not configured',
			detail: 'This provider isn’t set up on the server yet. An admin needs to add its app credentials.'
		},
		secret_unavailable: {
			title: 'Secret storage is off',
			detail: 'The server has no encryption key configured, so the token could not be saved. Ask an admin to set OPSYBOT_AUTH_SECRET_KEY.'
		}
	};
	const FALLBACK = {
		title: 'The connection could not be completed',
		detail: 'Something went wrong finishing the install. Open Integrations and try again.'
	};

	const message = $derived(MESSAGES[data.code] ?? FALLBACK);
</script>

<div class="mx-auto flex min-h-svh max-w-[440px] flex-col justify-center gap-5 px-6">
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

	<div>
		<Button href="/">Back to Opsybot</Button>
	</div>
</div>
