<script lang="ts">
	import XIcon from '@lucide/svelte/icons/x';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { ICON } from '$lib/components/escalation/icons';
	import { targetIcon, targetInvalid, type Target } from '$lib/escalation';

	let { target, onremove }: { target: Target; onremove?: () => void } = $props();

	const Icon = $derived(ICON[targetIcon(target.type)]);
	const bad = $derived(targetInvalid(target));
</script>

<div
	class="bg-inset flex items-center gap-2 rounded-md border px-2.5 py-1.5 {bad
		? 'border-critical-edge bg-critical-wash'
		: ''}"
>
	<Icon class="text-subtle-foreground size-[13px] shrink-0" />
	<span class="min-w-0 truncate text-[12.5px] font-medium">{target.value}</span>
	<Badge tone="neutral" size="sm">{target.type}</Badge>
	{#if bad}
		<Badge tone="critical" size="sm">can't page</Badge>
	{/if}
	{#if onremove}
		<div class="flex-1"></div>
		<Button
			variant="ghost"
			size="icon-sm"
			aria-label="Remove target"
			onclick={(event) => {
				event.stopPropagation();
				onremove?.();
			}}
		>
			<XIcon />
		</Button>
	{/if}
</div>
