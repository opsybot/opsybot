<script lang="ts">
	import SurfaceCorrelation from '$lib/components/ai/surface-correlation.svelte';
	import SurfaceDigest from '$lib/components/ai/surface-digest.svelte';
	import SurfaceRelated from '$lib/components/ai/surface-related.svelte';
	import SurfaceSummary from '$lib/components/ai/surface-summary.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

{#snippet label(title: string, where: string)}
	<div class="text-muted-foreground mb-2 text-[11px] font-semibold tracking-[0.07em] uppercase">
		{title}
		<span class="text-subtle-foreground font-normal tracking-normal normal-case">: {where}</span>
	</div>
{/snippet}

<div class="flex max-w-[760px] flex-col gap-[18px]">
	<p class="text-subtle-foreground m-0 text-[13px] leading-[1.6]">
		These components live inside other screens: the incident detail, the alert list, the declare
		dialog. Shown here in both states: flip the global AI switch on the Models tab to see every
		surface's "model unavailable" state.
	</p>

	<div>
		{@render label('Incident summary drawer', 'on the incident detail')}
		<SurfaceSummary enabled={data.enabled} />
	</div>
	<div>
		{@render label('Alert digest panel', 'on the alert list')}
		<SurfaceDigest enabled={data.enabled} />
	</div>
	<div>
		{@render label('Correlation suggestion', 'on the alert list and incident detail')}
		<SurfaceCorrelation enabled={data.enabled} />
	</div>
	<div>
		{@render label('Related-incident hint', 'in declare and on the incident detail')}
		<SurfaceRelated enabled={data.enabled} />
	</div>
</div>
