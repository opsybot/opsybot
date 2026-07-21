<script lang="ts">
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import type { Scope } from '$lib/chat';

	let { scopes }: { scopes: Scope[] } = $props();
</script>

<div class="flex flex-col gap-2">
	{#each scopes as scope (scope.what)}
		<div class="flex items-start gap-[9px]">
			<ShieldCheckIcon class="text-primary mt-0.5 size-3.5 shrink-0" />
			<span class="text-muted-foreground text-[12.5px] leading-[1.5]">
				<strong class="text-foreground font-semibold">{scope.what}</strong>: {scope.why}
			</span>
		</div>
	{/each}
</div>
