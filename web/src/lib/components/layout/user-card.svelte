<script lang="ts">
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import UserRoundIcon from '@lucide/svelte/icons/user-round';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import type { SessionUser } from '$lib/session';
	import UserAvatar from './user-avatar.svelte';

	let { user }: { user: SessionUser } = $props();
	let logoutForm: HTMLFormElement;
</script>

<div class="bg-card flex items-center gap-2.5 rounded-md border px-2.5 py-2">
	<UserAvatar name={user.name} onCall={!!user.onCallFor} presence="online" />

	<div class="min-w-0 flex-1">
		<div class="text-foreground truncate text-[13px] font-medium">{user.name}</div>
		<div class="truncate text-[11px] {user.onCallFor ? 'text-primary' : 'text-subtle-foreground'}">
			{user.onCallFor ? `On call · ${user.onCallFor}` : 'Off call'}
		</div>
	</div>

	<DropdownMenu.Root>
		<DropdownMenu.Trigger>
			{#snippet child({ props })}
				<Button {...props} variant="ghost" size="icon-sm" aria-label="Account menu">
					<SettingsIcon />
				</Button>
			{/snippet}
		</DropdownMenu.Trigger>
		<DropdownMenu.Content
			align="end"
			sideOffset={8}
			class="border-input shadow-pop w-[180px] rounded-md border p-[5px] ring-0"
		>
			<DropdownMenu.Item
				class="text-muted-foreground rounded-xs gap-[9px] px-[9px] py-2 text-[12.5px]"
				onSelect={() => goto('/account')}
			>
				<UserRoundIcon class="size-3.5 shrink-0" />
				Account settings
			</DropdownMenu.Item>
			<DropdownMenu.Separator class="mx-1 my-[5px]" />
			<DropdownMenu.Item
				class="text-muted-foreground rounded-xs gap-[9px] px-[9px] py-2 text-[12.5px]"
				onSelect={() => logoutForm.requestSubmit()}
			>
				<LogOutIcon class="size-3.5 shrink-0" />
				Log out
			</DropdownMenu.Item>
		</DropdownMenu.Content>
	</DropdownMenu.Root>

	<form bind:this={logoutForm} method="POST" action="/logout" class="hidden"></form>
</div>
