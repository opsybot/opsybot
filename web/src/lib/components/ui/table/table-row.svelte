<script lang="ts">
	import { cn, type WithElementRef } from "$lib/utils.js";
	import type { HTMLAttributes } from "svelte/elements";

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLTableRowElement>> = $props();
</script>

<tr bind:this={ref} data-slot="table-row" class={cn("data-[state=selected]:bg-accent border-t transition-colors duration-[120ms] ease-out first:border-t-0 data-[clickable=true]:cursor-pointer data-[clickable=true]:hover:bg-accent", className)} {...restProps}>
	{@render children?.()}
</tr>
