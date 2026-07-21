<script lang="ts">
	import type { Component } from 'svelte';
	import type { LucideProps } from '@lucide/svelte';
	import ArrowUpRightIcon from '@lucide/svelte/icons/arrow-up-right';
	import CheckIcon from '@lucide/svelte/icons/check';
	import CircleOffIcon from '@lucide/svelte/icons/circle-off';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import { NEXT_INVOICE_DATE, PLANS, getPlan, planCta, proration, statusBanner, type Plan } from '$lib/billing';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const banner = $derived(statusBanner(data.status, data.trialDaysLeft));
	const BANNER_ICON: Record<string, Component<LucideProps>> = {
		clock: ClockIcon,
		'triangle-alert': TriangleAlertIcon,
		'circle-off': CircleOffIcon
	};
	const bannerSurface = (tone: string) =>
		tone === 'warning' ? 'bg-warning-wash border-warning-edge' : 'bg-brand-wash border-brand-edge';
	const bannerInk = (tone: string) => (tone === 'warning' ? 'text-warning-ink' : 'text-brand-foreground');

	let changeTo = $state<Plan | null>(null);
	let changeForm: HTMLFormElement;
</script>

<div class="flex max-w-[900px] flex-col gap-4">
	{#if banner}
		{@const Icon = BANNER_ICON[banner.icon]}
		<div class="flex items-center gap-3 rounded-xl border px-4 py-3 text-[12.5px] leading-[1.5] {bannerSurface(banner.tone)}">
			<Icon class="size-[15px] shrink-0 {bannerInk(banner.tone)}" />
			<span class="text-muted-foreground flex-1">
				<strong class="text-foreground font-semibold">{banner.title}</strong>: {banner.body}
			</span>
			{#if banner.cta}
				<Button size="sm" onclick={() => (changeTo = getPlan(data.currentPlanId) ?? null)}>{banner.cta}</Button>
			{/if}
		</div>
	{/if}

	<div class="grid grid-cols-1 gap-3 md:grid-cols-3">
		{#each PLANS as plan (plan.id)}
			{@const cta = planCta(plan, data.currentPlanId)}
			{@const current = plan.id === data.currentPlanId}
			<div
				data-plan={plan.id}
				class="bg-card flex flex-col gap-2 rounded-xl border p-[18px] {current
					? 'border-brand-edge shadow-[var(--glow-brand)]'
					: ''}"
			>
				{#if current}
					<Badge tone="brand" size="sm" class="self-start">current plan</Badge>
				{:else}
					<span class="h-[18px]"></span>
				{/if}
				<div class="text-[15px] font-semibold">{plan.name}</div>
				<div class="flex items-baseline gap-[3px]">
					<span class="text-[30px] font-light tracking-[-0.02em]">{plan.price}</span>
					<span class="text-subtle-foreground text-[13px]">{plan.period}</span>
				</div>
				<div class="text-subtle-foreground min-h-8 text-[12px] leading-[1.45]">{plan.note}</div>
				<div class="mt-1 flex flex-1 flex-col gap-[7px]">
					{#each plan.caps as cap (cap)}
						<span class="text-muted-foreground flex items-center gap-2 text-[12.5px]">
							<CheckIcon class="size-[13px] shrink-0 text-[var(--mint-500)]" />
							{cap}
						</span>
					{/each}
				</div>
				{#if current}
					<Button variant="secondary" class="w-full" disabled>Current plan</Button>
				{:else}
					<Button variant={cta.primary ? 'default' : 'secondary'} class="w-full" onclick={() => (changeTo = plan)}>
						{cta.label}
					</Button>
				{/if}
			</div>
		{/each}
	</div>

	<p class="text-subtle-foreground m-0 text-[12px] leading-[1.5]">
		Every paid plan includes <strong class="font-semibold">unlimited responders</strong>. You're never billed per
		seat. Managed delivery (SMS/voice) is metered separately with volumes shown before any overage applies.
	</p>
</div>

<Dialog.Root open={!!changeTo} onOpenChange={(value) => (value ? undefined : (changeTo = null))}>
	<Dialog.Content class="sm:max-w-[460px]">
		{#if changeTo}
			{@const pr = proration(changeTo.id)}
			<div class="flex flex-col gap-4 p-6">
				<div class="flex items-start gap-3">
					<span class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg">
						<ArrowUpRightIcon class="size-5" />
					</span>
					<Dialog.Title class="tracking-heading text-xl font-semibold">Change to {changeTo.name}</Dialog.Title>
				</div>
				<div class="flex flex-col gap-3">
					<div class="bg-inset overflow-hidden rounded-md border">
						<div class="flex justify-between gap-3 px-3.5 py-[9px] text-[12.5px]">
							<span class="text-muted-foreground">New plan</span>
							<span class="text-foreground font-mono">{changeTo.name} · {changeTo.price}{changeTo.period}</span>
						</div>
						<div class="flex justify-between gap-3 border-t px-3.5 py-[9px] text-[12.5px]">
							<span class="text-muted-foreground">Prorated now (19 days left)</span>
							<span class="text-foreground font-mono">{pr.proratedNow}</span>
						</div>
						<div class="flex justify-between gap-3 border-t px-3.5 py-[9px] text-[12.5px] font-semibold">
							<span class="text-muted-foreground">Next invoice · {NEXT_INVOICE_DATE}</span>
							<span class="text-foreground font-mono">{changeTo.price}{changeTo.period}</span>
						</div>
					</div>
					<p class="text-subtle-foreground m-0 text-[12px] leading-[1.5]">
						Change takes effect immediately. Nothing is deleted on a downgrade: over-limit items become read-only.
					</p>
				</div>
			</div>
			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button variant="ghost" onclick={() => (changeTo = null)}>Cancel</Button>
				<Button onclick={() => changeForm.requestSubmit()}>Confirm change</Button>
			</div>
		{/if}
	</Dialog.Content>
</Dialog.Root>

<form
	bind:this={changeForm}
	method="POST"
	action="?/change"
	class="hidden"
	use:enhance={() => async ({ result, update }) => {
		if (result.type === 'failure') {
			toast.error(String(result.data?.error ?? 'Could not change the plan.'));
			return;
		}
		if (result.type !== 'success') return;
		const name = changeTo?.name;
		await update({ reset: false });
		changeTo = null;
		if (name) toast.success(`Plan changed to ${name}. Prorated today.`);
	}}
>
	<input type="hidden" name="plan" value={changeTo?.id ?? ''} />
</form>
