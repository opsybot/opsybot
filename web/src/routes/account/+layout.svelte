<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import UserRoundIcon from '@lucide/svelte/icons/user-round';
	import { page } from '$app/state';
	import ThemeToggle from '$lib/components/layout/theme-toggle.svelte';
	import './account.css';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();

	const initials = $derived(
		data.user.name
			.split(' ')
			.map((p) => p[0])
			.filter(Boolean)
			.slice(0, 2)
			.join('')
			.toUpperCase() || 'U'
	);

	const path = $derived(page.url.pathname);
	const tab = $derived(
		path.startsWith('/account/security')
			? 'security'
			: path.startsWith('/account/channels')
				? 'channels'
				: path.startsWith('/account/sessions')
					? 'sessions'
					: 'profile'
	);

	const tabs = [
		['profile', 'Profile', '/account'],
		['security', 'Security', '/account/security'],
		['channels', 'Channels', '/account/channels'],
		['sessions', 'Sessions', '/account/sessions']
	] as const;
</script>

<svelte:head>
	<title>Account · Opsybot</title>
</svelte:head>

<div class="acct-body-bg">
	<div class="acct-topbar">
		<span class="acct-wordmark">opsy<span style="color: var(--mint-500)">.</span>bot</span>
		<span class="acct-badge"><UserRoundIcon size={12} />{data.user.email} · personal account</span>
		<div style="flex: 1"></div>
		<ThemeToggle />
		{#if data.back}
			<a class="acct-back" href="/{data.back.id}"><ArrowLeftIcon size={13} /> Back to {data.back.name}</a>
		{/if}
	</div>

	<div class="acct-wrap">
		<div class="acct-hero">
			<span class="acct-crest">{initials}</span>
			<div>
				<h1 class="acct-h1">{data.user.name}</h1>
				<div class="acct-mono" style="font-size: 12px; margin-top: 3px">
					{data.user.email} · personal account
				</div>
			</div>
		</div>

		<div class="acct-tabs">
			{#each tabs as [id, label, href] (id)}
				<a class="acct-tab {tab === id ? 'is-on' : ''}" {href}>{label}</a>
			{/each}
		</div>

		{@render children()}
	</div>
</div>
