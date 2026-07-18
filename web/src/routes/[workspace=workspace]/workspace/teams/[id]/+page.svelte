<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import ArrowUpRightIcon from '@lucide/svelte/icons/arrow-up-right';
	import BoxesIcon from '@lucide/svelte/icons/boxes';
	import CalendarClockIcon from '@lucide/svelte/icons/calendar-clock';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import TeamEditDialog from '$lib/components/admin/team-edit-dialog.svelte';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let editOpen = $state(false);
	const roster = $derived(data.members.filter((member) => !member.deactivated));

	function onArchiveAction(msg: string) {
		return () =>
			async ({ result, update }: { result: { type: string; data?: Record<string, unknown> }; update: () => Promise<void> }) => {
				await update();
				if (result.type === 'success') toast.success(msg);
				else if (result.type === 'failure') toast.error(String(result.data?.error ?? 'Something went wrong.'));
				else if (result.type === 'error') toast.error('Something went wrong. Try again.');
			};
	}
	const rails = $derived([
		{ label: 'Schedules', icon: CalendarClockIcon, items: data.team.schedules, href: '/on-call' },
		{ label: 'Escalation policies', icon: ArrowUpRightIcon, items: data.team.policies, href: '/escalation-policies' },
		{ label: 'Services', icon: BoxesIcon, items: data.team.services, href: '/catalog' }
	]);
</script>

<div class="flex max-w-[760px] flex-col gap-3.5">
	<a
		href={ws('/workspace/teams')}
		class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[12.5px] transition-colors"
	>
		<ArrowLeftIcon class="size-3.5" />
		Teams
	</a>

	<div class="flex items-center gap-2.5">
		<h2 class="text-foreground m-0 font-mono text-[18px] font-semibold">{data.team.name}</h2>
		<Badge tone="neutral" size="sm">{data.team.members.length} members</Badge>
		{#if data.team.archived}<Badge tone="warning" size="sm">archived</Badge>{/if}
		<div class="flex-1"></div>
		{#if data.team.archived}
			<form method="POST" action="?/unarchive" use:enhance={onArchiveAction('Team restored.')}>
				<Button size="sm" variant="secondary" type="submit">Restore team</Button>
			</form>
		{:else}
			<form method="POST" action="?/archive" use:enhance={onArchiveAction('Team archived.')}>
				<Button size="sm" variant="ghost" type="submit">Archive</Button>
			</form>
			<Button size="sm" variant="secondary" onclick={() => (editOpen = true)}>
				<PencilIcon data-icon="inline-start" />
				Edit team
			</Button>
		{/if}
	</div>

	<div class="grid items-start gap-3.5 [grid-template-columns:1fr_1fr] max-[900px]:grid-cols-1">
		<div class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[13px] font-semibold">Members</span>
			</header>
			{#each data.team.members as name (name)}
				<div class="flex items-center gap-2.5 border-t px-3.5 py-2.5 first:border-t-0">
					<UserAvatar {name} size="xs" />
					<span class="text-[13px]">{name}</span>
				</div>
			{/each}
		</div>

		<div class="flex flex-col gap-3.5">
			{#each rails as rail (rail.label)}
				<div class="bg-card overflow-hidden rounded-xl border">
					<header class="flex items-center gap-2 border-b px-4 py-3">
						<rail.icon class="text-subtle-foreground size-3.5" />
						<span class="text-[13px] font-semibold">{rail.label}</span>
					</header>
					{#each rail.items as item (item)}
						<div class="border-t px-3.5 py-2.5 first:border-t-0">
							<a href={ws(rail.href)} class="text-brand-foreground font-mono text-[12.5px] hover:underline">{item}</a>
						</div>
					{/each}
				</div>
			{/each}
		</div>
	</div>
</div>

<TeamEditDialog bind:open={editOpen} action="?/save" team={data.team} {roster} />
