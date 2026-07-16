<script lang="ts">
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import { Input } from '$lib/components/ui/input';
	import type { Coverage } from '$lib/oncall';

	let {
		date,
		time,
		coverage,
		keep
	}: {
		date: string;
		time: string;
		coverage: Coverage;
		// Kept as hidden inputs; a GET submit replaces the whole query string
		keep: Record<string, string>;
	} = $props();
</script>

<section class="bg-card overflow-hidden rounded-xl border">
	<header class="flex items-center gap-2.5 border-b px-4 py-3">
		<span class="text-[13.5px] font-semibold">Who is on call at…</span>
	</header>

	<form method="GET" class="px-3.5 py-3">
		{#each Object.entries(keep) as [name, value] (name)}
			<input type="hidden" {name} {value} />
		{/each}

		<div class="flex gap-2">
			<Input
				type="date"
				name="date"
				value={date}
				aria-label="Date"
				onchange={(event: Event) => (event.target as HTMLInputElement).form?.requestSubmit()}
				class="h-[34px] flex-1 text-[13px]"
			/>
			<Input
				type="time"
				name="time"
				value={time}
				aria-label="Time, UTC"
				onchange={(event: Event) => (event.target as HTMLInputElement).form?.requestSubmit()}
				class="h-[34px] w-[100px] text-[13px]"
			/>
			<span class="text-subtle-foreground self-center font-mono text-[11px]">UTC</span>
		</div>

		<noscript>
			<button
				type="submit"
				class="text-muted-foreground hover:text-brand-foreground mt-2 text-[12.5px]"
			>
				Look up
			</button>
		</noscript>

		<div class="mt-2.5 flex items-center gap-2">
			{#if coverage.person}
				<UserAvatar name={coverage.person} size="xs" onCall />
				<span class="text-[13px] font-medium">{coverage.person}</span>
				<span class="text-subtle-foreground font-mono text-[11px]">via {coverage.via}</span>
			{:else}
				<TriangleAlertIcon class="text-warning-ink size-3.5 shrink-0" />
				<span class="text-warning-ink text-[12.5px]">
					No one is on call then — that is a gap.
				</span>
			{/if}
		</div>
	</form>
</section>
