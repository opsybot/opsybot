<script lang="ts">
	import * as Field from '$lib/components/ui/field';
	import * as Select from '$lib/components/ui/select';
	import { cn } from '$lib/utils';

	let {
		id,
		label = 'Route to policy',
		known = [],
		description,
		value = $bindable(''),
		class: className
	}: {
		id: string;
		label?: string;
		known?: string[];
		description?: string;
		value?: string;
		class?: string;
	} = $props();
</script>

<Field.Field class={cn('gap-1.5 space-y-0', className)}>
	{#if label}
		<Field.FieldLabel for={id} class="text-muted-foreground text-[13px] font-medium">
			{label}
		</Field.FieldLabel>
	{/if}
	{#if known.length === 0}
		<p class="text-warning-ink m-0 text-[12.5px]">
			No escalation policies yet. Create one under Escalation policies first.
		</p>
	{:else}
		<Select.Root type="single" bind:value>
			<Select.Trigger {id} size="sm" class="w-full font-mono" aria-label={label || 'Escalation policy'}>
				{value || 'Pick a policy'}
			</Select.Trigger>
			<Select.Content>
				<Select.Group>
					{#each known as slug (slug)}
						<Select.Item value={slug} label={slug}>{slug}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
	{/if}
	{#if description}
		<Field.FieldDescription class="text-subtle-foreground text-xs">
			{description}
		</Field.FieldDescription>
	{/if}
</Field.Field>
