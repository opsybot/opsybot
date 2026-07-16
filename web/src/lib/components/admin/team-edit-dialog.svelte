<script lang="ts">
	import { untrack } from 'svelte';
	import UsersIcon from '@lucide/svelte/icons/users';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import Tag from '$lib/components/tag.svelte';
	import type { Member, Team } from '$lib/admin';

	let {
		open = $bindable(false),
		isNew = false,
		action,
		team = null,
		roster
	}: { open?: boolean; isNew?: boolean; action: string; team?: Team | null; roster: Member[] } = $props();

	let name = $state('');
	let picked = $state<string[]>([]);
	let submitting = $state(false);
	let form: HTMLFormElement;

	$effect(() => {
		if (!open) return;
		untrack(() => {
			name = team?.name ?? '';
			picked = [...(team?.members ?? [])];
			submitting = false;
		});
	});

	function toggle(member: string) {
		picked = picked.includes(member) ? picked.filter((entry) => entry !== member) : [...picked, member];
	}
	const membersJson = $derived(JSON.stringify(picked));
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[460px]">
		<div class="flex flex-col gap-4 p-6">
			<div class="flex items-start gap-3">
				<span
					class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
				>
					<UsersIcon class="size-5" />
				</span>
				<div class="flex flex-1 flex-col gap-1">
					<Dialog.Title class="tracking-heading text-xl font-semibold">{isNew ? 'New team' : 'Edit team'}</Dialog.Title>
				</div>
			</div>
			<div class="flex flex-col gap-3.5">
				<Field.Field class="gap-1.5 space-y-0">
					<Field.FieldLabel for="team-name" class="text-muted-foreground text-[13px] font-medium">Name</Field.FieldLabel>
					<Input id="team-name" class="font-mono" bind:value={name} placeholder="security" />
				</Field.Field>
				<div>
					<div class="text-subtle-foreground mb-2 text-[11px] tracking-[0.08em] uppercase">Members</div>
					<div class="flex flex-wrap gap-1.5">
						{#each roster as member (member.id)}
							<Tag selected={picked.includes(member.name)} onclick={() => toggle(member.name)}>{member.name}</Tag>
						{/each}
					</div>
				</div>
			</div>
		</div>
		<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
			<Button variant="ghost" onclick={() => (open = false)}>Cancel</Button>
			<Button
				disabled={!name.trim() || submitting}
				onclick={() => {
					submitting = true;
					form.requestSubmit();
				}}
			>
				{isNew ? 'Create team' : 'Save'}
			</Button>
		</div>
	</Dialog.Content>
</Dialog.Root>

<form
	bind:this={form}
	method="POST"
	{action}
	class="hidden"
	use:enhance={() => async ({ result, update }) => {
		if (result.type === 'failure') {
			toast.error(String(result.data?.error ?? 'Could not save the team.'));
			return;
		}
		if (result.type !== 'success') return;
		const count = picked.length;
		await update({ reset: false });
		open = false;
		toast.success(`${isNew ? 'Team created.' : 'Team saved.'} ${count} members.`);
	}}
>
	<input type="hidden" name="name" value={name} />
	<input type="hidden" name="members" value={membersJson} />
</form>
