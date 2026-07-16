<script lang="ts">
	import ArrowRightLeftIcon from '@lucide/svelte/icons/arrow-right-left';
	import { formatWhen, type Handover } from '$lib/oncall';

	let { handovers, now }: { handovers: Handover[]; now: number } = $props();
</script>

<section class="bg-card overflow-hidden rounded-xl border">
	<header class="flex items-center gap-2.5 border-b px-4 py-3">
		<span class="text-[13.5px] font-semibold">Upcoming handovers</span>
	</header>

	{#each handovers as handover (handover.at)}
		<div class="flex items-center gap-2.5 border-t px-3.5 py-2.5 first:border-t-0">
			<ArrowRightLeftIcon class="text-subtle-foreground size-[13px] shrink-0" />
			<div class="min-w-0 flex-1">
				<div class="text-[12.5px]">
					{handover.from.split(' ')[0]} → <strong>{handover.to.split(' ')[0]}</strong>
				</div>
				<div class="text-subtle-foreground mt-px font-mono text-[10.5px]">
					{formatWhen(handover.at, now)}
				</div>
			</div>
		</div>
	{:else}
		<p class="text-subtle-foreground m-0 px-3.5 py-3 text-[12.5px]">
			No handovers in the next two weeks.
		</p>
	{/each}
</section>
