<script lang="ts">
	import BoxesIcon from '@lucide/svelte/icons/boxes';
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import Tag from '$lib/components/tag.svelte';
	import { Alert, AlertContent, AlertTitle } from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { Textarea } from '$lib/components/ui/textarea';
	import { CATALOG_TEAMS, LINK_KINDS, type Service } from '$lib/catalog';

	let {
		open = $bindable(false),
		service,
		names,
		error
	}: {
		open?: boolean;
		service: Service | null;
		names: string[];
		error?: string;
	} = $props();

	const editing = $derived(service?.id ?? null);

	let name = $state('');
	let team = $state('platform');
	let description = $state('');
	let links = $state<Record<string, string>>({ runbook: '', dashboard: '', repository: '' });
	let deps = $state<Set<string>>(new Set());

	let seededFor = $state<string | null | undefined>(undefined);
	$effect(() => {
		const subjectId = service?.id ?? (open ? '\0new' : null);
		if (!open || subjectId === seededFor) return;

		untrack(() => {
			seededFor = subjectId;
			name = service?.id ?? '';
			team = service?.team ?? 'platform';
			description = service?.description ?? '';
			links = { ...(service?.links ?? { runbook: '', dashboard: '', repository: '' }) };
			deps = new Set(service?.deps ?? []);
		});
	});

	function toggleDep(id: string) {
		const next = new Set(deps);
		next.has(id) ? next.delete(id) : next.add(id);
		deps = next;
	}

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

	const pickable = $derived(names.filter((id) => id !== editing));
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
						toast.success(
							editing
								? `${name} saved. Alerts and incidents on it show up here.`
								: `${name} created. Alerts where service = ${name} now attach here.`
						);
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
							A service ties its alerts, incidents, owner and runbook together.
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
							<Input
								id="name"
								name="name"
								bind:value={name}
								placeholder="payments-api"
								class="font-mono"
							/>
						</Field.Field>

						<Field.Field class="w-[160px] gap-1.5 space-y-0">
							<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
								Owning team
							</Field.FieldLabel>
							<Select.Root type="single" name="team" bind:value={team}>
								<Select.Trigger>{team}</Select.Trigger>
								<Select.Content>
									<Select.Group>
										{#each CATALOG_TEAMS as option (option)}
											<Select.Item value={option} label={option}>{option}</Select.Item>
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

					<div class="flex flex-col gap-2">
						{#each LINK_KINDS as { kind, label, placeholder } (kind)}
							<Field.Field class="gap-1.5 space-y-0">
								<Field.FieldLabel for="link-{kind}" class="text-muted-foreground text-[13px] font-medium">
									{label}
								</Field.FieldLabel>
								<Input
									id="link-{kind}"
									name={kind}
									bind:value={links[kind]}
									{placeholder}
									class="h-[34px] font-mono text-xs"
								/>
							</Field.Field>
						{/each}
					</div>

					<div>
						<div class="text-subtle-foreground tracking-label mb-2 text-[11px] uppercase">
							Depends on
						</div>
						<div class="flex flex-wrap gap-1.5">
							{#each pickable as id (id)}
								<Tag selected={deps.has(id)} onclick={() => toggleDep(id)}>{id}</Tag>
							{/each}
						</div>
						{#each [...deps] as dep (dep)}
							<input type="hidden" name="dep" value={dep} />
						{/each}
					</div>
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
