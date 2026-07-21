<script lang="ts">
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import LinkIcon from '@lucide/svelte/icons/link';
	import SparklesIcon from '@lucide/svelte/icons/sparkles';
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';
	import SurfaceOff from '$lib/components/ai/surface-off.svelte';

	let { enabled }: { enabled: boolean } = $props();
</script>

{#if !enabled}
	<SurfaceOff>Related-incident hints need a model: currently unavailable.</SurfaceOff>
{:else}
	<div class="bg-card rounded-xl border px-4 py-3.5">
		<header class="mb-1.5 flex items-center gap-2">
			<SparklesIcon class="text-primary size-3.5" />
			<span class="text-[13.5px] font-semibold">This looks like INC-2478</span>
		</header>
		<p class="text-muted-foreground mb-2.5 text-[12.5px] leading-[1.55]">
			Same service (<span class="text-foreground font-mono text-[12px]">payments-api</span>), same symptom
			(checkout errors after a deploy), 24 days ago. That one was a routing regression rolled back in 22
			minutes.
		</p>
		<div class="flex gap-2">
			<Button
				size="sm"
				variant="secondary"
				onclick={() => toast.success('Linked as "related to INC-2478".')}
			>
				<LinkIcon data-icon="inline-start" />
				Link it
			</Button>
			<Button
				size="sm"
				variant="ghost"
				onclick={() => toast('Opens the INC-2478 postmortem.')}
			>
				<FileTextIcon data-icon="inline-start" />
				Read its postmortem
			</Button>
		</div>
	</div>
{/if}
