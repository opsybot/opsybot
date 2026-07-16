<script lang="ts">
	import BellOffIcon from '@lucide/svelte/icons/bell-off';
	import XIcon from '@lucide/svelte/icons/x';
	import { untrack } from 'svelte';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { DURATIONS, SCOPE_FIELDS } from '$lib/alerts';
	import { ws } from '$lib/navigation';

	let { open = $bindable(false), source }: { open?: boolean; source: string | null } = $props();

	type Condition = { field: string; value: string };

	// Seed from the initial source only; a prop update must not reset edited rows
	let conditions = $state<Condition[]>(untrack(() => [{ field: 'source', value: source ?? '' }]));
	let start = $state('now');
	let duration = $state('1h');
	let reason = $state('');

	const today = new Date().toISOString().slice(0, 10);
	let date = $state(today);
	let time = $state('22:00');

	const placeholder = (field: string) =>
		field === 'label' ? 'env:staging' : field === 'source' ? 'Datadog' : 'payments-api';
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[520px]">
		<form
			method="POST"
			action="{ws('/alerts/silences')}?/create"
			use:enhance={() => async ({ update }) => {
				await update();
				open = false;
				conditions = [{ field: 'source', value: '' }];
				reason = '';
			}}
		>
			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						<BellOffIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">Create silence</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							Matching alerts are still recorded — they just don't page anyone.
						</Dialog.Description>
					</div>
				</div>

				<div class="mt-1 flex flex-col gap-4">
					<div class="flex flex-col gap-2">
						<div class="text-subtle-foreground tracking-label text-[11px] uppercase">
							Scope — silence alerts matching all conditions
						</div>

						{#each conditions as condition, index (index)}
							<div class="flex items-center gap-2">
								<Select.Root type="single" name="field" bind:value={condition.field}>
									<Select.Trigger size="sm" class="w-[110px]">{condition.field}</Select.Trigger>
									<Select.Content>
										<Select.Group>
											{#each SCOPE_FIELDS as field (field)}
												<Select.Item value={field} label={field}>{field}</Select.Item>
											{/each}
										</Select.Group>
									</Select.Content>
								</Select.Root>

								<span class="text-subtle-foreground font-mono text-xs">=</span>

								<Input
									name="value"
									bind:value={condition.value}
									placeholder={placeholder(condition.field)}
									class="h-[34px] flex-1 text-[13px]"
								/>

								{#if conditions.length > 1}
									<Button
										type="button"
										variant="ghost"
										size="icon-sm"
										aria-label="Remove condition"
										onclick={() => (conditions = conditions.filter((_, i) => i !== index))}
									>
										<XIcon />
									</Button>
								{/if}
							</div>
						{/each}

						<button
							type="button"
							onclick={() => (conditions = [...conditions, { field: 'service', value: '' }])}
							class="text-muted-foreground hover:text-brand-foreground self-start text-[12.5px]"
						>
							+ Add condition
						</button>
					</div>

					<div class="flex gap-2.5">
						<Field.Field class="flex-1 gap-1.5 space-y-0">
							<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
								Start
							</Field.FieldLabel>
							<Select.Root type="single" name="start" bind:value={start}>
								<Select.Trigger>{start === 'now' ? 'Now' : 'At a set time…'}</Select.Trigger>
								<Select.Content>
									<Select.Group>
										<Select.Item value="now" label="Now">Now</Select.Item>
										<Select.Item value="later" label="At a set time…">At a set time…</Select.Item>
									</Select.Group>
								</Select.Content>
							</Select.Root>
						</Field.Field>

						<Field.Field class="flex-1 gap-1.5 space-y-0">
							<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
								Duration
							</Field.FieldLabel>
							<Select.Root type="single" name="duration" bind:value={duration}>
								<Select.Trigger>
									{DURATIONS.find((entry) => entry.value === duration)?.label}
								</Select.Trigger>
								<Select.Content>
									<Select.Group>
										{#each DURATIONS as entry (entry.value)}
											<Select.Item value={entry.value} label={entry.label}>{entry.label}</Select.Item>
										{/each}
									</Select.Group>
								</Select.Content>
							</Select.Root>
						</Field.Field>
					</div>

					{#if start === 'later'}
						<div class="flex gap-2.5">
							<Field.Field class="flex-1 gap-1.5 space-y-0">
								<Field.FieldLabel for="date" class="text-muted-foreground text-[13px] font-medium">
									Start date
								</Field.FieldLabel>
								<Input id="date" name="date" type="date" bind:value={date} />
							</Field.Field>

							<Field.Field class="w-[130px] gap-1.5 space-y-0">
								<Field.FieldLabel for="time" class="text-muted-foreground text-[13px] font-medium">
									Start time
								</Field.FieldLabel>
								<Input id="time" name="time" type="time" bind:value={time} />
								<Field.FieldDescription class="text-subtle-foreground text-xs">
									UTC, 24-hour.
								</Field.FieldDescription>
							</Field.Field>
						</div>
					{/if}

					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel for="reason" class="text-muted-foreground text-[13px] font-medium">
							Reason
						</Field.FieldLabel>
						<Input id="reason" name="reason" bind:value={reason} placeholder="Planned failover test" />
						<Field.FieldDescription class="text-subtle-foreground text-xs">
							Shown wherever the silence mutes something.
						</Field.FieldDescription>
					</Field.Field>
				</div>
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (open = false)}>Cancel</Button>
				<Button type="submit">{start === 'now' ? 'Start silence' : 'Schedule silence'}</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
