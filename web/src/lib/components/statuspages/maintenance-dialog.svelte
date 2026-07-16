<script lang="ts">
	import WrenchIcon from '@lucide/svelte/icons/wrench';
	import { toast } from 'svelte-sonner';
	import { untrack } from 'svelte';
	import { enhance } from '$app/forms';
	import Tag from '$lib/components/tag.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { Textarea } from '$lib/components/ui/textarea';

	let {
		open = $bindable(false),
		components,
		notices
	}: {
		open?: boolean;
		components: string[];
		notices: string[];
	} = $props();

	const today = new Date().toISOString().slice(0, 10);

	let title = $state('');
	let picked = $state(new Set<string>());
	let date = $state(today);
	let startTime = $state('22:00');
	let endTime = $state('23:00');
	let notice = $state(untrack(() => notices[2] ?? notices[0]));
	let description = $state('');

	function toggle(component: string) {
		const next = new Set(picked);
		next.has(component) ? next.delete(component) : next.add(component);
		picked = next;
	}

	function reset() {
		title = '';
		picked = new Set();
		description = '';
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[540px]">
		<form
			method="POST"
			action="?/schedule"
			use:enhance={() =>
				async ({ result, update }) => {
					await update({ reset: false });
					if (result.type === 'success') {
						toast.success(`Maintenance scheduled. Subscribers get notified ${notice}.`);
						open = false;
						reset();
					}
				}}
		>
			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						<WrenchIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">
							Schedule maintenance
						</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							Published to the page and announced to subscribers ahead of time.
						</Dialog.Description>
					</div>
				</div>

				<div class="mt-1 flex flex-col gap-3.5">
					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel for="title" class="text-muted-foreground text-[13px] font-medium">
							Title
						</Field.FieldLabel>
						<Input
							id="title"
							name="title"
							bind:value={title}
							placeholder="Database maintenance — primary failover test"
						/>
					</Field.Field>

					<div>
						<div class="text-subtle-foreground tracking-label mb-2 text-[11px] uppercase">
							Affected components
						</div>
						<div class="flex flex-wrap gap-1.5">
							{#each components as component (component)}
								<Tag selected={picked.has(component)} onclick={() => toggle(component)}>
									{component}
								</Tag>
							{/each}
						</div>
						{#each [...picked] as component (component)}
							<input type="hidden" name="component" value={component} />
						{/each}
					</div>

					<div class="flex flex-wrap gap-2.5">
						<Field.Field class="min-w-[160px] flex-1 gap-1.5 space-y-0">
							<Field.FieldLabel for="date" class="text-muted-foreground text-[13px] font-medium">
								Date
							</Field.FieldLabel>
							<Input id="date" name="date" type="date" bind:value={date} />
						</Field.Field>
						<Field.Field class="w-[110px] gap-1.5 space-y-0">
							<Field.FieldLabel for="start" class="text-muted-foreground text-[13px] font-medium">
								Start
							</Field.FieldLabel>
							<Input id="start" name="startTime" type="time" bind:value={startTime} />
						</Field.Field>
						<Field.Field class="w-[110px] gap-1.5 space-y-0">
							<Field.FieldLabel for="end" class="text-muted-foreground text-[13px] font-medium">
								End
							</Field.FieldLabel>
							<Input id="end" name="endTime" type="time" bind:value={endTime} />
							<Field.FieldDescription class="text-subtle-foreground text-xs">UTC.</Field.FieldDescription>
						</Field.Field>
					</div>

					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel
							for="description"
							class="text-muted-foreground text-[13px] font-medium"
						>
							Description
						</Field.FieldLabel>
						<Textarea
							id="description"
							name="description"
							bind:value={description}
							rows={2}
							placeholder="What subscribers should expect: brief interruptions to checkout while the primary fails over."
						/>
					</Field.Field>

					<Field.Field class="max-w-[240px] gap-1.5 space-y-0">
						<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
							Subscriber notice
						</Field.FieldLabel>
						<Select.Root type="single" name="notice" bind:value={notice}>
							<Select.Trigger>{notice}</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each notices as option (option)}
										<Select.Item value={option} label={option}>{option}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</Field.Field>
				</div>
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (open = false)}>Cancel</Button>
				<Button type="submit" disabled={!title.trim() || picked.size === 0}>Schedule</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
