<script lang="ts">
	import type { Component } from 'svelte';
	import type { LucideProps } from '@lucide/svelte';
	import CalendarClockIcon from '@lucide/svelte/icons/calendar-clock';
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import DatabaseIcon from '@lucide/svelte/icons/database';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import HardDriveIcon from '@lucide/svelte/icons/hard-drive';
	import HeartPulseIcon from '@lucide/svelte/icons/heart-pulse';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';
	import RotateCwIcon from '@lucide/svelte/icons/rotate-cw';
	import ServerIcon from '@lucide/svelte/icons/server';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import WorkflowIcon from '@lucide/svelte/icons/workflow';
	import { toast } from 'svelte-sonner';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import { ws } from '$lib/navigation';
	import { depthClass, dotColor, type HealthState } from '$lib/operations';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const SUBSYSTEM_ICON: Record<string, Component<LucideProps>> = {
		server: ServerIcon,
		database: DatabaseIcon,
		'hard-drive': HardDriveIcon,
		workflow: WorkflowIcon,
		'calendar-clock': CalendarClockIcon
	};

	function reRun() {
		toast.success(
			data.overall.degraded
				? 'Re-ran all checks: worker-3 back to healthy in a moment.'
				: 'Re-ran all checks: all systems still healthy.'
		);
	}
</script>

