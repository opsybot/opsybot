<script lang="ts" module>
	const STEP_MINUTES = 15;

	export const TIME_OPTIONS = Array.from(
		{ length: (24 * 60) / STEP_MINUTES },
		(_, index) =>
			`${String(Math.floor((index * STEP_MINUTES) / 60)).padStart(2, '0')}:` +
			`${String((index * STEP_MINUTES) % 60).padStart(2, '0')}`
	);

	export function timeOptions(value: string): string[] {
		if (!value || TIME_OPTIONS.includes(value)) return TIME_OPTIONS;
		return [...TIME_OPTIONS, value].sort();
	}
</script>

<script lang="ts">
	import * as Select from '$lib/components/ui/select';

	let {
		value = $bindable(''),
		name,
		id,
		label,
		size = 'default',
		class: className,
		onChange
	}: {
		value?: string;
		name?: string;
		id?: string;
		label?: string;
		size?: 'default' | 'sm';
		class?: string;
		onChange?: (value: string) => void;
	} = $props();

	const options = $derived(timeOptions(value));

	function pick(next: string) {
		value = next;
		onChange?.(next);
	}
</script>

{#if name}
	<input type="hidden" {name} {value} />
{/if}

<Select.Root type="single" {value} onValueChange={pick}>
	<Select.Trigger {id} aria-label={label} size={size === 'sm' ? 'sm' : undefined} class={className}>
		{value || 'Pick a time'}
	</Select.Trigger>
	<Select.Content>
		<Select.Group>
			{#each options as option (option)}
				<Select.Item value={option} label={option}>{option}</Select.Item>
			{/each}
		</Select.Group>
	</Select.Content>
</Select.Root>
