<script lang="ts">
	import { PinInput as InputOTPPrimitive } from "bits-ui";
	import { cn } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		cell,
		class: className,
		...restProps
	}: InputOTPPrimitive.CellProps = $props();
</script>

<InputOTPPrimitive.Cell
	{cell}
	bind:ref
	data-slot="input-otp-slot"
	class={cn(
		"bg-inset border-border-strong text-foreground data-[active=true]:border-primary data-[active=true]:shadow-[var(--focus-ring)] aria-invalid:border-critical rounded-sm relative flex h-[46px] w-10 items-center justify-center border font-mono text-[18px] font-semibold outline-none transition-[border-color,box-shadow] duration-[120ms] ease-out data-[active=true]:z-10",
		className
	)}
	{...restProps}
>
	{cell.char}
</InputOTPPrimitive.Cell>
