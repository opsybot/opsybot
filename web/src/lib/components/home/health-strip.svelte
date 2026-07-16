<script lang="ts">
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { Button } from '$lib/components/ui/button';
	import type { InstanceHealth } from '$lib/dashboard';
	import { ws } from '$lib/navigation';
	import { formatUtc } from '$lib/time';

	let { instance }: { instance: InstanceHealth } = $props();

	const unhealthy = $derived(instance.workersTotal - instance.workersHealthy);
</script>

<div
	role="status"
	class="bg-warning-wash border-warning-edge text-muted-foreground flex items-center gap-2.5 rounded-md border px-3.5 py-2.5 text-[12.5px] leading-normal"
>
	<TriangleAlertIcon class="text-warning size-[15px] shrink-0" />
	<span>
		Instance health: {unhealthy} of {instance.workersTotal} background workers unhealthy — page delivery
		may be delayed.
		<span class="text-subtle-foreground font-mono">
			Last check {formatUtc(instance.checkedAt)}
		</span>
	</span>
	<Button variant="ghost" size="sm" href={ws('/operations')} class="ml-auto shrink-0">
		View diagnostics
	</Button>
</div>
