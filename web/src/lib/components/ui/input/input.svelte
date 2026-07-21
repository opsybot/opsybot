<script lang="ts">
	import type { HTMLInputAttributes, HTMLInputTypeAttribute } from "svelte/elements";
	import { cn, type WithElementRef } from "$lib/utils.js";

	type InputType = Exclude<HTMLInputTypeAttribute, "file">;

	type Props = WithElementRef<
		Omit<HTMLInputAttributes, "type"> &
			({ type: "file"; files?: FileList } | { type?: InputType; files?: undefined })
	>;

	let {
		ref = $bindable(null),
		value = $bindable(),
		type,
		files = $bindable(),
		class: className,
		"data-slot": dataSlot = "input",
		...restProps
	}: Props = $props();

	const base =
		"bg-inset border-border-strong text-foreground placeholder:text-subtle-foreground focus-visible:border-primary focus-visible:shadow-[var(--focus-ring)] aria-invalid:border-critical h-10 w-full min-w-0 rounded-sm border px-3 text-sm outline-none transition-[border-color,box-shadow] duration-[120ms] ease-out disabled:pointer-events-none disabled:opacity-50";
</script>

{#if type === "file"}
	<input
		bind:this={ref}
		data-slot={dataSlot}
		class={cn(
			base,
			"file:text-foreground file:inline-flex file:border-0 file:bg-transparent file:text-sm file:font-medium",
			className
		)}
		type="file"
		bind:files
		bind:value
		{...restProps}
	/>
{:else}
	<input
		bind:this={ref}
		data-slot={dataSlot}
		class={cn(base, className)}
		{type}
		bind:value
		{...restProps}
	/>
{/if}
