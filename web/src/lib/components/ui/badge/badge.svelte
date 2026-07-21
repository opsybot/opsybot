<script lang="ts" module>
	import { type VariantProps, tv } from "tailwind-variants";

	export const badgeVariants = tv({
		base: "inline-flex w-fit shrink-0 items-center justify-center gap-[5px] overflow-hidden rounded-sm border leading-none font-semibold tracking-[0.01em] whitespace-nowrap [&>svg]:pointer-events-none [&>svg]:size-3",
		variants: {
			tone: {
				neutral: "text-neutral-ink bg-neutral-wash border-neutral-edge",
				brand: "text-brand-foreground bg-brand-wash border-brand-edge",
				critical: "text-critical-ink bg-critical-wash border-critical-edge",
				high: "text-high-ink bg-high-wash border-high-edge",
				warning: "text-warning-ink bg-warning-wash border-warning-edge",
				info: "text-info-ink bg-info-wash border-info-edge",
				success: "text-success-ink bg-success-wash border-success-edge",
			},
			variant: {
				soft: "",
				solid: "text-critical-foreground",
			},
			size: {
				sm: "h-[18px] px-[7px] text-2xs",
				md: "h-[22px] px-[9px] text-xs",
			},
		},
		compoundVariants: [
			{ variant: "solid", tone: "neutral", class: "bg-neutral border-neutral" },
			{
				variant: "solid",
				tone: "brand",
				class: "bg-primary border-primary text-primary-foreground",
			},
			{ variant: "solid", tone: "critical", class: "bg-critical border-critical" },
			{ variant: "solid", tone: "high", class: "bg-high border-high" },
			{ variant: "solid", tone: "warning", class: "bg-warning border-warning" },
			{ variant: "solid", tone: "info", class: "bg-info border-info" },
			{ variant: "solid", tone: "success", class: "bg-success border-success" },
		],
		defaultVariants: {
			tone: "neutral",
			variant: "soft",
			size: "md",
		},
	});

	export type BadgeTone = VariantProps<typeof badgeVariants>["tone"];
	export type BadgeVariant = VariantProps<typeof badgeVariants>["variant"];
	export type BadgeSize = VariantProps<typeof badgeVariants>["size"];
</script>

<script lang="ts">
	import type { HTMLAnchorAttributes } from "svelte/elements";
	import { cn, type WithElementRef } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		href,
		class: className,
		tone = "neutral",
		variant = "soft",
		size = "md",
		dot = false,
		children,
		...restProps
	}: WithElementRef<HTMLAnchorAttributes> & {
		tone?: BadgeTone;
		variant?: BadgeVariant;
		size?: BadgeSize;
		/** A 6px disc before the label, for a badge that stands for a live state. */
		dot?: boolean;
	} = $props();
</script>

<svelte:element
	this={href ? "a" : "span"}
	bind:this={ref}
	data-slot="badge"
	{href}
	class={cn(badgeVariants({ tone, variant, size }), className)}
	{...restProps}
>
	{#if dot}
		<span class="size-1.5 shrink-0 rounded-full bg-current"></span>
	{/if}
	{@render children?.()}
</svelte:element>
