<script lang="ts">
	import { cn } from '$lib/utils';

	let {
		value,
		max = 100,
		tone = 'brand',
		class: className
	}: {
		value: number;
		max?: number;
		tone?: 'brand' | 'info' | 'warning' | 'critical' | 'success';
		class?: string;
	} = $props();

	const percent = $derived(Math.min(100, Math.max(0, (value / max) * 100)));
	const colour = $derived(tone === 'brand' ? 'var(--primary)' : `var(--${tone})`);
</script>

<div class={cn('bg-inset h-[7px] overflow-hidden rounded-full', className)}>
	<span class="block h-full rounded-full" style="width: {percent}%; background: {colour}"></span>
</div>
