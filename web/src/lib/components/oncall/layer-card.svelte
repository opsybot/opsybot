<script lang="ts">
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import ChevronUpIcon from '@lucide/svelte/icons/chevron-up';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import XIcon from '@lucide/svelte/icons/x';
	import { Button } from '$lib/components/ui/button';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as RadioGroup from '$lib/components/ui/radio-group';
	import * as Select from '$lib/components/ui/select';
	import { formatDuty, layerName, ROTATIONS, type Layer } from '$lib/oncall';

	let {
		layer,
		index,
		total,
		people,
		errors,
		update,
		move,
		remove
	}: {
		layer: Layer;
		index: number;
		total: number;
		people: string[];
		errors?: { participants?: string[]; intervalDays?: string[]; startsOn?: string[] };
		update: (index: number, patch: Partial<Layer>) => void;
		move: (index: number, by: number) => void;
		remove: (index: number) => void;
	} = $props();

	const addable = $derived(people.filter((person) => !layer.participants.includes(person)));
	const nobody = $derived(layer.participants.length === 0);

	let adding = $state('');

	function add(person: string) {
		if (!person) return;
		update(index, { participants: [...layer.participants, person] });
		adding = '';
	}

	function swap(from: number, to: number) {
		const people = [...layer.participants];
		[people[from], people[to]] = [people[to], people[from]];
		update(index, { participants: people });
	}

	const hour = (value: number) => `${String(value).padStart(2, '0')}:00`;
	const HOURS = Array.from({ length: 24 }, (_, index) => index);
</script>