{#snippet dot(state: HealthState)}
	<span class="size-2 shrink-0 rounded-full" style="background: {dotColor(state)}"></span>
{/snippet}

<div class="flex flex-col gap-3.5">
	<div
		class="flex items-center gap-3 rounded-xl border px-4 py-3.5 {data.overall.degraded
			? 'bg-warning-wash border-warning-edge'
			: 'bg-card'}"
	>
		<span class="bg-inset flex size-[34px] shrink-0 items-center justify-center rounded-full border">
			{#if data.overall.degraded}
				<TriangleAlertIcon class="size-[18px] text-[var(--warning)]" />
			{:else}
				<CircleCheckIcon class="size-[18px] text-[var(--success)]" />
			{/if}
		</span>
		<div class="flex-1">
			<div class="text-[14px] font-semibold">{data.overall.title}</div>
			<div class="text-muted-foreground mt-px text-[12px]">{data.overall.detail}</div>
		</div>
		<Button size="sm" variant="secondary" onclick={reRun}>
			<RotateCwIcon data-icon="inline-start" />
			Re-run checks
		</Button>
	</div>

	<div class="grid items-start gap-3.5 [grid-template-columns:1fr_1fr] max-[900px]:grid-cols-1">
		<div class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-[11px]">
				<span class="text-[13.5px] font-semibold">Subsystems</span>
			</header>
			<div>
				{#each data.subsystems as system (system.id)}
					{@const Icon = SUBSYSTEM_ICON[system.icon]}
					<div class="flex items-center gap-2.5 border-t px-4 py-2.5 first:border-t-0" data-subsystem={system.id}>
						<Icon class="text-subtle-foreground size-[15px] shrink-0" />
						<span class="w-[130px] text-[13px] font-medium">{system.label}</span>
						<span class="text-subtle-foreground flex-1 font-mono text-[11.5px]">{system.detail}</span>
						{@render dot(system.state)}
					</div>
				{/each}
			</div>
		</div>

		<div class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-[11px]">
				<span class="text-[13.5px] font-semibold">Queue depths</span>
			</header>
			<div>
				{#each data.queues as queue (queue.name)}
					<div class="flex items-center gap-2.5 border-t px-4 py-2.5 first:border-t-0" data-queue={queue.name}>
						{@render dot(queue.state)}
						<span class="text-foreground flex-1 font-mono text-[12px]">{queue.name}</span>
						<span class="text-subtle-foreground font-mono text-[11px]">{queue.rate}</span>
						<span class="w-[54px] text-right font-mono text-[12.5px] {depthClass(queue.state)}">{queue.depth}</span>
					</div>
				{/each}
			</div>
		</div>

		<div class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-[11px]">
				<span class="text-[13.5px] font-semibold">Last successful notification</span>
			</header>
			<div>
				{#each data.channels as channel (channel.ch)}
					<div class="flex items-center gap-2.5 border-t px-4 py-2.5 first:border-t-0" data-channel={channel.ch}>
						{@render dot(channel.state)}
						<span class="flex-1 text-[13px]">{channel.ch}</span>
						<span class="text-subtle-foreground font-mono text-[11.5px]">{channel.last}</span>
					</div>
				{:else}
					<p class="text-subtle-foreground m-0 px-4 py-6 text-center text-[12.5px]">No notifications sent yet.</p>
				{/each}
			</div>
		</div>

		<div class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-[11px]">
				<span class="text-[13.5px] font-semibold">Integration connectivity</span>
			</header>
			<div>
				{#each data.integrations as integration (integration.name)}
					<div class="flex items-center gap-2.5 border-t px-4 py-2.5 first:border-t-0" data-integration={integration.name}>
						{@render dot(integration.state)}
						<span class="text-foreground flex-1 font-mono text-[12px]">{integration.name}</span>
						<span class="text-subtle-foreground font-mono text-[11px]">{integration.detail}</span>
					</div>
				{:else}
					<p class="text-subtle-foreground m-0 px-4 py-6 text-center text-[12.5px]">No integrations connected yet.</p>
				{/each}
			</div>
		</div>
	</div>

	<div class="grid gap-3.5 [grid-template-columns:1fr_1fr] max-[900px]:grid-cols-1">
		<div class="bg-card flex items-center gap-3 rounded-xl border p-4">
			<span class="bg-inset flex size-[34px] shrink-0 items-center justify-center rounded-full border">
				<KeyRoundIcon class="size-4 {data.license.tone === 'success' ? 'text-[var(--success)]' : 'text-muted-foreground'}" />
			</span>
			<div class="flex-1">
				<div class="text-[13px] font-semibold">{data.license.title}</div>
				<div class="text-subtle-foreground mt-px font-mono text-[11px]">{data.license.detail}</div>
			</div>
			<a href={ws('/billing/license')} class="text-brand-foreground text-[12.5px] hover:underline">Manage</a>
		</div>

		<div class="bg-card flex items-center gap-3 rounded-xl border p-4">
			<span class="bg-inset flex size-[34px] shrink-0 items-center justify-center rounded-full border">
				{#if data.update.latest}
					<DownloadIcon class="text-brand-foreground size-4" />
				{:else}
					<CircleCheckIcon class="size-4 text-[var(--success)]" />
				{/if}
			</span>
			<div class="flex-1">
				<div class="text-[13px] font-semibold">{data.update.latest ? 'Update available' : 'Up to date'}</div>
				<div class="text-subtle-foreground mt-px font-mono text-[11px]">
					{#if data.update.latest}
						running {data.update.current} · {data.update.latest} released {data.update.released}
					{:else}
						running {data.update.current} · latest
					{/if}
				</div>
			</div>
			{#if data.update.latest}
				<Button size="sm" variant="secondary" onclick={() => toast(`Release notes for ${data.update.latest} open in a new tab.`)}>
					Release notes
				</Button>
			{/if}
		</div>
	</div>

	<Alert.Root tone="info">
		<HeartPulseIcon />
		<Alert.Content>
			<Alert.Title>Cross-check that Opsybot itself is up</Alert.Title>
			<Alert.Description>
				Run an external heartbeat that pings a dead-man's-switch elsewhere every minute, so you're paged through a
				second path if this instance goes dark.
				<button
					type="button"
					class="text-brand-foreground hover:underline"
					onclick={() => toast('Self-monitoring setup guide opens in the docs.')}
				>
					Read the self-monitoring guide</button
				>.
			</Alert.Description>
		</Alert.Content>
	</Alert.Root>
</div>
