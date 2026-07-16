<script lang="ts">
	import { Select as SelectPrimitive } from "bits-ui";
	import { cn, type WithoutChild } from "$lib/utils.js";
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';

	let {
		ref = $bindable(null),
		class: className,
		children,
		size = "default",
		...restProps
	}: WithoutChild<SelectPrimitive.TriggerProps> & {
		size?: "sm" | "default";
	} = $props();
</script>

<SelectPrimitive.Trigger
	bind:ref
	data-slot="select-trigger"
	data-size={size}
	class={cn(
		"bg-inset border-border-strong text-foreground data-[state=open]:border-primary data-[state=open]:shadow-[var(--focus-ring)] focus-visible:border-primary focus-visible:shadow-[var(--focus-ring)] data-[size=sm]:h-[34px] data-[size=sm]:text-[13px] flex h-10 w-full items-center justify-between gap-2 rounded-md border py-0 pr-2.5 pl-3 text-sm whitespace-nowrap outline-none transition-[border-color,box-shadow] duration-[120ms] ease-out disabled:cursor-not-allowed disabled:opacity-50 data-placeholder:text-subtle-foreground *:data-[slot=select-value]:min-w-0 *:data-[slot=select-value]:truncate [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-[15px] [&_svg:not([class*='text-'])]:text-subtle-foreground",
		className
	)}
	{...restProps}
>
	{@render children?.()}
	<ChevronDownIcon class="text-muted-foreground size-4 pointer-events-none" />
</SelectPrimitive.Trigger>
