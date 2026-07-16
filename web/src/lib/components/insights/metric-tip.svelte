<script lang="ts">
	import InfoIcon from '@lucide/svelte/icons/info';
	import * as Tooltip from '$lib/components/ui/tooltip';

	let { def }: { def: string } = $props();
</script>

<Tooltip.Provider delayDuration={120}>
	<Tooltip.Root>
		<Tooltip.Trigger
			class="text-subtle-foreground hover:text-muted-foreground inline-flex cursor-help items-center transition-colors"
			aria-label="What this measures"
		>
			<InfoIcon class="size-3" />
		</Tooltip.Trigger>
		<Tooltip.Content
			side="top"
			arrowClasses="bg-transparent"
			class="bg-popover text-muted-foreground border-border block max-w-[220px] rounded-md border px-[11px] py-[9px] text-[11.5px] leading-[1.5] font-normal shadow-[var(--shadow-pop)]"
		>
			{def}
		</Tooltip.Content>
	</Tooltip.Root>
</Tooltip.Provider>
