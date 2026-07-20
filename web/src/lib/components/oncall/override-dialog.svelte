<script lang="ts">
	import RepeatIcon from '@lucide/svelte/icons/repeat';
	import { untrack } from 'svelte';
	import { enhance } from '$app/forms';
	import { Alert, AlertContent, AlertTitle } from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as RadioGroup from '$lib/components/ui/radio-group';
	import * as Select from '$lib/components/ui/select';
	import { Label } from '$lib/components/ui/label';
	import { formatShift, weekdayName } from '$lib/oncall';

	let {
		open = $bindable(false),
		schedule,
		target,
		people,
		now,
		error
	}: {
		open?: boolean;
		schedule: string;
		target: { startsAt: string; endsAt: string; person: string | null };
		people: string[];
		now: number;
		error?: string;
	} = $props();

	let person = $state(untrack(() => people.find((name) => name !== target.person) ?? people[0] ?? ''));
	let mode = $state('full');
	let reason = $state('');

	const day = (iso: string) => iso.slice(0, 10);
	const clock = (iso: string) => iso.slice(11, 16);

	// Seed once from the loaded shift; a re-render must not clobber edits
	let startDate = $state(untrack(() => day(target.startsAt)));
	let startTime = $state(untrack(() => clock(target.startsAt)));
	let endDate = $state(untrack(() => day(target.endsAt)));
	let endTime = $state(untrack(() => clock(target.endsAt)));

	const conflict = $derived(person === target.person);

	const window = $derived(formatShift(target, now));
	const held = $derived(target.person ? `shift held by ${target.person}` : 'nobody is on call then');
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[480px]">
		<form
			method="POST"
			action="?/override"
			use:enhance={() =>
				async ({ result, update }) => {
					await update({ reset: false });
					if (result.type === 'success') open = false;
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
						<Dialog.Title class="tracking-heading text-xl font-semibold">Add override</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							{weekdayName(target.startsAt)}
							{day(target.startsAt)} · {schedule} · {held}
						</Dialog.Description>
					</div>
				</div>

				<div class="mt-1 flex flex-col gap-4">
					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
							Who takes the shift
						</Field.FieldLabel>
						<Select.Root type="single" name="person" bind:value={person}>
							<Select.Trigger>{person}</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each people as name (name)}
										<Select.Item value={name} label={name}>{name}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</Field.Field>

					{#if conflict}
						<Alert tone="critical">
							<AlertContent>
								<AlertTitle>That does not change anything</AlertTitle>
								{target.person} already holds this shift. Pick someone else, or change the window.
							</AlertContent>
						</Alert>
					{:else if error}
						<Alert tone="critical">
							<AlertContent>
								<AlertTitle>The override was not saved</AlertTitle>
								{error}
							</AlertContent>
						</Alert>
					{/if}

					<RadioGroup.Root bind:value={mode} name="mode" class="gap-2.5">
						<div class="flex items-start gap-2.5">
							<RadioGroup.Item value="full" id="mode-full" class="mt-0.5" />
							<Label for="mode-full" class="flex flex-col items-start gap-0.5 font-normal">
								<span class="text-foreground text-sm leading-[1.3]">Full shift</span>
								<span class="text-subtle-foreground text-[13px] leading-[1.4]">{window}</span>
							</Label>
						</div>
						<div class="flex items-start gap-2.5">
							<RadioGroup.Item value="partial" id="mode-partial" class="mt-0.5" />
							<Label for="mode-partial" class="flex flex-col items-start gap-0.5 font-normal">
								<span class="text-foreground text-sm leading-[1.3]">Part of the shift</span>
								<span class="text-subtle-foreground text-[13px] leading-[1.4]">
									Pick a start and end
								</span>
							</Label>
						</div>
					</RadioGroup.Root>

					{#if mode === 'partial'}
						<div class="flex flex-col gap-2.5">
							<div class="flex gap-2.5">
								<Field.Field class="flex-1 gap-1.5 space-y-0">
									<Field.FieldLabel
										for="start-date"
										class="text-muted-foreground text-[13px] font-medium"
									>
										Start date
									</Field.FieldLabel>
									<Input id="start-date" name="startDate" type="date" bind:value={startDate} />
								</Field.Field>
								<Field.Field class="w-[120px] gap-1.5 space-y-0">
									<Field.FieldLabel
										for="start-time"
										class="text-muted-foreground text-[13px] font-medium"
									>
										Start time
									</Field.FieldLabel>
									<Input id="start-time" name="startTime" type="time" bind:value={startTime} />
								</Field.Field>
							</div>

							<div class="flex gap-2.5">
								<Field.Field class="flex-1 gap-1.5 space-y-0">
									<Field.FieldLabel
										for="end-date"
										class="text-muted-foreground text-[13px] font-medium"
									>
										End date
									</Field.FieldLabel>
									<Input id="end-date" name="endDate" type="date" bind:value={endDate} />
								</Field.Field>
								<Field.Field class="w-[120px] gap-1.5 space-y-0">
									<Field.FieldLabel
										for="end-time"
										class="text-muted-foreground text-[13px] font-medium"
									>
										End time
									</Field.FieldLabel>
									<Input id="end-time" name="endTime" type="time" bind:value={endTime} />
									<Field.FieldDescription class="text-subtle-foreground text-xs">
										UTC, 24-hour.
									</Field.FieldDescription>
								</Field.Field>
							</div>
						</div>
					{/if}

					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel for="reason" class="text-muted-foreground text-[13px] font-medium">
							Reason
						</Field.FieldLabel>
						<Input
							id="reason"
							name="reason"
							bind:value={reason}
							placeholder="Covering the Sunday night"
						/>
						<Field.FieldDescription class="text-subtle-foreground text-xs">
							Optional. Shows on the calendar and in the audit trail.
						</Field.FieldDescription>
					</Field.Field>
				</div>
			</div>

			<input type="hidden" name="targetStart" value={target.startsAt} />
			<input type="hidden" name="targetEnd" value={target.endsAt} />

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (open = false)}>Cancel</Button>
				<Button type="submit" disabled={conflict}>Save override</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
