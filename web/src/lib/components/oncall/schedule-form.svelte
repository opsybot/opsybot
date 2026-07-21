<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import CheckIcon from '@lucide/svelte/icons/check';
	import GlobeIcon from '@lucide/svelte/icons/globe';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import { untrack } from 'svelte';
	import { superForm, type SuperValidated } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import { page } from '$app/state';
	import { Alert, AlertContent, AlertTitle } from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import TimezoneSelect from '$lib/components/timezone-select.svelte';
	import { scheduleSchema } from '$lib/schemas/oncall';
	import { isSoloLayer, SOLO_LAYER_NOTE, type DaySummary, type Layer } from '$lib/oncall';
	import LayerCard from './layer-card.svelte';
	import PreviewGrid from './preview-grid.svelte';

	type FormValues = { name: string; team: string; timezone: string; layers: Layer[] };
	type PreviewData = {
		days: { label: string; num: number; iso: string }[];
		effective: DaySummary[];
		rows: { label: string; title: string; days: DaySummary[] }[];
	};

	let {
		data,
		heading,
		submitLabel,
		back,
		previewFrom,
		people,
		teams
	}: {
		data: SuperValidated<FormValues>;
		heading: string;
		submitLabel: string;
		back: string;
		previewFrom: string;
		people: string[];
		teams: string[];
	} = $props();

	const form = superForm(untrack(() => data), {
		dataType: 'json',
		validators: zod4Client(scheduleSchema)
	});
	const { form: formData, errors, enhance } = form;

	const layers = $derived($formData.layers);
	const nobodyIn = $derived(layers.some((layer) => layer.participants.length === 0));
	const everyLayerSolo = $derived(layers.length > 0 && layers.every(isSoloLayer));
	const layerErrors = $derived(
		($errors.layers ?? {}) as Record<
			number,
			{ participants?: string[]; intervalDays?: string[]; startsOn?: string[] } | undefined
		>
	);

	let preview = $state<PreviewData | null>(null);
	let previewLoading = $state(false);

	$effect(() => {
		const payload = { timezone: $formData.timezone, layers: $formData.layers, from: previewFrom };
		const empty = payload.layers.some((layer) => layer.participants.length === 0);
		if (empty) {
			preview = null;
			return;
		}
		previewLoading = true;
		const handle = setTimeout(async () => {
			try {
				const res = await fetch(`/${page.params.workspace}/on-call/preview`, {
					method: 'POST',
					headers: { 'content-type': 'application/json' },
					body: JSON.stringify(payload)
				});
				if (res.ok) preview = await res.json();
			} catch {
			}
			previewLoading = false;
		}, 300);
		return () => clearTimeout(handle);
	});

	function update(index: number, patch: Partial<Layer>) {
		$formData.layers = layers.map((layer, position) =>
			position === index ? { ...layer, ...patch } : layer
		);
	}

	function move(index: number, by: number) {
		const next = [...layers];
		[next[index], next[index + by]] = [next[index + by], next[index]];
		$formData.layers = next;
	}

	function remove(index: number) {
		$formData.layers = layers.filter((_, position) => position !== index);
	}

	function addLayer() {
		$formData.layers = [
			...layers,
			{
				id: `l-${Math.random().toString(36).slice(2, 7)}`,
				participants: [],
				rotation: 'weekly',
				intervalDays: 7,
				handoverHour: 9,
				startsOn: previewFrom,
				restrictions: []
			}
		];
	}
</script>

