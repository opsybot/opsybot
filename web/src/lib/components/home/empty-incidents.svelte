<script lang="ts">
	import SirenIcon from '@lucide/svelte/icons/siren';
	import { Button } from '$lib/components/ui/button';
</script>

<div
	class="text-muted-foreground flex flex-col items-center gap-2.5 rounded-xl border border-dashed px-5 py-14"
>
	<span
		class="bg-inset flex size-[42px] items-center justify-center rounded-full border"
	>
		<SirenIcon class="text-subtle-foreground size-5" />
	</span>
	<div class="text-[15px] font-medium">No incidents</div>
	<p class="text-subtle-foreground m-0 max-w-[380px] text-center text-[13px] leading-relaxed">
		Once an alert source is connected, anything that pages someone shows up here with its timeline.
	</p>
	<Button variant="ghost" size="sm">
		<SirenIcon data-icon="inline-start" />
		Declare an incident manually
	</Button>
</div>
