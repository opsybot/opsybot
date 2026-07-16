<script lang="ts">
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import { Button } from '$lib/components/ui/button';
	import type { SessionUser } from '$lib/session';
	import UserAvatar from './user-avatar.svelte';

	let { user }: { user: SessionUser } = $props();
</script>

<div class="bg-card flex items-center gap-2.5 rounded-md border px-2.5 py-2">
	<UserAvatar name={user.name} onCall={!!user.onCallFor} presence="online" />

	<div class="min-w-0 flex-1">
		<div class="text-foreground truncate text-[13px] font-medium">{user.name}</div>
		<div class="truncate text-[11px] {user.onCallFor ? 'text-primary' : 'text-subtle-foreground'}">
			{user.onCallFor ? `On call · ${user.onCallFor}` : 'Off call'}
		</div>
	</div>

	<Button variant="ghost" size="icon-sm" aria-label="Profile settings">
		<SettingsIcon />
	</Button>
</div>
