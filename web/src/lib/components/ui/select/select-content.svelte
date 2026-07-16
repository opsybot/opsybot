<script lang="ts">
	import { Select as SelectPrimitive } from "bits-ui";
	import SelectPortal from "./select-portal.svelte";
	import SelectScrollUpButton from "./select-scroll-up-button.svelte";
	import SelectScrollDownButton from "./select-scroll-down-button.svelte";
	import { cn, type WithoutChild } from "$lib/utils.js";
	import type { ComponentProps } from "svelte";
	import type { WithoutChildrenOrChild } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		sideOffset = 4,
		portalProps,
		children,
		preventScroll = true,
		...restProps
	}: WithoutChild<SelectPrimitive.ContentProps> & {
		portalProps?: WithoutChildrenOrChild<ComponentProps<typeof SelectPortal>>;
	} = $props();
</script>

<SelectPortal {...portalProps}>
	<SelectPrimitive.Content
		bind:ref
		{sideOffset}
		{preventScroll}
		data-slot="select-content"
		class={cn(
			"bg-popover text-popover-foreground shadow-pop border-input data-open:animate-in data-closed:animate-out data-closed:fade-out-0 data-open:fade-in-0 relative z-50 max-h-[260px] min-w-[8rem] overflow-x-hidden overflow-y-auto rounded-lg border p-[5px] ring-0",
			className
		)}
		{...restProps}
	>
		<SelectScrollUpButton />
		<SelectPrimitive.Viewport
			class={cn(
				"h-(--bits-select-anchor-height) w-full min-w-(--bits-select-anchor-width) scroll-my-1"
			)}
		>
			{@render children?.()}
		</SelectPrimitive.Viewport>
		<SelectScrollDownButton />
	</SelectPrimitive.Content>
</SelectPortal>
