<script lang="ts">
	import { untrack } from 'svelte';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import MegaphoneIcon from '@lucide/svelte/icons/megaphone';
	import { Button } from '$lib/components/ui/button';
	import { formatAge } from '$lib/time';

	let {
		dueAt,
		now: serverNow
	}: {
		dueAt: string;
		now: number;
	} = $props();

	let now = $state(untrack(() => serverNow));

	$effect(() => {
		const timer = setInterval(() => (now = Date.now()), 1000);
		return () => clearInterval(timer);
	});

	const remaining = $derived(Date.parse(dueAt) - now);
	const overdue = $derived(remaining < 0);
	const label = $derived(formatAge(Math.abs(remaining)));
</script>

<div class="flex items-center gap-2">
	<ClockIcon class="size-[13px] shrink-0 {overdue ? 'text-critical-ink' : 'text-subtle-foreground'}" />
	<span class="font-mono text-xs {overdue ? 'text-critical-ink' : 'text-muted-foreground'}">
		{overdue ? `${label} overdue` : `update due in ${label}`}
	</span>
	{#if overdue}
		<Button type="submit" variant="secondary" size="sm">
			<MegaphoneIcon data-icon="inline-start" />
			Post update
		</Button>
	{/if}
</div>
