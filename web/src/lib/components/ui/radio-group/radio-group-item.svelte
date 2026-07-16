<script lang="ts">
	import { RadioGroup as RadioGroupPrimitive } from "bits-ui";
	import CircleIcon from '@lucide/svelte/icons/circle';
	import { cn, type WithoutChildrenOrChild } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: WithoutChildrenOrChild<RadioGroupPrimitive.ItemProps> = $props();
</script>

<RadioGroupPrimitive.Item
	bind:ref
	data-slot="radio-group-item"
	class={cn(
		"bg-inset border-border-strong data-checked:border-primary focus-visible:border-primary focus-visible:shadow-[var(--focus-ring)] aria-invalid:border-critical relative flex size-[18px] shrink-0 items-center justify-center rounded-full border transition-[border-color,box-shadow] duration-[120ms] ease-out outline-none disabled:cursor-not-allowed disabled:opacity-50",
		className
	)}
	{...restProps}
>
	{#snippet children({ checked })}
		<div data-slot="radio-group-indicator" class="flex size-4 items-center justify-center">
			{#if checked}
				<CircleIcon class="bg-primary absolute top-1/2 left-1/2 size-2 -translate-x-1/2 -translate-y-1/2 rounded-full text-transparent" />
			{/if}
		</div>
	{/snippet}
</RadioGroupPrimitive.Item>
