<script lang="ts">
	import Building2Icon from '@lucide/svelte/icons/building-2';
	import CheckIcon from '@lucide/svelte/icons/check';
	import ChevronsUpDownIcon from '@lucide/svelte/icons/chevrons-up-down';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import { goto, invalidateAll } from '$app/navigation';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import {
		healthColour,
		WORKSPACE_COOKIE,
		WORKSPACE_COOKIE_MAX_AGE,
		type Session,
		type Workspace
	} from '$lib/session';

	let { session }: { session: Session } = $props();

	async function select(workspace: Workspace) {
		if (workspace.id === session.activeWorkspace.id) return;
		document.cookie = `${WORKSPACE_COOKIE}=${workspace.id}; path=/; max-age=${WORKSPACE_COOKIE_MAX_AGE}; samesite=lax`;
		await invalidateAll();
	}
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger
		class="bg-card border-input text-muted-foreground hover:text-foreground hover:border-border-strong data-[state=open]:text-foreground data-[state=open]:border-border-strong ml-auto inline-flex items-center gap-1.5 rounded-sm border px-2 py-1 text-[11.5px] transition-colors"
	>
		<span
			class="size-1.5 shrink-0 rounded-full"
			style="background: {healthColour(session.activeWorkspace.health)}"
		></span>
		<span class="max-w-[90px] truncate">{session.activeWorkspace.name}</span>
		<ChevronsUpDownIcon class="size-3 shrink-0" />
	</DropdownMenu.Trigger>

	<DropdownMenu.Content
		align="end"
		alignOffset={-5}
		sideOffset={21}
		class="border-input shadow-pop w-[220px] rounded-md border p-[5px] ring-0"
	>
		<div
			class="text-foreground flex items-center gap-[7px] px-[9px] pt-2 pb-1.5 text-[12.5px] font-semibold"
		>
			<Building2Icon class="text-subtle-foreground size-[13px] shrink-0" />
			<span class="truncate">{session.organization}</span>
			<span class="text-subtle-foreground tracking-label ml-auto text-[9.5px] uppercase">
				organization
			</span>
		</div>

		<DropdownMenu.Group>
			<DropdownMenu.GroupHeading
				class="text-subtle-foreground tracking-label px-[9px] pt-1.5 pb-[3px] text-[9.5px] font-normal uppercase"
			>
				Workspaces
			</DropdownMenu.GroupHeading>
			{#each session.workspaces as workspace (workspace.id)}
				<DropdownMenu.Item
					class="rounded-xs gap-[9px] px-[9px] py-2 text-[13px]"
					onSelect={() => select(workspace)}
				>
					<span
						class="size-[7px] shrink-0 rounded-full"
						style="background: {healthColour(workspace.health)}"
					></span>
					<span class="min-w-0 flex-1 truncate">
						{workspace.name}
						<span class="text-subtle-foreground ml-[7px] text-[10.5px]">
							{workspace.environment}
						</span>
					</span>
					{#if workspace.id === session.activeWorkspace.id}
						<CheckIcon class="text-primary size-3.5 shrink-0" />
					{/if}
				</DropdownMenu.Item>
			{/each}
		</DropdownMenu.Group>

		<DropdownMenu.Separator class="mx-1 my-[5px]" />

		<DropdownMenu.Group>
			<DropdownMenu.Item
				class="text-muted-foreground rounded-xs gap-[9px] px-[9px] py-2 text-[12.5px]"
				onSelect={() => goto('/workspace')}
			>
				<SettingsIcon class="size-3.5 shrink-0" />
				Workspace settings
			</DropdownMenu.Item>
			<DropdownMenu.Item
				class="text-muted-foreground rounded-xs gap-[9px] px-[9px] py-2 text-[12.5px]"
			>
				<Building2Icon class="size-3.5 shrink-0" />
				Manage organization
			</DropdownMenu.Item>
			<DropdownMenu.Item
				class="text-muted-foreground rounded-xs gap-[9px] px-[9px] py-2 text-[12.5px]"
			>
				<PlusIcon class="size-3.5 shrink-0" />
				Create workspace
			</DropdownMenu.Item>
		</DropdownMenu.Group>
	</DropdownMenu.Content>
</DropdownMenu.Root>
