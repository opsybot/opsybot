<script lang="ts">
	import PlusIcon from '@lucide/svelte/icons/plus';
	import UserXIcon from '@lucide/svelte/icons/user-x';
	import { enhance } from '$app/forms';
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Table from '$lib/components/ui/table';
	import Tag from '$lib/components/tag.svelte';
	import ConfirmDialog from '$lib/components/statuspages/confirm-dialog.svelte';
	import DeactivateDialog from '$lib/components/admin/deactivate-dialog.svelte';
	import InviteDialog from '$lib/components/admin/invite-dialog.svelte';
	import RoleSelect from '$lib/components/admin/role-select.svelte';
	import { twoFactorBadge, type Member } from '$lib/admin';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let inviteOpen = $state(false);
	let deactivating = $state<Member | null>(null);
	let confirmMember = $state<Member | null>(null);
	let confirmOpen = $state(false);

	function startDeactivate(member: Member) {
		if (member.references.length) {
			deactivating = member;
		} else {
			confirmMember = member;
			confirmOpen = true;
		}
	}
</script>

<div class="flex flex-col gap-3.5">
	<div class="flex items-center">
		<span class="text-subtle-foreground text-[13px]">{data.members.length} members</span>
		<div class="flex-1"></div>
		<Button size="sm" onclick={() => (inviteOpen = true)}>
			<PlusIcon data-icon="inline-start" />
			Invite member
		</Button>
	</div>

	<section class="bg-card overflow-hidden rounded-xl border">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head class="pl-[18px]">Member</Table.Head>
					<Table.Head class="w-[150px]">Role</Table.Head>
					<Table.Head class="w-[80px]">2FA</Table.Head>
					<Table.Head class="w-[130px]">Auth</Table.Head>
					<Table.Head class="w-[130px]">Last active</Table.Head>
					<Table.Head class="w-[130px] pr-[18px]"></Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each data.members as member (member.id)}
					{@const badge = twoFactorBadge(member.twoFactor)}
					<Table.Row
						data-member={member.id}
						data-deactivated={member.deactivated ? 'true' : 'false'}
						style={member.deactivated ? 'opacity:.5' : undefined}
					>
						<Table.Cell class="py-2.5 pl-[18px]">
							<div class="flex items-center gap-2.5">
								<UserAvatar name={member.name} size="sm" />
								<div class="min-w-0">
									<div class="text-foreground text-[13px] font-medium">
										{member.name}{member.deactivated ? ' · deactivated' : ''}
									</div>
									<div class="text-subtle-foreground truncate font-mono text-[11px]">{member.email}</div>
								</div>
							</div>
						</Table.Cell>
						<Table.Cell>
							{#if member.deactivated}
								<Tag>{member.role}</Tag>
							{:else}
								<RoleSelect id={member.id} value={member.role} />
							{/if}
						</Table.Cell>
						<Table.Cell><Badge tone={badge.tone} size="sm">{badge.label}</Badge></Table.Cell>
						<Table.Cell class="font-mono text-[12px]">{member.auth}</Table.Cell>
						<Table.Cell class="text-subtle-foreground font-mono text-[12px]">{member.lastActive}</Table.Cell>
						<Table.Cell class="pr-[18px] text-right">
							{#if member.deactivated}
								<form
									method="POST"
									action="?/reactivate"
									class="inline"
									use:enhance={() => async ({ update }) => update({ reset: false })}
								>
									<input type="hidden" name="id" value={member.id} />
									<Button type="submit" size="sm" variant="ghost">Reactivate</Button>
								</form>
							{:else}
								<Button size="sm" variant="ghost" onclick={() => startDeactivate(member)}>
									<UserXIcon data-icon="inline-start" />
									Deactivate
								</Button>
							{/if}
						</Table.Cell>
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</section>
</div>

<InviteDialog bind:open={inviteOpen} />
<DeactivateDialog member={deactivating} onclose={() => (deactivating = null)} />
<ConfirmDialog
	bind:open={confirmOpen}
	tone="critical"
	title={confirmMember ? `Deactivate ${confirmMember.name}?` : ''}
	action="?/deactivate"
	confirmLabel="Deactivate"
>
	<input type="hidden" name="id" value={confirmMember?.id ?? ''} />
	<input type="hidden" name="replacements" value={'{}'} />
	Nothing references {confirmMember?.name}, so this is safe. They lose access immediately — reactivate any
	time.
</ConfirmDialog>
