<script lang="ts">
	import { ICON } from '$lib/components/escalation/icons';
	import { TONE_STYLE, type Tone } from '$lib/escalation';

	let { meta }: { meta: { label: string; tone: Tone; icon: string } } = $props();

	const Icon = $derived(ICON[meta.icon]);
	const style = $derived(TONE_STYLE[meta.tone]);
</script>

<div
	class="inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-[11px] font-semibold"
	style="background:{style.bg};border-color:{style.border};color:{style.color}"
>
	<Icon class="size-[11px]" />
	{meta.label}
</div>
