<script lang="ts">
	import { cn } from '$lib/utils';

	let {
		value,
		max = 100,
		tone = 'brand',
		size = 'md',
		class: className
	}: {
		value: number;
		max?: number;
		tone?: 'brand' | 'info' | 'warning' | 'critical' | 'success';
		size?: 'sm' | 'md' | 'lg';
		class?: string;
	} = $props();

	const percent = $derived(Math.min(100, Math.max(0, (value / max) * 100)));
	const colour = $derived(tone === 'brand' ? 'var(--primary)' : `var(--${tone})`);
	const height = $derived({ sm: 4, md: 6, lg: 10 }[size]);
</script>

<div
	role="progressbar"
	aria-valuenow={value}
	aria-valuemax={max}
	class={cn('bg-accent overflow-hidden rounded-full', className)}
	style="height: {height}px"
>
	<div
		class="h-full rounded-full transition-[width] duration-[360ms] ease-out"
		style="width: {percent}%; background: {colour}"
	></div>
</div>
