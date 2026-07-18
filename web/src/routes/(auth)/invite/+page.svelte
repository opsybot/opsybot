<script lang="ts">
	import { untrack } from 'svelte';
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import AuthShell from '$lib/components/auth/auth-shell.svelte';
	import TextField from '$lib/components/text-field.svelte';
	import TimezoneSelect from '$lib/components/timezone-select.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import { inviteSchema } from '$lib/schemas/auth';
	import { formatUtcDate } from '$lib/time';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const form = superForm(untrack(() => data.form), { validators: zod4Client(inviteSchema) });
	const { message } = form;

	const invite = $derived(data.invite);
	const initials = $derived(
		invite.invitedBy
			.split(' ')
			.map((part) => part[0])
			.join('')
			.slice(0, 2)
	);
</script>

{#if data.state === 'expired'}
	<AuthShell
		title="This invite has expired"
		subtitle="Invites are valid for 7 days. Yours was sent {formatUtcDate(
			invite.sentAt
		)} and is no longer usable."
	>
		{#snippet footer()}
			<span>
				Already a member?
				<a href="/login" class="text-brand-foreground hover:underline">Log in</a>
			</span>
		{/snippet}

		<Alert.Root tone="warning">
			<TriangleAlertIcon />
			<Alert.Content>
				<Alert.Description>
					Ask {invite.invitedBy} — or any {invite.workspace} admin — to send a new invite to the same
					email address. Nothing else is needed on your side.
				</Alert.Description>
			</Alert.Content>
		</Alert.Root>
	</AuthShell>
{:else}
	<AuthShell title="Join {invite.workspace}">
		{#snippet footer()}
			<span>
				Already a member?
				<a href="/login" class="text-brand-foreground hover:underline">Log in</a>
			</span>
		{/snippet}

		<div class="flex flex-col gap-5">
			<div class="text-muted-foreground flex items-center gap-3 text-sm leading-[1.5]">
				<span
					class="bg-brand-wash border-brand-edge text-brand-foreground flex size-9 shrink-0 items-center justify-center rounded-full border text-xs font-semibold"
					aria-hidden="true"
				>
					{initials}
				</span>
				<span>
					<strong class="text-foreground font-semibold">{invite.invitedBy}</strong>
					invited you to the
					<strong class="text-foreground font-semibold">{invite.workspace}</strong>
					workspace as
					<span class="font-mono text-[12.5px]">{invite.email}</span>
				</span>
			</div>

			{#if $message}
				<Alert.Root tone="critical">
					<OctagonAlertIcon />
					<Alert.Content>
						<Alert.Description>{$message}</Alert.Description>
					</Alert.Content>
				</Alert.Root>
			{/if}

			<form method="POST" use:form.enhance class="flex flex-col gap-4">
				<TextField {form} name="name" label="Name" placeholder="Jordan Okafor" autocomplete="name" />
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
				<TimezoneSelect {form} name="timezone" label="Your timezone" />

				<Button type="submit" class="w-full">Join workspace</Button>
			</form>
		</div>
	</AuthShell>
{/if}
