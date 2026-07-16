<script lang="ts">
	let { tones }: { tones: string[] } = $props();

	const colour = (tone: string) =>
		tone === 'mint' ? 'var(--primary)' : tone === 'faint' ? 'var(--border)' : 'var(--border-strong)';
</script>

<svg class="block h-[30px] w-full" viewBox="0 0 100 30" preserveAspectRatio="none" aria-hidden="true">
	<path d="M50 0 L50 9" fill="none" stroke="var(--border-strong)" stroke-width="1.5" vector-effect="non-scaling-stroke" />
	<path
		d="M50 9 C 50 24 23.7 15 23.7 30"
		fill="none"
		stroke={colour(tones[0])}
		stroke-width="1.5"
		vector-effect="non-scaling-stroke"
	/>
	<path
		d="M50 9 C 50 24 76.3 15 76.3 30"
		fill="none"
		stroke={colour(tones[1])}
		stroke-width="1.5"
		vector-effect="non-scaling-stroke"
	/>
</svg>