<div class="flex flex-col gap-3.5">
	<a
		href={back}
		class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[12.5px]"
	>
		<ArrowLeftIcon class="size-3.5" />
		{back.endsWith('/on-call') ? 'All schedules' : 'Back to schedule'}
	</a>

	<h2 class="tracking-heading m-0 text-[18px] font-semibold">{heading}</h2>

	<form method="POST" use:enhance>
		<div class="grid items-start gap-3.5 min-[1100px]:grid-cols-[minmax(0,1fr)_360px]">
			<div class="flex min-w-0 flex-col gap-3.5">
				<div class="bg-card flex flex-wrap gap-3 rounded-xl border p-4">
					<Field.Field class="min-w-[200px] flex-1 gap-1.5 space-y-0">
						<Field.FieldLabel for="name" class="text-muted-foreground text-[13px] font-medium">
							Name
						</Field.FieldLabel>
						<Input
							id="name"
							name="name"
							bind:value={$formData.name}
							aria-invalid={$errors.name ? 'true' : undefined}
							class="font-mono"
						/>
						{#if $errors.name}
							<Field.FieldError class="text-critical-ink text-xs font-normal">
								{$errors.name}
							</Field.FieldError>
						{:else}
							<Field.FieldDescription class="text-subtle-foreground text-xs">
								Used in the calendar feed URL.
							</Field.FieldDescription>
						{/if}
					</Field.Field>

					<Field.Field class="w-[180px] gap-1.5 space-y-0">
						<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
							Team
						</Field.FieldLabel>
						<Select.Root type="single" name="team" bind:value={$formData.team}>
							<Select.Trigger>{$formData.team || 'Pick a team'}</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each teams as team (team)}
										<Select.Item value={team} label={team}>{team}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
						{#if $errors.team}
							<Field.FieldError class="text-critical-ink text-xs font-normal">
								{$errors.team}
							</Field.FieldError>
						{/if}
					</Field.Field>

					<div class="w-[220px]">
						<TimezoneSelect {form} name="timezone" label="Timezone" />
					</div>
				</div>

				<Alert tone="info">
					<GlobeIcon />
					<AlertContent>
						<AlertTitle>How timezones work</AlertTitle>
						The schedule runs in the timezone you pick. Handovers and restriction hours happen at that
						local time: a 09:00 handover in Europe/Berlin stays 09:00 there across daylight saving.
						Everyone still sees the calendar in their own timezone.
					</AlertContent>
				</Alert>

				{#each layers as layer, index (layer.id)}
					<LayerCard
						{layer}
						{index}
						{people}
						total={layers.length}
						errors={layerErrors[index]}
						{update}
						{move}
						{remove}
					/>
				{/each}

				<Button type="button" variant="secondary" size="sm" class="self-start" onclick={addLayer}>
					<PlusIcon data-icon="inline-start" />
					Add layer
				</Button>

				<div class="flex gap-2.5 pt-1">
					<Button type="submit" disabled={nobodyIn}>
						<CheckIcon data-icon="inline-start" />
						{submitLabel}
					</Button>
					<Button variant="ghost" href={back}>Cancel</Button>
				</div>
			</div>

			<section class="bg-card overflow-hidden rounded-xl border">
				<header class="flex items-center gap-2.5 border-b px-4 py-3">
					<span class="text-[13.5px] font-semibold">Preview</span>
					<span class="text-subtle-foreground ml-auto font-mono text-[10.5px]">
						{$formData.timezone}
					</span>
				</header>
				<div class="overflow-x-auto p-3">
					{#if preview}
						<PreviewGrid days={preview.days} effective={preview.effective} rows={preview.rows} />
					{:else if previewLoading}
						<p class="text-subtle-foreground m-0 px-0.5 py-6 text-center text-[12px]">
							Building preview…
						</p>
					{:else}
						<p class="text-subtle-foreground m-0 px-0.5 py-6 text-center text-[12px]">
							Add someone to every layer to preview the rotation.
						</p>
					{/if}
					{#if everyLayerSolo}
						<p class="text-warning-ink m-0 mt-2.5 px-0.5 text-[11px] leading-[1.5]">
							{SOLO_LAYER_NOTE}
						</p>
					{/if}
					<p class="text-subtle-foreground m-0 mt-2.5 px-0.5 text-[11px] leading-[1.5]">
						The top row is who actually gets paged: the highest layer on duty with people in it wins.
					</p>
				</div>
			</section>
		</div>
	</form>
</div>
