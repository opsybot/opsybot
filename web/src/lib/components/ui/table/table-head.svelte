<script lang="ts">
	import { cn, type WithElementRef } from "$lib/utils.js";
	import type { HTMLThAttributes } from "svelte/elements";

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLThAttributes> = $props();
</script>

<th bind:this={ref} data-slot="table-head" class={cn("text-subtle-foreground border-b px-4 py-2.5 text-left align-middle text-2xs font-semibold tracking-[0.05em] whitespace-nowrap uppercase [&:has([role=checkbox])]:pr-0", className)} {...restProps}>
	{@render children?.()}
</th>
