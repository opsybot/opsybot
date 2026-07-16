<script lang="ts">
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import UsersIcon from '@lucide/svelte/icons/users';
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import { Button } from '$lib/components/ui/button';
	import TeamEditDialog from '$lib/components/admin/team-edit-dialog.svelte';
	import { teamSummary } from '$lib/admin';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let editOpen = $state(false);
	const roster = $derived(data.members.filter((member) => !member.deactivated));
</script>

<div class="flex max-w-[760px] flex-col gap-3.5">
	<div class="flex items-center">
		<span class="text-subtle-foreground text-[13px]">{data.teams.length} teams</span>
		<div class="flex-1"></div>
		<Button size="sm" onclick={() => (editOpen = true)}>
			<PlusIcon data-icon="inline-start" />
			New team
		</Button>
	</div>

	<div class="bg-card overflow-hidden rounded-xl border">
		{#each data.teams as team (team.id)}
			<a
				href="/workspace/teams/{team.id}"
				data-team={team.id}
				class="hover:bg-accent flex items-center gap-3 border-t px-4 py-[13px] first:border-t-0"
			>
				<span
					class="bg-inset text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-sm border"
				>
					<UsersIcon class="size-[15px]" />
				</span>
				<div class="min-w-0 flex-1">
					<span class="text-foreground font-mono text-[13.5px] font-medium">{team.name}</span>
					<div class="text-subtle-foreground mt-0.5 font-mono text-[11px]">{teamSummary(team)}</div>
				</div>
				<div class="flex -space-x-1.5">
					{#each team.members.slice(0, 4) as name (name)}
						<div class="ring-card rounded-full ring-2"><UserAvatar {name} size="xs" /></div>
					{/each}
				</div>
				<ChevronRightIcon class="text-subtle-foreground size-4 shrink-0" />
			</a>
		{:else}
			<p class="text-subtle-foreground m-0 px-4 py-10 text-center text-[13px]">
				No teams yet. Create one to group people with their schedules, policies, and services.
			</p>
		{/each}
	</div>
</div>

<TeamEditDialog bind:open={editOpen} isNew action="?/create" {roster} />
