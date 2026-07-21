<script lang="ts">
	import type { Component } from 'svelte';
	import type { LucideProps } from '@lucide/svelte';
	import CreditCardIcon from '@lucide/svelte/icons/credit-card';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import SmartphoneIcon from '@lucide/svelte/icons/smartphone';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { formatCount, usageFill, usagePercent } from '$lib/billing';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const USAGE_ICON: Record<string, Component<LucideProps>> = {
		'message-square': MessageSquareIcon,
		phone: PhoneIcon,
		smartphone: SmartphoneIcon
	};

	let profileForm: HTMLFormElement;
</script>

<div class="flex max-w-[760px] flex-col gap-3.5">
	<div class="grid grid-cols-1 items-start gap-3.5 md:grid-cols-2">
		<div class="bg-card flex flex-col gap-3 rounded-xl border p-4">
			<span class="text-[13.5px] font-semibold">Payment method</span>
			{#if data.payment}
				<div class="bg-inset flex items-center gap-3 rounded-md border px-3 py-2.5">
					<CreditCardIcon class="text-muted-foreground size-[18px]" />
					<div class="flex-1">
						<div class="text-[13px] font-medium">{data.payment.brand} ending {data.payment.last4}</div>
						<div class="text-subtle-foreground font-mono text-[11px]">expires {data.payment.expires}</div>
					</div>
					<Button size="sm" variant="ghost" onclick={() => toast('Card update opens a secure form.')}>Update</Button>
				</div>
			{:else}
				<div class="bg-inset flex items-center gap-3 rounded-md border border-dashed px-3 py-2.5">
					<CreditCardIcon class="text-subtle-foreground size-[18px]" />
					<span class="text-subtle-foreground flex-1 text-[12.5px]">No card yet. Add one before the trial ends.</span>
					<Button size="sm" variant="secondary" onclick={() => toast('Card setup opens a secure form.')}>Add card</Button>
				</div>
			{/if}
		</div>

		<div class="bg-card flex flex-col gap-2.5 rounded-xl border p-4">
			<span class="text-[13.5px] font-semibold">Company & VAT</span>
			<form
				bind:this={profileForm}
				method="POST"
				action="?/saveProfile"
				class="flex flex-col gap-2.5"
				use:enhance={() => async ({ result, update }) => {
					await update({ reset: false });
					if (result.type === 'success') toast.success('Billing details saved.');
				}}
			>
				<Field.Field class="gap-1.5 space-y-0">
					<Field.FieldLabel for="bill-company" class="text-muted-foreground text-[13px] font-medium">Company</Field.FieldLabel>
					<Input id="bill-company" name="company" value={data.profile.company} onchange={() => profileForm.requestSubmit()} />
				</Field.Field>
				<Field.Field class="gap-1.5 space-y-0">
					<Field.FieldLabel for="bill-vat" class="text-muted-foreground text-[13px] font-medium">VAT ID</Field.FieldLabel>
					<Input id="bill-vat" name="vat" class="font-mono text-[12px]" value={data.profile.vat} onchange={() => profileForm.requestSubmit()} />
				</Field.Field>
			</form>
		</div>
	</div>

	<div class="bg-card overflow-hidden rounded-xl border">
		<header class="flex items-center gap-2 border-b px-4 py-3">
			<span class="text-[13.5px] font-semibold">Usage this period</span>
			<span class="text-subtle-foreground ml-1 text-[11.5px]">managed delivery · resets 2026-08-01</span>
		</header>
		<div>
			{#each data.usage as meter (meter.kind)}
				{@const Icon = USAGE_ICON[meter.icon]}
				{@const pct = usagePercent(meter)}
				<div class="flex items-center gap-3 border-t px-4 py-[11px] first:border-t-0" data-usage={meter.kind}>
					<Icon class="text-subtle-foreground size-[14px] shrink-0" />
					<span class="w-[60px] text-[13px]">{meter.kind}</span>
					<div class="flex-1">
						{#if pct === null}
							<span class="text-subtle-foreground text-[12px]">included, unmetered</span>
						{:else}
							<div class="bg-inset h-1.5 overflow-hidden rounded-full">
								<span class="block h-full rounded-full" style="width:{pct}%; background:{usageFill(pct)}"></span>
							</div>
						{/if}
					</div>
					<span class="text-subtle-foreground w-[120px] text-right font-mono text-[12px]">
						{formatCount(meter.used)}{pct === null ? '' : ` / ${formatCount(meter.included as number)}`}
					</span>
				</div>
			{/each}
		</div>
		<div class="text-subtle-foreground border-t px-4 py-2.5 text-[11.5px] leading-[1.5]">
			Over the included volume, SMS is €0.04 and voice €0.09 each. You'll see a banner before any overage begins: nothing is charged silently.
		</div>
	</div>

	<div class="bg-card overflow-hidden rounded-xl border">
		<header class="flex items-center gap-2 border-b px-4 py-3">
			<span class="text-[13.5px] font-semibold">Invoices</span>
		</header>
		<div>
			{#each data.invoices as invoice (invoice.id)}
				<div class="flex items-center gap-3 border-t px-4 py-[11px] first:border-t-0" data-invoice={invoice.id}>
					<span class="text-foreground flex-1 font-mono text-[12.5px]">{invoice.id}</span>
					<span class="text-muted-foreground font-mono text-[12px]">{invoice.date}</span>
					<span class="w-[70px] text-right font-mono text-[12.5px]">{invoice.amount}</span>
					<Badge tone="success" size="sm">{invoice.status}</Badge>
					<Button
						size="sm"
						variant="ghost"
						onclick={() => toast.success(`${invoice.id}.pdf downloading.`)}
					>
						<DownloadIcon data-icon="inline-start" />
						PDF
					</Button>
				</div>
			{:else}
				<p class="text-subtle-foreground m-0 px-4 py-8 text-center text-[13px]">
					No invoices yet. Your first invoice arrives when the trial converts to a paid plan.
				</p>
			{/each}
		</div>
	</div>
</div>