<div class="bg-card overflow-hidden rounded-xl border {nobody ? 'border-critical-edge' : ''}">
	<header class="bg-inset flex items-center gap-2.5 border-b px-3.5 py-2.5">
		<span
			class="bg-brand-wash border-brand-edge text-brand-foreground flex size-5 shrink-0 items-center justify-center rounded-full border font-mono text-[11px] font-semibold"
		>
			{total - index}
		</span>
		<span class="text-[13px] font-semibold">{layerName(total, index)}</span>
		<span class="text-subtle-foreground text-[11px]">
			{index === 0 ? 'highest precedence' : 'overridden by layers above'}
		</span>

		<div class="flex-1"></div>

		<Button
			type="button"
			variant="ghost"
			size="icon-sm"
			aria-label="Raise precedence"
			disabled={index === 0}
			onclick={() => move(index, -1)}
		>
			<ChevronUpIcon />
		</Button>
		<Button
			type="button"
			variant="ghost"
			size="icon-sm"
			aria-label="Lower precedence"
			disabled={index === total - 1}
			onclick={() => move(index, 1)}
		>
			<ChevronDownIcon />
		</Button>
		<Button
			type="button"
			variant="ghost"
			size="icon-sm"
			aria-label="Delete layer"
			disabled={total === 1}
			onclick={() => remove(index)}
		>
			<Trash2Icon />
		</Button>
	</header>

	<div class="flex flex-col gap-3.5 px-3.5 py-3">
		<div>
			<div class="text-subtle-foreground tracking-label mb-2 text-[11px] uppercase">
				Participants — rotation order
			</div>

			{#if nobody}
				<div class="text-critical-ink mb-1.5 text-[12.5px]">
					A layer needs at least one participant.
				</div>
			{:else}
				<div class="mb-2 flex flex-col gap-1.5">
					{#each layer.participants as person, position (person)}
						<div class="bg-inset flex items-center gap-2.5 rounded-md border px-2.5 py-1.5">
							<span class="text-subtle-foreground w-3.5 text-right font-mono text-[11px]">
								{position + 1}
							</span>
							<span class="text-[13px] font-medium">{person}</span>

							<div class="flex-1"></div>

							<Button
								type="button"
								variant="ghost"
								size="icon-sm"
								aria-label="Move {person} up"
								disabled={position === 0}
								onclick={() => swap(position, position - 1)}
							>
								<ChevronUpIcon />
							</Button>
							<Button
								type="button"
								variant="ghost"
								size="icon-sm"
								aria-label="Move {person} down"
								disabled={position === layer.participants.length - 1}
								onclick={() => swap(position, position + 1)}
							>
								<ChevronDownIcon />
							</Button>
							<Button
								type="button"
								variant="ghost"
								size="icon-sm"
								aria-label="Remove {person}"
								onclick={() =>
									update(index, {
										participants: layer.participants.filter((name) => name !== person)
									})}
							>
								<XIcon />
							</Button>
						</div>
					{/each}
				</div>
			{/if}

			{#if addable.length}
				<Select.Root type="single" bind:value={adding} onValueChange={add}>
					<Select.Trigger size="sm" class="w-[220px]">Add participant…</Select.Trigger>
					<Select.Content>
						<Select.Group>
							{#each addable as person (person)}
								<Select.Item value={person} label={person}>{person}</Select.Item>
							{/each}
						</Select.Group>
					</Select.Content>
				</Select.Root>
			{/if}
		</div>

		<div class="flex flex-wrap gap-3.5">
			<div class="min-w-[240px]">
				<div class="text-subtle-foreground tracking-label mb-2 text-[11px] uppercase">Rotation</div>

				<RadioGroup.Root
					value={layer.rotation}
					onValueChange={(value) => update(index, { rotation: value as Layer['rotation'] })}
					class="gap-2.5"
				>
					{#each ROTATIONS as option (option.value)}
						<div class="flex items-start gap-2.5">
							<RadioGroup.Item
								value={option.value}
								id="{layer.id}-{option.value}"
								class="mt-0.5"
							/>
							<Label
								for="{layer.id}-{option.value}"
								class="flex flex-col items-start gap-0.5 font-normal"
							>
								<span class="text-foreground text-sm leading-[1.3]">{option.label}</span>
								<span class="text-subtle-foreground text-[13px] leading-[1.4]">
									{option.description}
								</span>
							</Label>
						</div>
					{/each}
				</RadioGroup.Root>

				{#if layer.rotation === 'custom'}
					<Field.Field class="mt-2 w-[120px] gap-1.5 space-y-0">
						<Field.FieldLabel
							for="{layer.id}-interval"
							class="text-muted-foreground text-[13px] font-medium"
						>
							Every N days
						</Field.FieldLabel>
						<Input
							id="{layer.id}-interval"
							type="number"
							min="1"
							max="30"
							value={layer.intervalDays}
							aria-invalid={errors?.intervalDays ? 'true' : undefined}
							oninput={(event: Event) =>
								update(index, {
									intervalDays: Number((event.currentTarget as HTMLInputElement).value)
								})}
							class="h-[34px] text-[13px]"
						/>
						{#if errors?.intervalDays}
							<Field.FieldError class="text-critical-ink text-xs font-normal">
								{errors.intervalDays}
							</Field.FieldError>
						{/if}
					</Field.Field>
				{/if}
			</div>

			<div class="flex flex-col gap-2.5">
				<Field.Field class="w-[140px] gap-1.5 space-y-0">
					<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
						Handover time
					</Field.FieldLabel>
					<Select.Root
						type="single"
						value={String(layer.handoverHour)}
						onValueChange={(value) => update(index, { handoverHour: Number(value) })}
					>
						<Select.Trigger size="sm">{hour(layer.handoverHour)}</Select.Trigger>
						<Select.Content>
							<Select.Group>
								{#each HOURS as value (value)}
									<Select.Item value={String(value)} label={hour(value)}>{hour(value)}</Select.Item>
								{/each}
							</Select.Group>
						</Select.Content>
					</Select.Root>
					<Field.FieldDescription class="text-subtle-foreground text-xs">
						Local to the schedule, 24-hour.
					</Field.FieldDescription>
				</Field.Field>

				<Field.Field class="w-[170px] gap-1.5 space-y-0">
					<Field.FieldLabel
						for="{layer.id}-start"
						class="text-muted-foreground text-[13px] font-medium"
					>
						Start date
					</Field.FieldLabel>
					<Input
						id="{layer.id}-start"
						type="date"
						required
						value={layer.startsOn}
						aria-invalid={errors?.startsOn ? 'true' : undefined}
						oninput={(event: Event) =>
							update(index, { startsOn: (event.currentTarget as HTMLInputElement).value })}
						class="h-[34px] text-[13px]"
					/>
					{#if errors?.startsOn}
						<Field.FieldError class="text-critical-ink text-xs font-normal">
							{errors.startsOn}
						</Field.FieldError>
					{/if}
				</Field.Field>
			</div>
		</div>

		<div class="text-subtle-foreground text-[11.5px]">On duty {formatDuty(layer)}</div>
	</div>
</div>
