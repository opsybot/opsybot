<script lang="ts">
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';

	let { repeat, active = false }: { repeat: string; active?: boolean } = $props();

	const repeats = $derived(!!repeat && repeat !== '0');
</script>

<div
	class="border-warning-edge bg-warning-wash flex w-[288px] gap-[9px] rounded-xl border px-[13px] py-2.5 {active
		? 'shadow-[0_0_0_1px_var(--warning),var(--focus-ring)]'
		: ''}"
>
	<OctagonAlertIcon class="text-warning-ink mt-px size-3.5 shrink-0" />
	<div class="min-w-0">
		<div class="text-[12.5px] font-semibold">Escalation exhausted</div>
		<div class="text-muted-foreground mt-px text-[11.5px] leading-[1.5]">
			{repeats ? `Repeats ×${repeat} first, then the ` : 'The '}alert stays open, flagged exhausted.
			No one else is paged.
		</div>
	</div>
</div>
