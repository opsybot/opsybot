<script lang="ts">
	import * as Field from '$lib/components/ui/field';
	import * as Select from '$lib/components/ui/select';
	import { cn } from '$lib/utils';

	type Option = { value: string; label: string };

	let {
		label,
		options,
		value = $bindable(),
		size = 'default',
		class: className,
		'aria-label': ariaLabel
	}: {
		label?: string;
		options: (string | Option)[];
		value: string;
		size?: 'sm' | 'default';
		class?: string;
		'aria-label'?: string;
	} = $props();

	const normalized = $derived(
		options.map((option) => (typeof option === 'string' ? { value: option, label: option } : option))
	);
	const current = $derived(normalized.find((option) => option.value === value)?.label ?? value);
</script>

<Field.Field class={cn('gap-1.5 space-y-0', className)}>
	{#if label}
		<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">{label}</Field.FieldLabel>
	{/if}
	<Select.Root type="single" bind:value>
		<Select.Trigger {size} aria-label={ariaLabel ?? label}>{current}</Select.Trigger>
		<Select.Content>
			{#each normalized as option (option.value)}
				<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
			{/each}
		</Select.Content>
	</Select.Root>
</Field.Field>
