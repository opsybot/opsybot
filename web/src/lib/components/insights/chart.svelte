<script lang="ts">
	let {
		type,
		data,
		height = 120,
		tone = 'brand',
		labels,
		showAxis = true
	}: {
		type: 'line' | 'bar';
		data: number[];
		height?: number;
		tone?: 'brand' | 'info' | 'warning' | 'critical' | 'success';
		labels?: [string, string];
		showAxis?: boolean;
	} = $props();

	const colour = $derived(tone === 'brand' ? 'var(--primary)' : `var(--${tone})`);
	const max = $derived(Math.max(...data, 1));

	const span = $derived(Math.max(1, data.length - 1));
	const points = $derived(
		data.map((value, index) => `${(index / span) * 100},${100 - (value / max) * 100}`).join(' ')
	);

	const showLabels = $derived(showAxis && !!labels && labels.some(Boolean));
</script>

<!-- Decorative: the surrounding card states these numbers in text -->
<div class="font-sans" aria-hidden="true">
	{#if type === 'line'}
		<svg
			viewBox="0 0 100 100"
			preserveAspectRatio="none"
			style="display: block; width: 100%; height: {height}px"
		>
			<polyline
				points="0,100 {points} 100,100"
				fill="color-mix(in srgb, {colour} 14%, transparent)"
				stroke="none"
			/>
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
	{:else}
		<div class="flex items-end gap-1" style="height: {height}px">
			{#each data as value, index (index)}
				<div
					title={String(value)}
					style="flex: 1; height: {(value / max) * 100}%; min-height: 2px; border-radius: 2px 2px 0 0; background: {colour}; opacity: {0.55 +
						0.45 * (value / max)}; transition: height 300ms var(--ease-out)"
				></div>
			{/each}
		</div>
	{/if}

	{#if showLabels && labels}
		<div class="text-subtle-foreground mt-1.5 flex justify-between text-2xs">
			<span>{labels[0]}</span>
			<span>{labels[1]}</span>
		</div>
	{/if}
</div>
