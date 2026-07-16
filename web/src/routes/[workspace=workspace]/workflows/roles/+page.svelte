<script lang="ts">
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import UsersIcon from '@lucide/svelte/icons/users';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import Page from '$lib/components/layout/page.svelte';
	import RoleDialog from '$lib/components/workflows/role-dialog.svelte';
	import WfTabs from '$lib/components/workflows/wf-tabs.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import type { Role } from '$lib/workflows';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let dialogOpen = $state(false);
	let editing = $state<Role | null>(null);

	function open(role: Role | null) {
		editing = role;
		dialogOpen = true;
	}
</script>

<Page title="Workflows" subtitle="Automate the boring half of incident response">
	<WfTabs current="roles" />

	<div class="mt-3.5 flex max-w-[720px] flex-col gap-3.5">
		<div class="flex items-center">
			<span class="text-subtle-foreground text-[13px]">
				Shown with their description whenever someone is assigned
			</span>
			<div class="flex-1"></div>
			<Button size="sm" onclick={() => open(null)}>
				<PlusIcon data-icon="inline-start" />
				Add role
			</Button>
		</div>

		<div class="bg-card overflow-hidden rounded-xl border">
			{#each data.roles as role (role.id)}
				<div data-role={role.id} class="flex items-start gap-3 border-t px-4 py-[13px] first:border-t-0">
					<span
						class="bg-inset text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-sm border"
					>
						<UsersIcon class="size-[15px]" />
					</span>
					<div class="min-w-0 flex-1">
						<div class="flex items-center gap-2">
							<span class="text-[13.5px] font-semibold">{role.name}</span>
							{#if role.builtin}
								<Badge tone="neutral" size="sm">default</Badge>
							{/if}
						</div>
						<p class="text-muted-foreground mt-1 text-[12.5px] leading-[1.55]">{role.description}</p>
					</div>
					<Button variant="ghost" size="icon-sm" aria-label="Edit role" onclick={() => open(role)}>
						<PencilIcon />
					</Button>
					{#if !role.builtin}
						<form
							method="POST"
							action="?/remove"
							use:enhance={() =>
								async ({ result, update }) => {
									await update({ reset: false });
									if (result.type === 'success') {
										toast(`${role.name} removed. Existing assignments keep their history.`);
									}
								}}
						>
							<input type="hidden" name="id" value={role.id} />
							<Button type="submit" variant="ghost" size="icon-sm" aria-label="Delete role">
								<Trash2Icon />
							</Button>
						</form>
					{/if}
				</div>
			{/each}
		</div>
	</div>

	<RoleDialog bind:open={dialogOpen} role={editing} />
</Page>
