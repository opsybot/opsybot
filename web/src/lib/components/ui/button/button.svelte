<script lang="ts" module>
	import { cn, type WithElementRef } from "$lib/utils.js";
	import type { HTMLAnchorAttributes, HTMLButtonAttributes } from "svelte/elements";
	import { type VariantProps, tv } from "tailwind-variants";

	export const buttonVariants = tv({
		base: "focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:aria-invalid:border-destructive/50 rounded-md border border-transparent bg-clip-padding font-semibold tracking-[0.01em] focus-visible:ring-3 active:not-aria-[haspopup]:translate-y-px aria-invalid:ring-3 [&_svg:not([class*='size-'])]:size-4 group/button inline-flex shrink-0 items-center justify-center leading-none whitespace-nowrap transition-all outline-none select-none disabled:pointer-events-none disabled:opacity-45 [&_svg]:pointer-events-none [&_svg]:shrink-0",
		variants: {
			variant: {
				default: "bg-primary text-primary-foreground border-primary hover:bg-primary-hover",
				secondary:
					"border-border-strong text-foreground hover:border-primary/45 hover:text-brand-foreground bg-transparent",
				ghost:
					"text-muted-foreground hover:bg-accent hover:text-accent-foreground aria-expanded:bg-accent aria-expanded:text-accent-foreground",
				outline:
					"border-border-strong text-foreground hover:border-primary/45 bg-transparent aria-expanded:bg-accent",
				destructive:
					"bg-critical text-critical-foreground border-critical hover:bg-critical-hover focus-visible:ring-destructive/40",
				link: "text-brand-foreground underline-offset-4 hover:underline",
			},
			size: {
				default: "h-[38px] gap-2 px-4 text-sm [&_svg:not([class*='size-'])]:size-[17px]",
				sm: "h-8 gap-1.5 px-3 text-[13px] [&_svg:not([class*='size-'])]:size-[15px]",
				lg: "h-[46px] gap-[9px] px-[22px] text-base [&_svg:not([class*='size-'])]:size-[19px]",
				icon: "size-9 rounded-sm [&_svg:not([class*='size-'])]:size-[18px]",
				"icon-sm": "size-[30px] rounded-sm [&_svg:not([class*='size-'])]:size-4",
				"icon-lg": "size-[42px] rounded-sm [&_svg:not([class*='size-'])]:size-5",
			},
		},
		defaultVariants: {
			variant: "default",
			size: "default",
		},
	});

	export type ButtonVariant = VariantProps<typeof buttonVariants>["variant"];
	export type ButtonSize = VariantProps<typeof buttonVariants>["size"];

	export type ButtonProps = WithElementRef<HTMLButtonAttributes> &
		WithElementRef<HTMLAnchorAttributes> & {
			variant?: ButtonVariant;
			size?: ButtonSize;
		};
</script>

<script lang="ts">
	let {
		class: className,
		variant = "default",
		size = "default",
		ref = $bindable(null),
		href = undefined,
		type = "button",
		disabled,
		children,
		...restProps
	}: ButtonProps = $props();
</script>

{#if href}
	<a
		bind:this={ref}
		data-slot="button"
		class={cn(buttonVariants({ variant, size }), className)}
		href={disabled ? undefined : href}
		aria-disabled={disabled}
		role={disabled ? "link" : undefined}
		tabindex={disabled ? -1 : undefined}
		{...restProps}
	>
		{@render children?.()}
	</a>
{:else}
	<button
		bind:this={ref}
		data-slot="button"
		class={cn(buttonVariants({ variant, size }), className)}
		{type}
		{disabled}
		{...restProps}
	>
		{@render children?.()}
	</button>
{/if}
