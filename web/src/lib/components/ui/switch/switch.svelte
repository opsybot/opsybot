<script lang="ts">
	import { Switch as SwitchPrimitive } from "bits-ui";
	import { cn, type WithoutChildrenOrChild } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		checked = $bindable(false),
		size = "default",
		...restProps
	}: WithoutChildrenOrChild<SwitchPrimitive.RootProps> & {
		size?: "sm" | "default";
	} = $props();
</script>

<SwitchPrimitive.Root
	bind:ref
	bind:checked
	data-slot="switch"
	data-size={size}
	class={cn(
		"data-checked:bg-primary data-checked:border-primary border-border-strong focus-visible:border-primary focus-visible:shadow-[var(--focus-ring)] data-[size=default]:h-6 data-[size=default]:w-[42px] data-[size=sm]:h-5 data-[size=sm]:w-[34px] peer group/switch relative inline-flex shrink-0 items-center rounded-full border bg-[var(--ink-5)] transition-[background-color,border-color] duration-200 ease-out outline-none after:absolute after:-inset-x-3 after:-inset-y-2 data-disabled:cursor-not-allowed data-disabled:opacity-50",
		className
	)}
	{...restProps}
>
	<SwitchPrimitive.Thumb
		data-slot="switch-thumb"
		class="bg-popover data-checked:bg-[var(--text-inverse)] group-data-[size=default]/switch:size-[18px] group-data-[size=sm]/switch:size-3.5 pointer-events-none block translate-x-0.5 rounded-full shadow-[0_1px_3px_rgb(0_0_0/0.5)] ring-0 transition-transform duration-200 ease-[var(--ease-spring)] group-data-[size=default]/switch:data-checked:translate-x-[21px] group-data-[size=sm]/switch:data-checked:translate-x-[16px]"
	/>
</SwitchPrimitive.Root>
