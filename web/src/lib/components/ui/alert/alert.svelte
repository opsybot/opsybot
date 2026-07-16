<script lang="ts" module>
	import { type VariantProps, tv } from "tailwind-variants";

	// The design system's Alert: a tone's wash, its hairline, and its icon. The body text stays
	// neutral so the message is readable regardless of how loud the tone is.
	export const alertVariants = tv({
		base: "flex w-full gap-3 rounded-md border px-[15px] py-[13px] text-left *:[svg]:mt-px *:[svg]:size-[18px] *:[svg]:shrink-0",
		variants: {
			tone: {
				neutral: "bg-neutral-wash border-neutral-edge *:[svg]:text-neutral-ink",
				brand: "bg-brand-wash border-brand-edge *:[svg]:text-brand-foreground",
				info: "bg-info-wash border-info-edge *:[svg]:text-info-ink",
				success: "bg-success-wash border-success-edge *:[svg]:text-success-ink",
				warning: "bg-warning-wash border-warning-edge *:[svg]:text-warning-ink",
				critical: "bg-critical-wash border-critical-edge *:[svg]:text-critical-ink",
			},
		},
		defaultVariants: {
			tone: "info",
		},
	});

	export type AlertTone = VariantProps<typeof alertVariants>["tone"];
</script>

<script lang="ts">
	import type { HTMLAttributes } from "svelte/elements";
	import { cn, type WithElementRef } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		tone = "info",
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLDivElement>> & {
		tone?: AlertTone;
	} = $props();
</script>

<div
	bind:this={ref}
	data-slot="alert"
	role="alert"
	class={cn(alertVariants({ tone }), className)}
	{...restProps}
>
	{@render children?.()}
</div>
