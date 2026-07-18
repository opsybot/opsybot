<script lang="ts">
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import MonitorIcon from '@lucide/svelte/icons/monitor';
	import MonitorSmartphoneIcon from '@lucide/svelte/icons/monitor-smartphone';
	import SmartphoneIcon from '@lucide/svelte/icons/smartphone';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import ConfirmDialog from '$lib/components/account/confirm-dialog.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import type { components } from '$lib/api/schema';
	import type { PageProps } from './$types';

	type SessionRow = components['schemas']['Session'];

	let { data }: PageProps = $props();

	function parseUA(ua?: string): { label: string; mobile: boolean } {
		if (!ua) return { label: 'Unknown device', mobile: false };
		const mobile = /Mobile|Android|iPhone|iPad/i.test(ua);
		let browser = 'Browser';
		if (/Edg\//.test(ua)) browser = 'Edge';
		else if (/Chrome\//.test(ua)) browser = 'Chrome';
		else if (/Firefox\//.test(ua)) browser = 'Firefox';
		else if (/Safari\//.test(ua)) browser = 'Safari';
		let os = '';
		if (/Windows/.test(ua)) os = 'Windows';
		else if (/Mac OS X|Macintosh/.test(ua)) os = 'macOS';
		else if (/iPhone|iPad/.test(ua)) os = 'iOS';
		else if (/Android/.test(ua)) os = 'Android';
		else if (/Linux/.test(ua)) os = 'Linux';
		return { label: [os, browser].filter(Boolean).join(' · ') || 'Device', mobile };
	}

	function fmtLast(s: SessionRow): string {
		if (s.current) return 'Active now';
		const d = new Date(s.lastSeenAt);
		if (Number.isNaN(d.getTime())) return 'recently';
		return d.toISOString().replace('T', ' ').slice(0, 16) + ' UTC';
	}

	const rows = $derived(
		data.sessions.map((s) => ({ ...s, ...parseUA(s.userAgent), last: fmtLast(s) }))
	);
	const others = $derived(data.sessions.filter((s) => !s.current).length);

	let revoking = $state<(SessionRow & { label: string }) | null>(null);
	let allOpen = $state(false);
	let revokeForm: HTMLFormElement;
	let revokeId = $state('');
	let othersForm: HTMLFormElement;

	function confirmRevoke() {
		if (!revoking) return;
		revokeId = revoking.id;
		revokeForm.requestSubmit();
	}
</script>

<div style="display: flex; flex-direction: column; gap: 18px">
	<div class="acct-note">
		<MonitorSmartphoneIcon size={15} style="color: var(--text-tertiary); flex-shrink: 0; margin-top: 1px" />
		<span>Everywhere you're signed in. Don't recognize one? <strong>Sign it out</strong> and change your password.</span>
	</div>

	<div class="acct-card">
		<header class="acct-card-head">
			<span class="acct-card-title">Active sessions</span>
			<span style="font-size: 11.5px; color: var(--text-tertiary); margin-left: 4px">{data.sessions.length} total</span>
			<div style="flex: 1"></div>
			{#if others > 0}
				<Button size="sm" variant="ghost" class="text-critical" onclick={() => (allOpen = true)}>
					<LogOutIcon class="size-3.5" />
					Sign out others
				</Button>
			{/if}
		</header>
		{#each rows as s (s.id)}
			<div class="acct-row">
				<span class="acct-ic">
					{#if s.mobile}<SmartphoneIcon size={17} />{:else}<MonitorIcon size={17} />{/if}
				</span>
				<div style="min-width: 0; flex: 1">
					<div style="font-size: 13.5px; font-weight: 500; display: flex; align-items: center; gap: 8px">
						{s.label}
						{#if s.current}<Badge tone="brand" size="sm">this device</Badge>{/if}
					</div>
					<div class="acct-mono" style="font-size: 11px; margin-top: 2px">
						{s.ip || 'unknown IP'} · {s.last}
					</div>
				</div>
				{#if s.current}
					<span style="font-size: 12px; color: var(--text-tertiary)">Current</span>
				{:else}
					<Button size="sm" variant="ghost" onclick={() => (revoking = s)}>Sign out</Button>
				{/if}
			</div>
		{/each}
	</div>
</div>

<ConfirmDialog
	open={!!revoking}
	tone="warning"
	icon={LogOutIcon}
	title={revoking ? `Sign out ${revoking.label}?` : ''}
	description="That device will need to sign in again. Anything unsaved there could be lost."
	confirmLabel="Sign it out"
	cancelLabel="Cancel"
	onConfirm={confirmRevoke}
	onCancel={() => (revoking = null)}
/>

<ConfirmDialog
	open={allOpen}
	tone="warning"
	icon={LogOutIcon}
	title="Sign out of all other sessions?"
	description="Every device except this one is signed out immediately. This one stays signed in."
	confirmLabel="Sign out others"
	cancelLabel="Cancel"
	onConfirm={() => othersForm.requestSubmit()}
	onCancel={() => (allOpen = false)}
/>

<form
	bind:this={revokeForm}
	method="POST"
	action="?/revoke"
	class="hidden"
	use:enhance={() =>
		async ({ result, update }) => {
			revoking = null;
			await update({ reset: false });
			if (result.type === 'success') toast.success('Session signed out.');
			else toast.error('Could not sign out that session.');
		}}
>
	<input type="hidden" name="id" value={revokeId} />
</form>

<form
	bind:this={othersForm}
	method="POST"
	action="?/revokeOthers"
	class="hidden"
	use:enhance={() =>
		async ({ result, update }) => {
			allOpen = false;
			await update({ reset: false });
			if (result.type === 'success') toast.success('Other sessions signed out.');
		}}
></form>
