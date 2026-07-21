<script lang="ts">
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';
	import RotateCwIcon from '@lucide/svelte/icons/rotate-cw';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import { formatUtc } from '$lib/time';
	import type { WorkflowRun } from '$lib/workflows';

	let { workflowId, run }: { workflowId: string; run: WorkflowRun } = $props();
</script>

<div class="flex items-start gap-2.5 border-t px-0.5 py-[9px] first:border-t-0">
	{#if run.ok}
		<CircleCheckIcon class="text-success mt-0.5 size-3.5 shrink-0" />
	{:else}
		<OctagonAlertIcon class="text-critical mt-0.5 size-3.5 shrink-0" />
	{/if}

	<div class="min-w-0 flex-1">
		<div class="text-muted-foreground text-[12.5px] leading-[1.5]">{run.summary}</div>
		<div class="text-subtle-foreground mt-0.5 font-mono text-[10.5px]">
			{formatUtc(run.at)} · fired by {run.incident}{run.error ? ` · ${run.error}` : ''}
		</div>
	</div>

	{#if run.retriable}
		<form
			method="POST"
			action="?/retry"
			use:enhance={() =>
				async ({ result, update }) => {
					await update();
					if (result.type === 'success') toast.success('Retried: the webhook returned 200.');
				}}
		>
			<input type="hidden" name="workflow" value={workflowId} />
			<input type="hidden" name="run" value={run.id} />
			<Button type="submit" size="sm" variant="ghost">
				<RotateCwIcon data-icon="inline-start" />
				Retry
			</Button>
		</form>
	{/if}
</div>
