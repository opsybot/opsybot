<script lang="ts">
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import { cn } from '$lib/utils';

	let { text, class: className }: { text: string; class?: string } = $props();
</script>

<div class={cn('text-muted-foreground flex items-center gap-[9px] p-4 text-[13px]', className)}>
	<CircleCheckIcon class="text-success size-[15px] shrink-0" />
	<span>{text}</span>
</div>
