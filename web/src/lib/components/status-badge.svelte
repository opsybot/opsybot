<script lang="ts" module>
	export type Status =
		| 'declared'
		| 'firing'
		| 'investigating'
		| 'identified'
		| 'monitoring'
		| 'acknowledged'
		| 'resolved'
		| 'closed'
		| 'operational'
		| 'degraded'
		| 'outage'
		| 'maintenance';

	const STATUS: Record<Status, { label: string; tone: string; pulse: boolean }> = {
		declared: { label: 'Declared', tone: 'critical', pulse: true },
		firing: { label: 'Triggered', tone: 'critical', pulse: true },
		investigating: { label: 'Investigating', tone: 'high', pulse: true },
		identified: { label: 'Identified', tone: 'warning', pulse: false },
		monitoring: { label: 'Monitoring', tone: 'info', pulse: false },
		acknowledged: { label: 'Acknowledged', tone: 'warning', pulse: false },
		resolved: { label: 'Resolved', tone: 'success', pulse: false },
		closed: { label: 'Closed', tone: 'neutral', pulse: false },
		operational: { label: 'Operational', tone: 'success', pulse: false },
		degraded: { label: 'Degraded', tone: 'warning', pulse: true },
		outage: { label: 'Major outage', tone: 'critical', pulse: true },
		maintenance: { label: 'Maintenance', tone: 'info', pulse: false }
	};
</script>

<script lang="ts">
	import { cn } from '$lib/utils';

	let {
		status,
		label,
		size = 'md',
		class: className
	}: {
		status: Status;
		label?: string;
		size?: 'sm' | 'md';
		class?: string;
	} = $props();

	const config = $derived(STATUS[status]);
	const dot = $derived(size === 'sm' ? 'size-1.5' : 'size-[7px]');
</script>

<span
	class={cn(
		'inline-flex items-center gap-[7px] rounded-full border bg-white/3 leading-none font-semibold whitespace-nowrap',
		size === 'sm' ? 'h-5 px-2 text-2xs' : 'h-6 px-2.5 text-xs',
		className
	)}
	style="--tone: var(--{config.tone}); color: var(--tone); border-color: color-mix(in srgb, var(--tone) 40%, transparent)"
>
	<span class="relative shrink-0 {dot}">
		{#if config.pulse}
			<span
				class="motion-safe:animate-status-ping absolute inset-0 rounded-full"
				style="background: var(--tone)"
			></span>
		{/if}
		<span class="absolute inset-0 rounded-full" style="background: var(--tone)"></span>
	</span>
	{label ?? config.label}
</span>
