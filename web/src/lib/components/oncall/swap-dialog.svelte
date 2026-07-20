<script lang="ts">
	import RepeatIcon from '@lucide/svelte/icons/repeat';
	import { untrack } from 'svelte';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';

	let {
		open = $bindable(false),
		shift,
		me,
		people
	}: {
		open?: boolean;
		shift: { when: string; schedule: string } | null;
		me: string;
		people: string[];
	} = $props();

	const others = $derived(people.filter((person) => person !== me));

	let person = $state(untrack(() => people.find((name) => name !== me) ?? people[0] ?? ''));
	let message = $state('');
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[440px]">
		<form
			method="POST"
			action="?/swap"
			use:enhance={() =>
				async ({ update }) => {
					await update();
					open = false;
					message = '';
				}}
		>
			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						<RepeatIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">
							Request a swap
						</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							{shift ? `${shift.when} · ${shift.schedule}` : ''}
						</Dialog.Description>
					</div>
				</div>

				<div class="mt-1 flex flex-col gap-4">
					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
							Swap with
						</Field.FieldLabel>
						<Select.Root type="single" name="person" bind:value={person}>
							<Select.Trigger>{person}</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each others as name (name)}
										<Select.Item value={name} label={name}>{name}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</Field.Field>

					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel for="message" class="text-muted-foreground text-[13px] font-medium">
							Message
						</Field.FieldLabel>
						<Input
							id="message"
							name="message"
							bind:value={message}
							placeholder="Family thing on Sunday — can you take the day part?"
						/>
						<Field.FieldDescription class="text-subtle-foreground text-xs">
							Optional. Shown with the request.
						</Field.FieldDescription>
					</Field.Field>
				</div>
			</div>

			<input type="hidden" name="when" value={shift?.when ?? ''} />

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (open = false)}>Cancel</Button>
				<Button type="submit" disabled={!shift}>Send request</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
