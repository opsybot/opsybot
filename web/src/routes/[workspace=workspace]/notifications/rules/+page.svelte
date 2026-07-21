<script lang="ts">
	import { untrack } from 'svelte';
	import BellIcon from '@lucide/svelte/icons/bell';
	import CheckIcon from '@lucide/svelte/icons/check';
	import QuoteIcon from '@lucide/svelte/icons/quote';
	import SirenIcon from '@lucide/svelte/icons/siren';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import QuietHours from '$lib/components/notifications/quiet-hours.svelte';
	import StepsBuilder from '$lib/components/notifications/steps-builder.svelte';
	import { previewSentence } from '$lib/notifications';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let high = $state(untrack(() => data.high.map((step) => ({ ...step }))));
	let low = $state(untrack(() => data.low.map((step) => ({ ...step }))));
	let quiet = $state(untrack(() => structuredClone(data.quietHours)));

	const rulesJson = $derived(JSON.stringify({ high, low, quietHours: quiet }));
</script>

<div class="flex max-w-[720px] flex-col gap-3.5">
	<div class="bg-card overflow-hidden rounded-xl border">
		<header class="flex items-center gap-2 border-b px-4 py-3">
			<SirenIcon class="text-subtle-foreground size-3.5" />
			<span class="text-[13.5px] font-semibold">High urgency: pages</span>
			<span class="text-subtle-foreground text-[12px]">SEV1/SEV2 and anything that demands action</span>
		</header>
		<div class="flex flex-col gap-3 p-3.5">
			<StepsBuilder bind:steps={high} label="High urgency" />
			<div
				class="bg-brand-wash border-brand-edge text-muted-foreground flex items-start gap-2 rounded-md border px-3 py-2.5 text-[13px] leading-[1.55]"
			>
				<QuoteIcon class="text-brand-foreground mt-0.5 size-3 shrink-0" />
				<span>{previewSentence(high)}</span>
			</div>
		</div>
	</div>

	<div class="bg-card overflow-hidden rounded-xl border">
		<header class="flex items-center gap-2 border-b px-4 py-3">
			<BellIcon class="text-subtle-foreground size-3.5" />
			<span class="text-[13.5px] font-semibold">Low urgency</span>
			<span class="text-subtle-foreground text-[12px]">FYIs, digests, non-paging alerts</span>
		</header>
		<div class="flex flex-col gap-3 p-3.5">
			<StepsBuilder bind:steps={low} label="Low urgency" />
			<div
				class="bg-brand-wash border-brand-edge text-muted-foreground flex items-start gap-2 rounded-md border px-3 py-2.5 text-[13px] leading-[1.55]"
			>
				<QuoteIcon class="text-brand-foreground mt-0.5 size-3 shrink-0" />
				<span>{previewSentence(low)}</span>
			</div>
		</div>
	</div>

	<QuietHours bind:value={quiet} />

	<form
		method="POST"
		action="?/save"
		class="self-start"
		use:enhance={() => async ({ result, update }) => {
			await update({ reset: false });
			if (result.type === 'success') toast.success('Notification rules saved. They apply to the next page.');
		}}
	>
		<input type="hidden" name="rules" value={rulesJson} />
		<Button type="submit">
			<CheckIcon data-icon="inline-start" />
			Save rules
		</Button>
	</form>
</div>
