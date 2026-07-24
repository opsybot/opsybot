<script lang="ts">
	import GlobeIcon from '@lucide/svelte/icons/globe';
	import Panel from '$lib/components/incidents/panel.svelte';
</script>

<div class="flex max-w-[720px] flex-col gap-3.5">
	<Panel class="flex flex-col items-center gap-2.5 px-5 py-12 text-center">
		<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
			<GlobeIcon class="text-subtle-foreground size-5" />
		</span>
		<div class="text-[15px] font-medium">Not published</div>
		<p class="text-subtle-foreground m-0 max-w-[420px] text-[13px] leading-[1.55]">
			This incident isn't published to a public status page.
		</p>
	</Panel>
</div>
