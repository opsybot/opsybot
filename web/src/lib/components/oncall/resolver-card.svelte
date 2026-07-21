<script lang="ts">
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { onMount, tick } from 'svelte';
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import DatePicker from '$lib/components/date-picker.svelte';
	import TimeSelect from '$lib/components/time-select.svelte';
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
		keep: Record<string, string>;
	} = $props();

	let enhanced = $state(false);
	let form = $state<HTMLFormElement | null>(null);

	onMount(() => (enhanced = true));

	async function submitLookup() {
		await tick();
		form?.requestSubmit();
	}
</script>

<section class="bg-card overflow-hidden rounded-xl border">
	<header class="flex items-center gap-2.5 border-b px-4 py-3">
		<span class="text-[13.5px] font-semibold">Who is on call at…</span>
	</header>

	<form method="GET" class="px-3.5 py-3" bind:this={form}>
		{#each Object.entries(keep) as [name, value] (name)}
			<input type="hidden" {name} {value} />
		{/each}

		<div class="flex gap-2">
			{#if enhanced}
				<div class="flex-1">
					<DatePicker
						name="date"
						label="Date"
						size="sm"
						value={date}
						onChange={submitLookup}
					/>
				</div>
				<TimeSelect
					name="time"
					label="Time, UTC"
					size="sm"
					value={time}
					onChange={submitLookup}
					class="w-[110px]"
				/>
			{:else}
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
			{/if}
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
					No one is on call then. That is a gap.
				</span>
			{/if}
		</div>
	</form>
</section>
