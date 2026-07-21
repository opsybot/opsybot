<script lang="ts">
	import MoonIcon from '@lucide/svelte/icons/moon';
	import SirenIcon from '@lucide/svelte/icons/siren';
	import * as Alert from '$lib/components/ui/alert';
	import * as Select from '$lib/components/ui/select';
	import { Switch } from '$lib/components/ui/switch';
	import { DAYS, HOUR_OPTIONS, TIMEZONE_OPTIONS, type QuietHours } from '$lib/notifications';

	let { value = $bindable() }: { value: QuietHours } = $props();

	function toggleDay(day: string) {
		const next = value.days.includes(day)
			? value.days.filter((other) => other !== day)
			: [...value.days, day];
		value = { ...value, days: DAYS.map((entry) => entry.value).filter((entry) => next.includes(entry)) };
	}
</script>

{#snippet hourField(label: string, selected: string, width: string, options: { value: string; label: string }[], onpick: (value: string) => void)}
	<div class="flex flex-col gap-1.5" style="width:{width}">
		<span class="text-muted-foreground text-[13px] font-medium">{label}</span>
		<Select.Root type="single" value={selected} onValueChange={onpick}>
			<Select.Trigger size="sm" aria-label={label}>{selected}</Select.Trigger>
			<Select.Content>
				<Select.Group>
					{#each options as option (option.value)}
						<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
	</div>
{/snippet}

<div class="bg-card overflow-hidden rounded-xl border">
	<header class="flex items-center gap-2 border-b px-4 py-3">
		<MoonIcon class="text-subtle-foreground size-3.5" />
		<span class="text-[13.5px] font-semibold">Quiet hours</span>
		<div class="flex-1"></div>
		<Switch bind:checked={value.enabled} aria-label="Enable quiet hours" />
	</header>
	{#if value.enabled}
		<div class="flex flex-col gap-3.5 p-3.5">
			<div class="flex flex-wrap items-start gap-2.5">
				{@render hourField('From', value.start, '110px', HOUR_OPTIONS, (v) => (value = { ...value, start: v }))}
				{@render hourField('Until', value.end, '110px', HOUR_OPTIONS, (v) => (value = { ...value, end: v }))}
				{@render hourField(
					'Timezone',
					value.timezone,
					'200px',
					TIMEZONE_OPTIONS.map((zone) => ({ value: zone, label: zone })),
					(v) => (value = { ...value, timezone: v })
				)}
			</div>
			<div role="group" aria-label="Quiet hours days" class="flex flex-wrap gap-1.5">
				{#each DAYS as day (day.value)}
					{@const on = value.days.includes(day.value)}
					<button
						type="button"
						aria-pressed={on}
						aria-label={day.full}
						onclick={() => toggleDay(day.value)}
						class="w-11 rounded-full border py-[5px] text-center text-[11.5px] transition-colors
						{on
							? 'bg-brand-wash border-brand-edge text-brand-foreground font-semibold'
							: 'bg-inset text-subtle-foreground hover:text-muted-foreground'}"
					>
						{day.value}
					</button>
				{/each}
			</div>
			<Alert.Root tone="info">
				<SirenIcon />
				<Alert.Content>
					<Alert.Description>
						Pages always break through quiet hours. Quiet hours only mute the low-urgency set.
					</Alert.Description>
				</Alert.Content>
			</Alert.Root>
		</div>
	{:else}
		<div class="text-subtle-foreground p-3.5 text-[12.5px]">
			Off: low-urgency notifications arrive around the clock.
		</div>
	{/if}
</div>
