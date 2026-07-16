<script lang="ts">
	import { cn } from '$lib/utils';

	let { class: className }: { class?: string } = $props();
</script>

<span
	class={cn(
		'text-foreground text-[18px] leading-none font-bold tracking-[-0.02em] select-none',
		className
	)}
>
	opsy<span class="text-primary">.</span>bot
</span>
