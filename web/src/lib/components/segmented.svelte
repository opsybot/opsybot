<script lang="ts">
	let {
		label,
		options,
		current
	}: {
		label: string;
		options: { value: string; label: string; href: string }[];
		current: string;
	} = $props();
</script>

<div
	role="group"
	aria-label={label}
	class="bg-inset flex gap-0.5 rounded-full border p-0.5"
>
	{#each options as option (option.value)}
		{@const on = option.value === current}
		<a
			href={option.href}
			aria-current={on ? 'true' : undefined}
			class="rounded-full px-[11px] py-1 text-[11.5px] whitespace-nowrap transition-colors duration-[120ms] ease-out
			{on ? 'text-foreground bg-[var(--ink-5)]' : 'text-muted-foreground hover:text-foreground'}"
		>
			{option.label}
		</a>
	{/each}
</div>
