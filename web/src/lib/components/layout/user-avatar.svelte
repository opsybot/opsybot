<script lang="ts" module>
	export type Presence = 'online' | 'busy' | 'away' | 'offline';
	export type AvatarSize = 'xs' | 'sm' | 'md';

	const DIAMETER: Record<AvatarSize, number> = { xs: 22, sm: 28, md: 36 };

	const PRESENCE_COLOUR: Record<Presence, string> = {
		online: 'var(--success)',
		busy: 'var(--critical)',
		away: 'var(--warning)',
		offline: 'var(--neutral)'
	};

	const PALETTE = ['#00e5ac', '#4da3ff', '#ff7a45', '#f5b23d', '#c084fc', '#ff4d5e', '#22d3a0'];

	function hueFor(name: string): string {
		let hash = 0;
		for (let i = 0; i < name.length; i++) hash = (hash * 31 + name.charCodeAt(i)) >>> 0;
		return PALETTE[hash % PALETTE.length];
	}

	function initialsFor(name: string): string {
		const parts = name.trim().split(/\s+/);
		return ((parts[0]?.[0] ?? '') + (parts[1]?.[0] ?? '')).toUpperCase() || '?';
	}
</script>

<script lang="ts">
	import * as Avatar from '$lib/components/ui/avatar';

	let {
		name,
		size = 'sm',
		onCall = false,
		presence
	}: {
		name: string;
		size?: AvatarSize;
		onCall?: boolean;
		presence?: Presence;
	} = $props();

	const diameter = $derived(DIAMETER[size]);
	const hue = $derived(hueFor(name));

	const ring = $derived(
		onCall
			? '0 0 0 2px var(--ink-1), 0 0 0 4px var(--primary)'
			: 'inset 0 0 0 1px var(--hairline)'
	);
	const dot = $derived(Math.max(8, diameter * 0.28));
</script>

<span class="relative inline-flex shrink-0">
	<Avatar.Root
		style="width: {diameter}px; height: {diameter}px; box-shadow: {ring}"
		class="size-auto"
	>
		<Avatar.Fallback
			class="font-semibold tracking-[0.01em]"
			style="font-size: {Math.round(diameter * 0.38)}px; background: color-mix(in srgb, {hue} 22%, var(--ink-4)); color: color-mix(in srgb, {hue} 55%, var(--text-primary))"
		>
			{initialsFor(name)}
		</Avatar.Fallback>
	</Avatar.Root>

	{#if presence}
		<span
			class="absolute -right-px -bottom-px rounded-full"
			style="width: {dot}px; height: {dot}px; background: {PRESENCE_COLOUR[
				presence
			]}; box-shadow: 0 0 0 2px var(--ink-1)"
		></span>
	{/if}
</span>
