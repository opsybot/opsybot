<script lang="ts">
	import UsersIcon from '@lucide/svelte/icons/users';
	import { toast } from 'svelte-sonner';
	import { untrack } from 'svelte';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import type { Role } from '$lib/workflows';

	let { open = $bindable(false), role }: { open?: boolean; role: Role | null } = $props();

	let name = $state('');
	let description = $state('');

	// Reseed on open so add starts blank and edit starts from the stored role
	$effect(() => {
		if (!open) return;
		const target = role;
		untrack(() => {
			name = target?.name ?? '';
			description = target?.description ?? '';
		});
	});
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[460px]">
		<form
			method="POST"
			action="?/save"
			use:enhance={() =>
				async ({ result, update }) => {
					await update({ reset: false });
					if (result.type !== 'success') return;
					toast.success(
						role
							? 'Role updated.'
							: 'Role added. It appears in every incident’s assignment menu.'
					);
					open = false;
				}}
		>
			<input type="hidden" name="id" value={role?.id ?? ''} />
			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						<UsersIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">
							{role ? 'Edit role' : 'Add role'}
						</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							Shown with its description at the moment someone is assigned.
						</Dialog.Description>
					</div>
				</div>

				<div class="mt-1 flex flex-col gap-3.5">
					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel for="role-name" class="text-muted-foreground text-[13px] font-medium">
							Name
						</Field.FieldLabel>
						<Input id="role-name" name="name" bind:value={name} placeholder="Comms lead" />
					</Field.Field>

					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel for="role-desc" class="text-muted-foreground text-[13px] font-medium">
							Responsibility
						</Field.FieldLabel>
						<Textarea
							id="role-desc"
							name="description"
							bind:value={description}
							rows={3}
							placeholder="One or two plain sentences. Shown to the person at the moment they're assigned."
						/>
						<Field.FieldDescription class="text-subtle-foreground text-xs">
							Write it for someone joining mid-incident at 3am.
						</Field.FieldDescription>
					</Field.Field>
				</div>
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (open = false)}>Cancel</Button>
				<Button type="submit" disabled={!name.trim() || !description.trim()}>
					{role ? 'Save' : 'Add role'}
				</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
