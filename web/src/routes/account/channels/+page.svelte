<script lang="ts">
	import { untrack } from 'svelte';
	import BellIcon from '@lucide/svelte/icons/bell';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import { toast } from 'svelte-sonner';
	import { enhance as formEnhance } from '$app/forms';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { channelSchema } from '$lib/schemas/auth';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const CHANNEL_TYPES = [
		['slack', 'Slack'],
		['teams', 'Microsoft Teams'],
		['discord', 'Discord'],
		['telegram', 'Telegram'],
		['ntfy', 'ntfy'],
		['email', 'Email'],
		['webhook', 'Webhook']
	] as const;

	const LABELS: Record<string, string> = Object.fromEntries(CHANNEL_TYPES);
	const PLACEHOLDERS: Record<string, string> = {
		slack: '@you or U0123ABCD',
		teams: 'you@company.com',
		discord: 'you#0001 or a channel URL',
		telegram: '@you or a chat id',
		ntfy: 'https://ntfy.sh/your-topic',
		email: 'you@company.com',
		webhook: 'https://hooks.example.com/…'
	};

	const form = superForm(untrack(() => data.form), {
		validators: zod4Client(channelSchema),
		onUpdated: ({ form }) => {
			if (form.message === 'added')
				toast.success('Channel added', { description: 'Verify it to start receiving alerts there.' });
			else if (typeof form.message === 'string') toast.error(form.message);
		}
	});
	const { form: formData, errors, enhance, submitting } = form;

	function rowAction(successMsg: string) {
		return () =>
			async ({ result, update }: { result: { type: string; data?: Record<string, unknown> }; update: () => Promise<void> }) => {
				await update();
				if (result.type === 'success') toast.success(successMsg);
				else if (result.type === 'failure') toast.error(String(result.data?.error ?? 'Something went wrong.'));
				else if (result.type === 'error') toast.error('Something went wrong. Try again.');
			};
	}
</script>

<div style="display: flex; flex-direction: column; gap: 18px">
	<div class="acct-card">
		<header class="acct-card-head">
			<span class="acct-card-title">Notification channels</span>
			<span style="font-size: 11.5px; color: var(--text-tertiary); margin-left: 4px">
				where Opsybot reaches you · shared across workspaces
			</span>
		</header>

		{#if data.channels.length === 0}
			<div class="acct-row" style="background: var(--ink-2)">
				<span style="font-size: 12px; color: var(--text-tertiary); flex: 1">
					No channels yet. Add one below so alerts and on-call pages can reach you.
				</span>
			</div>
		{:else}
			{#each data.channels as ch (ch.id)}
				<div class="acct-row">
					<span class="acct-ic"><BellIcon size={16} /></span>
					<div style="min-width: 0; flex: 1">
						<div style="font-size: 13.5px; font-weight: 500; display: flex; align-items: center; gap: 8px">
							{LABELS[ch.type] ?? ch.type}
							{#if ch.verified}
								<Badge tone="success" size="sm">verified</Badge>
							{:else}
								<Badge tone="warning" size="sm">unverified</Badge>
							{/if}
						</div>
						<div class="acct-mono" style="font-size: 11px; margin-top: 2px">{ch.detail}</div>
					</div>
					{#if !ch.verified}
						<form method="POST" action="?/verify" use:formEnhance={rowAction('Verification sent. Confirm it from My notifications.')}>
							<input type="hidden" name="id" value={ch.id} />
							<Button size="sm" variant="secondary" type="submit">Verify</Button>
						</form>
					{/if}
					<form method="POST" action="?/remove" use:formEnhance={rowAction('Channel removed.')}>
						<input type="hidden" name="id" value={ch.id} />
						<Button size="sm" variant="ghost" type="submit">Remove</Button>
					</form>
				</div>
			{/each}
		{/if}
	</div>

	<div class="acct-card">
		<header class="acct-card-head"><span class="acct-card-title">Add a channel</span></header>
		<form method="POST" action="?/add" use:enhance class="acct-card-body">
			<div class="acct-grid2">
				<div>
					<label class="acct-field-label" for="ch-type">Channel</label>
					<select
						id="ch-type"
						name="type"
						bind:value={$formData.type}
						class="border-border-strong bg-inset text-foreground focus-visible:border-primary focus-visible:shadow-[var(--focus-ring)] h-10 w-full rounded-sm border px-3 text-sm outline-none"
					>
						{#each CHANNEL_TYPES as [value, label] (value)}
							<option {value}>{label}</option>
						{/each}
					</select>
				</div>
				<div>
					<label class="acct-field-label" for="ch-detail">Address or URL</label>
					<Input
						id="ch-detail"
						name="detail"
						bind:value={$formData.detail}
						placeholder={PLACEHOLDERS[$formData.type] ?? ''}
						autocomplete="off"
					/>
					{#if $errors.detail}<div class="acct-err">{$errors.detail}</div>{/if}
				</div>
			</div>
			<div style="display: flex">
				<Button type="submit" disabled={$submitting}>
					<PlusIcon class="size-4" />
					Add channel
				</Button>
			</div>
		</form>
	</div>
</div>
