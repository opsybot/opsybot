<script lang="ts">
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<div class="bg-background min-h-svh auth-glow">
	{@render children()}
</div>

<style>
	.auth-glow {
		background-image: radial-gradient(
			720px 460px at 50% 0%,
			rgb(0 229 172 / 0.055),
			transparent 70%
		);
	}
</style>
