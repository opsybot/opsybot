<script lang="ts">
	import ArrowUpRightIcon from '@lucide/svelte/icons/arrow-up-right';
	import BellIcon from '@lucide/svelte/icons/bell';
	import CheckIcon from '@lucide/svelte/icons/check';
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import InboxIcon from '@lucide/svelte/icons/inbox';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import SmartphoneIcon from '@lucide/svelte/icons/smartphone';
	import { Badge } from '$lib/components/ui/badge';
	import type { EscalationEvent, EscalationEventKind } from '$lib/alerts';
	import { formatUtcTime } from '$lib/time';

	let { events }: { events: EscalationEvent[] } = $props();

	const ICON: Record<EscalationEventKind, typeof BellIcon> = {
		received: InboxIcon,
		escalation: ArrowUpRightIcon,
		push: SmartphoneIcon,
		sms: MessageSquareIcon,
		timeout: ClockIcon,
		chat: MessageSquareIcon,
		acked: CheckIcon,
		resolved: CircleCheckIcon
	};

	const TONE = { success: 'success', warning: 'warning', brand: 'brand' } as const;

	const local = (at: string) =>
		new Date(at).toLocaleTimeString(undefined, {
			hour: '2-digit',
			minute: '2-digit',
			hour12: false
		});
</script>

<div class="px-[18px] pt-3.5 pb-1.5">
	{#each events as event, index (event.id)}
		{@const Icon = ICON[event.kind]}
		{@const last = index === events.length - 1}

		<div class="grid grid-cols-[118px_26px_minmax(0,1fr)] gap-x-2.5">
			<div class="flex flex-col gap-px pt-0.5 text-right">
				<span class="text-muted-foreground font-mono text-xs">
					{formatUtcTime(event.at).replace(' UTC', '')} UTC
				</span>
				<span class="text-subtle-foreground font-mono text-[10.5px]">{local(event.at)} local</span>
			</div>

			<div class="flex flex-col items-center">
				<span
					class="flex size-[22px] shrink-0 items-center justify-center rounded-full border {event.tone ===
					'brand'
						? 'bg-primary border-primary text-primary-foreground'
						: 'bg-popover text-muted-foreground'}"
				>
					<Icon class="size-[11px]" />
				</span>
				{#if !last}
					<span class="bg-border min-h-2.5 w-px flex-1"></span>
				{/if}
			</div>

			<div class="flex flex-wrap items-baseline gap-2 pt-[3px] pb-4">
				<span class="text-foreground text-[13px]">{event.text}</span>
				{#if event.result}
					<Badge tone={event.tone ? TONE[event.tone] : 'neutral'} size="sm">{event.result}</Badge>
				{/if}
			</div>
		</div>
	{/each}
</div>
