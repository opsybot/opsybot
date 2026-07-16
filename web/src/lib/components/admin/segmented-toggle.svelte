<script lang="ts">
	let {
		value = $bindable(),
		options,
		label,
		onchange
	}: {
		value: string;
		options: { value: string; label: string }[];
		label: string;
		onchange?: (value: string) => void;
	} = $props();

	function pick(next: string) {
		value = next;
		onchange?.(next);
	}
</script>

<div role="group" aria-label={label} class="bg-inset flex gap-0.5 rounded-full border p-0.5">
	{#each options as option (option.value)}
		{@const on = value === option.value}
		<button
			type="button"
			aria-pressed={on}
			onclick={() => pick(option.value)}
			class="rounded-full px-3 py-1 text-[11.5px] whitespace-nowrap transition-colors
			{on ? 'bg-[var(--ink-5)] text-foreground' : 'text-muted-foreground hover:text-foreground'}"
		>
			{option.label}
		</button>
	{/each}
</div>
