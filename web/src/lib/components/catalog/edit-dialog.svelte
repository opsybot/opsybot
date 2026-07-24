<script lang="ts">
	import BoxesIcon from '@lucide/svelte/icons/boxes';
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Alert, AlertContent, AlertTitle } from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { Textarea } from '$lib/components/ui/textarea';
	import type { Service } from '$lib/catalog';

	let {
		open = $bindable(false),
		service,
		teams = [],
		error
	}: {
		open?: boolean;
		service: Service | null;
		teams?: { slug: string; name: string }[];
		error?: string;
	} = $props();

	const editing = $derived(service?.id ?? null);

	let name = $state('');
	let team = $state('');
	let description = $state('');

	const teamName = $derived(teams.find((entry) => entry.slug === team)?.name ?? 'No team');

	let seededFor = $state<string | null | undefined>(undefined);
	$effect(() => {
		const subjectId = service?.id ?? (open ? '\0new' : null);
		if (!open || subjectId === seededFor) return;

		untrack(() => {
			seededFor = subjectId;
			name = service?.id ?? '';
			team = service?.team ?? '';
			description = service?.description ?? '';
		});
	});

	function close() {
		open = false;
		seededFor = undefined;
		if (page.url.searchParams.has('new') || page.url.searchParams.has('edit')) {
			const url = new URL(page.url);
			url.searchParams.delete('new');
			url.searchParams.delete('edit');
			goto(url, { replaceState: true, noScroll: true, keepFocus: true });
		}
	}
</script>

<Dialog.Root
	bind:open
	onOpenChangeComplete={(next) => {
		if (!next) close();
	}}
>
	<Dialog.Content class="sm:max-w-[520px]">
		<form
			method="POST"
			action="?/save"
			use:enhance={() =>
				async ({ result, update }) => {
					if (result.type === 'success' || result.type === 'redirect') {
						toast.success(editing ? `${name} saved.` : `${name} created.`);
						open = false;
					}
					if (result.type === 'redirect') await update();
					else {
						await update({ reset: false });
						if (result.type === 'success') close();
					}
				}}
		>
			{#if editing}
				<input type="hidden" name="editing" value={editing} />
			{/if}

			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						<BoxesIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">
							{editing ? `Edit ${editing}` : 'New service'}
						</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							A service ties its alerts and incidents to an owning team.
						</Dialog.Description>
					</div>
				</div>

				<div class="mt-1 flex flex-col gap-3.5">
					{#if error}
						<Alert tone="critical">
							<AlertContent>
								<AlertTitle>The service was not saved</AlertTitle>
								{error}
							</AlertContent>
						</Alert>
					{/if}

					<div class="flex gap-2.5">
						<Field.Field class="flex-1 gap-1.5 space-y-0">
							<Field.FieldLabel for="name" class="text-muted-foreground text-[13px] font-medium">
								Name
							</Field.FieldLabel>
							<Input id="name" name="name" bind:value={name} placeholder="Payments API" />
						</Field.Field>

						<Field.Field class="w-[160px] gap-1.5 space-y-0">
							<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
								Owning team
							</Field.FieldLabel>
							<Select.Root type="single" name="team" bind:value={team}>
								<Select.Trigger>{teamName}</Select.Trigger>
								<Select.Content>
									<Select.Group>
										<Select.Item value="" label="No team">No team</Select.Item>
										{#each teams as option (option.slug)}
											<Select.Item value={option.slug} label={option.name}>{option.name}</Select.Item>
										{/each}
									</Select.Group>
								</Select.Content>
							</Select.Root>
						</Field.Field>
					</div>

					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel for="description" class="text-muted-foreground text-[13px] font-medium">
							Description
						</Field.FieldLabel>
						<Textarea
							id="description"
							name="description"
							bind:value={description}
							rows={2}
							placeholder="One sentence: what breaks when this breaks?"
						/>
					</Field.Field>
				</div>
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={close}>Cancel</Button>
				<Button type="submit" disabled={!name.trim()}>
					{editing ? 'Save' : 'Create service'}
				</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
