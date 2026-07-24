<script lang="ts">
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import Panel from '$lib/components/incidents/panel.svelte';
</script>

<div class="flex max-w-[720px] flex-col gap-3.5">
	<Panel class="flex flex-col items-center gap-2.5 px-5 py-12 text-center">
		<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
			<FileTextIcon class="text-subtle-foreground size-5" />
		</span>
		<div class="text-[15px] font-medium">Postmortem not started</div>
		<p class="text-subtle-foreground m-0 max-w-[420px] text-[13px] leading-[1.55]">
			Once this incident is resolved, its postmortem will be drafted and tracked here.
		</p>
	</Panel>
</div>
