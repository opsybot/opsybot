<script lang="ts">
	import type { Component } from 'svelte';
	import { untrack } from 'svelte';
	import type { LucideProps } from '@lucide/svelte';
	import ArrowUpRightIcon from '@lucide/svelte/icons/arrow-up-right';
	import CalendarClockIcon from '@lucide/svelte/icons/calendar-clock';
	import ListChecksIcon from '@lucide/svelte/icons/list-checks';
	import UserXIcon from '@lucide/svelte/icons/user-x';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Select from '$lib/components/ui/select';
	import type { Member } from '$lib/admin';

	let { member, onclose }: { member: Member | null; onclose: () => void } = $props();

	const REPLACEMENTS = ['Maya Chen', 'Priya Nair', 'Dev Patel', 'Sana Ito'];
	const REF_ICONS: Record<string, Component<LucideProps>> = {
		'calendar-clock': CalendarClockIcon,
		'arrow-up-right': ArrowUpRightIcon,
		'list-checks': ListChecksIcon
	};

	// Snapshot of the last member so labels stay populated during the close animation
	let current = $state<Member | null>(null);
	let picks = $state<Record<string, string>>({});
	const open = $derived(!!member);

	$effect(() => {
		if (member)
			untrack(() => {
				current = member;
				picks = {};
			});
	});
	let form: HTMLFormElement;

	const refs = $derived(current?.references ?? []);
	const allPicked = $derived(refs.length > 0 && refs.every((ref) => picks[ref.id]));
	const need = $derived(refs.filter((ref) => !picks[ref.id]).length);
	const picksJson = $derived(JSON.stringify(picks));
</script>

<Dialog.Root {open} onOpenChange={(value) => (value ? undefined : onclose())}>
	<Dialog.Content class="sm:max-w-[560px]">
		{#if current}
			<div class="flex flex-col gap-4 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-critical-wash text-critical-ink flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						<UserXIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">Deactivate {current.name}</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							They're referenced in places that must keep working. Pick a replacement for each before
							deactivating.
						</Dialog.Description>
					</div>
				</div>
				<div class="flex flex-col gap-2.5">
					{#each refs as ref (ref.id)}
						{@const Icon = REF_ICONS[ref.icon]}
						<div class="bg-inset flex items-start gap-2.5 rounded-md border px-3 py-2.5">
							{#if Icon}<Icon class="text-subtle-foreground mt-[3px] size-3.5 shrink-0" />{/if}
							<div class="min-w-0 flex-1">
								<div class="text-[13px] font-medium">{ref.label}</div>
								<div class="text-subtle-foreground mt-px text-[11.5px]">{ref.detail}</div>
							</div>
							<Select.Root
								type="single"
								value={picks[ref.id] ?? ''}
								onValueChange={(value) => (picks = { ...picks, [ref.id]: value })}
							>
								<Select.Trigger size="sm" class="w-[160px]" aria-label="Replacement for {ref.label}">
									{picks[ref.id] || 'Replace with…'}
								</Select.Trigger>
								<Select.Content>
									<Select.Group>
										{#each REPLACEMENTS as name (name)}
											<Select.Item value={name} label={name}>{name}</Select.Item>
										{/each}
									</Select.Group>
								</Select.Content>
							</Select.Root>
						</div>
					{/each}
				</div>
			</div>
			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button variant="ghost" onclick={onclose}>Cancel</Button>
				<Button variant="destructive" disabled={!allPicked} onclick={() => form.requestSubmit()}>
					{allPicked ? 'Reassign all & deactivate' : `${need} replacements needed`}
				</Button>
			</div>
		{/if}
	</Dialog.Content>
</Dialog.Root>

<form
	bind:this={form}
	method="POST"
	action="?/deactivate"
	class="hidden"
	use:enhance={() => async ({ result, update }) => {
		if (result.type === 'failure') {
			toast.error(String(result.data?.error ?? 'Could not deactivate.'));
			return;
		}
		if (result.type !== 'success') return;
		const name = current?.name;
		await update({ reset: false });
		onclose();
		if (name) toast.success(`${name} deactivated. All references reassigned — the audit log records each change.`);
	}}
>
	<input type="hidden" name="id" value={current?.id ?? ''} />
	<input type="hidden" name="replacements" value={picksJson} />
</form>
