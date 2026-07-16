<script lang="ts">
	import { cn } from '$lib/utils';

	let {
		data,
		tone = 'brand',
		height = 22,
		class: className
	}: {
		data: number[];
		tone?: 'brand' | 'info' | 'warning' | 'critical' | 'success';
		height?: number;
		class?: string;
	} = $props();

	const colour = $derived(tone === 'brand' ? 'var(--primary)' : `var(--${tone})`);

	const points = $derived.by(() => {
		const max = Math.max(...data, 1);
		return data
			.map((value, index) => `${(index / (data.length - 1)) * 100},${100 - (value / max) * 100}`)
			.join(' ');
	});
</script>

<svg
	viewBox="0 0 100 100"
	preserveAspectRatio="none"
	aria-hidden="true"
	style="height: {height}px"
	class={cn('block w-full', className)}
>
	<polyline
		{points}
		fill="none"
		stroke={colour}
		stroke-width="2"
		vector-effect="non-scaling-stroke"
		stroke-linejoin="round"
		stroke-linecap="round"
	/>
</svg>
