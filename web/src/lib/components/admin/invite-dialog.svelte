<script lang="ts">
	import { untrack } from 'svelte';
	import MailIcon from '@lucide/svelte/icons/mail';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { ROLES } from '$lib/admin';

	let { open = $bindable(false) }: { open?: boolean } = $props();

	let email = $state('');
	let role = $state('responder');
	let form: HTMLFormElement;

	$effect(() => {
		if (!open) return;
		untrack(() => {
			email = '';
			role = 'responder';
		});
	});

	const valid = $derived(email.includes('@'));
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[440px]">
		<div class="flex flex-col gap-4 p-6">
			<div class="flex items-start gap-3">
				<span
					class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
				>
					<MailIcon class="size-5" />
				</span>
				<div class="flex flex-1 flex-col gap-1">
					<Dialog.Title class="tracking-heading text-xl font-semibold">Invite member</Dialog.Title>
				</div>
			</div>
			<div class="flex flex-col gap-3.5">
				<Field.Field class="gap-1.5 space-y-0">
					<Field.FieldLabel for="invite-email" class="text-muted-foreground text-[13px] font-medium">Email</Field.FieldLabel>
					<Input id="invite-email" type="email" bind:value={email} placeholder="new-hire@acme.dev" />
				</Field.Field>
				<Field.Field class="gap-1.5 space-y-0">
					<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">Role</Field.FieldLabel>
					<Select.Root type="single" value={role} onValueChange={(value) => (role = value)}>
						<Select.Trigger class="w-full" aria-label="Role">{role}</Select.Trigger>
						<Select.Content>
							<Select.Group>
								{#each ROLES as option (option)}
									<Select.Item value={option} label={option}>{option}</Select.Item>
								{/each}
							</Select.Group>
						</Select.Content>
					</Select.Root>
					<Field.FieldDescription class="text-subtle-foreground text-xs">
						Admins manage the workspace; responders run incidents; viewers read.
					</Field.FieldDescription>
				</Field.Field>
			</div>
		</div>
		<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
			<Button variant="ghost" onclick={() => (open = false)}>Cancel</Button>
			<Button disabled={!valid} onclick={() => form.requestSubmit()}>Send invite</Button>
		</div>
	</Dialog.Content>
</Dialog.Root>

<form
	bind:this={form}
	method="POST"
	action="?/invite"
	class="hidden"
	use:enhance={() => async ({ result, update }) => {
		if (result.type === 'failure') {
			toast.error(String(result.data?.error ?? 'Could not send the invite.'));
			return;
		}
		if (result.type !== 'success') return;
		const sent = email;
		await update({ reset: false });
		open = false;
		toast.success(`Invite sent to ${sent}. Valid 7 days.`);
	}}
>
	<input type="hidden" name="email" value={email} />
	<input type="hidden" name="role" value={role} />
</form>
