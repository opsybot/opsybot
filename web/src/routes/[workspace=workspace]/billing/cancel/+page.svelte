<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import BracesIcon from '@lucide/svelte/icons/braces';
	import CheckIcon from '@lucide/svelte/icons/check';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Field from '$lib/components/ui/field';
	import { Textarea } from '$lib/components/ui/textarea';
	import { ws } from '$lib/navigation';

	const STEPS = ['Reason', 'Export', 'Confirm', 'Done'];
	const TIMELINE: [string, string][] = [
		['Today', 'Billing stops. The workspace stays fully usable until the period ends.'],
		['2026-08-01', 'Workspace goes read-only. Paging is disabled: move your on-call elsewhere first.'],
		['2026-08-31', 'Data is permanently deleted. Recoverable until this date by resubscribing.']
	];

	let step = $state(0);
	let reason = $state('');
	let cancelForm: HTMLFormElement;
	const back = () => goto(ws('/billing/account'));
</script>

<div class="flex max-w-[620px] flex-col gap-4">
	<button type="button" class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[12.5px] transition-colors" onclick={back}>
		<ArrowLeftIcon class="size-3.5" />
		Billing
	</button>

	<div class="flex items-center gap-2" data-step={step} aria-label="Cancellation progress">
		{#each STEPS as label, i (label)}
			{#if i > 0}
				<span class="h-px min-w-[16px] flex-1 {i <= step ? 'bg-[var(--mint-500)]' : 'bg-border'}"></span>
			{/if}
			<span
				class="inline-flex items-center gap-[7px] text-[12px] {i === step
					? 'text-foreground font-semibold'
					: i < step
						? 'text-muted-foreground'
						: 'text-subtle-foreground'}"
			>
				<span
					class="flex size-5 shrink-0 items-center justify-center rounded-full font-mono text-[10.5px] {i === step
						? 'bg-brand-wash border-brand-edge text-brand-foreground border'
						: i < step
							? 'border-[var(--mint-500)] bg-[var(--mint-500)] text-[var(--text-inverse)] border'
							: 'bg-inset text-subtle-foreground border'}"
				>
					{#if i < step}<CheckIcon class="size-[11px]" />{:else}{i + 1}{/if}
				</span>
				{label}
			</span>
		{/each}
	</div>

	{#if step === 0}
		<div class="bg-card flex flex-col gap-3.5 rounded-xl border p-[18px]">
			<h2 class="m-0 text-[17px] font-semibold">Sorry to see you go</h2>
			<Field.Field class="gap-1.5 space-y-0">
				<Field.FieldLabel for="cancel-reason" class="text-muted-foreground text-[13px] font-medium">
					What's prompting the cancellation? (optional)
				</Field.FieldLabel>
				<Textarea id="cancel-reason" rows={3} bind:value={reason} placeholder="Helps us fix what's not working." />
			</Field.Field>
			<div class="flex gap-2.5">
				<Button onclick={() => (step = 1)}>Continue</Button>
				<Button variant="ghost" onclick={back}>Keep my plan</Button>
			</div>
		</div>
	{:else if step === 1}
		<div class="bg-card flex flex-col gap-3.5 rounded-xl border p-[18px]">
			<h2 class="m-0 text-[17px] font-semibold">Take your data with you</h2>
			<p class="text-muted-foreground m-0 text-[13px] leading-[1.6]">
				Download everything before anything is scheduled for deletion. Both exports work regardless of plan.
			</p>
			<div class="flex gap-2.5">
				<Button variant="secondary" onclick={() => toast.success('Config export (YAML) downloading.')}>
					<BracesIcon data-icon="inline-start" />
					Config as code
				</Button>
				<Button
					variant="secondary"
					onclick={() => toast.success('Full export (incidents, timelines, postmortems) downloading.')}
				>
					<DownloadIcon data-icon="inline-start" />
					Full data export
				</Button>
			</div>
			<div class="mt-1 flex gap-2.5">
				<Button onclick={() => (step = 2)}>Continue</Button>
				<Button variant="ghost" onclick={() => (step = 0)}>Back</Button>
			</div>
		</div>
	{:else if step === 2}
		<div class="bg-card flex flex-col gap-3.5 rounded-xl border p-[18px]">
			<h2 class="m-0 text-[17px] font-semibold">What happens next</h2>
			<div class="flex flex-col gap-2.5">
				{#each TIMELINE as [when, what] (when)}
					<div class="flex gap-3">
						<span class="text-brand-foreground w-[90px] shrink-0 font-mono text-[12px]">{when}</span>
						<span class="text-muted-foreground text-[12.5px] leading-[1.5]">{what}</span>
					</div>
				{/each}
			</div>
			<Alert.Root tone="warning">
				<TriangleAlertIcon />
				<Alert.Content>
					<Alert.Description>
						Paging is disabled the moment the workspace goes read-only. Make sure another workspace or tool covers
						on-call before then.
					</Alert.Description>
				</Alert.Content>
			</Alert.Root>
			<div class="flex gap-2.5">
				<Button variant="destructive" onclick={() => cancelForm.requestSubmit()}>Cancel my plan</Button>
				<Button variant="ghost" onclick={() => (step = 1)}>Back</Button>
			</div>
		</div>
	{:else}
		<div class="bg-card flex flex-col items-center gap-3 rounded-xl border p-8 text-center" role="status">
			<span class="flex size-11 items-center justify-center rounded-full bg-[var(--mint-500)] shadow-[var(--glow-brand)]">
				<CheckIcon class="size-5 text-[var(--text-inverse)]" />
			</span>
			<h2 class="m-0 text-[18px] font-semibold">Plan cancelled</h2>
			<p class="text-muted-foreground m-0 max-w-[380px] text-[13px] leading-[1.6]">
				Billing has stopped. The workspace stays usable until 2026-08-01. Resubscribe any time before 2026-08-31 to
				keep your data.
			</p>
			<Button variant="secondary" onclick={back}>Back to billing</Button>
		</div>
	{/if}
</div>

<form
	bind:this={cancelForm}
	method="POST"
	action="?/cancel"
	class="hidden"
	use:enhance={() => async ({ result, update }) => {
		if (result.type === 'failure') {
			toast.error(String(result.data?.error ?? 'Could not cancel the plan.'));
			return;
		}
		if (result.type !== 'success') return;
		await update({ reset: false });
		step = 3;
	}}
>
	<input type="hidden" name="reason" value={reason} />
</form>
