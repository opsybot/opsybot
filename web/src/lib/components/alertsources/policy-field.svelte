<script lang="ts">
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
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
	<Field.FieldLabel for={id} class="text-muted-foreground text-[13px] font-medium">
		{label}
	</Field.FieldLabel>
	<Input {id} bind:value list="{id}-known" class="font-mono" placeholder="platform-default" />
	<datalist id="{id}-known">
		{#each known as ref (ref)}
			<option value={ref}></option>
		{/each}
	</datalist>
	{#if description}
		<Field.FieldDescription class="text-subtle-foreground text-xs">
			{description}
		</Field.FieldDescription>
	{/if}
</Field.Field>
