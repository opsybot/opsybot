<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import type { Mapping } from '$lib/alertsources';

	let {
		rows = $bindable(),
		editable = false,
		note,
		flush = false
	}: {
		rows: Mapping[];
		editable?: boolean;
		note?: string;
		flush?: boolean;
	} = $props();
</script>

<div class="bg-card overflow-hidden {flush ? '' : 'rounded-xl border'}">
	<div
		class="bg-inset text-subtle-foreground grid grid-cols-[120px_minmax(0,1fr)] items-center gap-3 px-[14px] py-[9px] text-[11px] tracking-[0.07em] uppercase"
	>
		<span>Opsybot field</span>
		<span>From payload</span>
	</div>
	{#each rows as row, index (index)}
		<div
			class="grid grid-cols-[120px_minmax(0,1fr)] items-center gap-3 border-t px-[14px] py-[9px]"
		>
			<span class="font-mono text-[12.5px]">{row.field}</span>
			{#if editable}
				<Input bind:value={rows[index].path} class="h-[34px] font-mono text-[12px]" aria-label="{row.field} path" />
			{:else}
				<span class="text-muted-foreground font-mono text-[12px]">{row.path}</span>
			{/if}
		</div>
	{/each}
	{#if note}
		<div class="text-subtle-foreground border-t px-3 py-[9px] text-[11.5px]">{note}</div>
	{/if}
</div>
