<script lang="ts">
	import { tick, untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import * as Select from '$lib/components/ui/select';
	import { ROLES, type Role } from '$lib/admin';

	let { id, value }: { id: string; value: Role } = $props();

	let role = $state(untrack(() => value));
	$effect(() => {
		role = value;
	});
	let form: HTMLFormElement;

	async function change(next: string) {
		role = next as Role;
		await tick();
		form.requestSubmit();
	}
</script>

<form
	method="POST"
	action="?/changeRole"
	bind:this={form}
	use:enhance={() => async ({ result, update }) => {
		await update({ reset: false });
		if (result.type !== 'success') {
			role = value;
			toast.error(String((result.type === 'failure' && result.data?.error) || 'Could not change the role.'));
			return;
		}
		toast.success('Role updated.');
	}}
>
	<input type="hidden" name="id" value={id} />
	<input type="hidden" name="role" value={role} />
	<Select.Root type="single" value={role} onValueChange={change}>
		<Select.Trigger size="sm" class="w-[130px]" aria-label="Role">{role}</Select.Trigger>
		<Select.Content>
			<Select.Group>
				{#each ROLES as option (option)}
					<Select.Item value={option} label={option}>{option}</Select.Item>
				{/each}
			</Select.Group>
		</Select.Content>
	</Select.Root>
</form>
