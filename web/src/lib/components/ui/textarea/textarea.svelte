<script lang="ts">
	import { cn, type WithElementRef, type WithoutChildren } from "$lib/utils.js";
	import type { HTMLTextareaAttributes } from "svelte/elements";

	let {
		ref = $bindable(null),
		value = $bindable(),
		class: className,
		"data-slot": dataSlot = "textarea",
		...restProps
	}: WithoutChildren<WithElementRef<HTMLTextareaAttributes>> = $props();
</script>

<textarea
	bind:this={ref}
	data-slot={dataSlot}
	class={cn(
		"bg-inset border-border-strong text-foreground placeholder:text-subtle-foreground focus-visible:border-primary focus-visible:shadow-[var(--focus-ring)] aria-invalid:border-critical min-h-[60px] w-full resize-y rounded-md border px-3 py-2.5 text-sm leading-[1.5] outline-none transition-[border-color,box-shadow] duration-[120ms] ease-out disabled:pointer-events-none disabled:opacity-50",
		className
	)}
	bind:value
	{...restProps}
></textarea>
