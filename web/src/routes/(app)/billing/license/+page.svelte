<script lang="ts">
	import type { Component } from 'svelte';
	import type { LucideProps } from '@lucide/svelte';
	import InfoIcon from '@lucide/svelte/icons/info';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import * as Alert from '$lib/components/ui/alert';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { licenseAlert, licenseBadge } from '$lib/billing';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const badge = $derived(licenseBadge(data.license.status));
	const alert = $derived(licenseAlert(data.license));
	const ALERT_ICON: Record<string, Component<LucideProps>> = {
		critical: OctagonAlertIcon,
		warning: TriangleAlertIcon,
		info: InfoIcon
	};

	const cells = $derived([
		{ k: 'Plan', v: data.license.plan, mono: false },
		{ k: 'Capacity', v: data.license.capacity, mono: false },
		{ k: 'Licensee', v: data.license.licensee, mono: false },
		{ k: 'Issued', v: data.license.issued, mono: true },
		{ k: 'Expires', v: data.license.expires, mono: true }
	]);

	let key = $state('');
	let phase = $state<'idle' | 'validating'>('idle');
	let error = $state('');
</script>

<div class="flex max-w-[720px] flex-col gap-3.5">
	{#if alert}
		{@const Icon = ALERT_ICON[alert.tone]}
		<Alert.Root tone={alert.tone}>
			<Icon />
			<Alert.Content>
				<Alert.Title>{alert.title}</Alert.Title>
				<Alert.Description>{alert.body}</Alert.Description>
			</Alert.Content>
		</Alert.Root>
	{/if}

	<div class="bg-card overflow-hidden rounded-xl border">
		<header class="flex items-center gap-2 border-b px-4 py-3">
			<span class="text-[13.5px] font-semibold">Current license</span>
			<Badge tone={badge.tone} size="sm" dot class="ml-1">{badge.label}</Badge>
		</header>
		<div class="grid [grid-template-columns:repeat(auto-fit,minmax(150px,1fr))]">
			{#each cells as cell (cell.k)}
				<div class="border-r border-b px-4 py-3">
					<div class="text-subtle-foreground mb-1 text-[10.5px] tracking-[0.07em] uppercase">{cell.k}</div>
					<div class="text-[13px] font-medium {cell.mono ? 'font-mono' : ''}">{cell.v}</div>
				</div>
			{/each}
		</div>
	</div>

	<div class="bg-card flex flex-col gap-3 rounded-xl border p-4">
		<span class="text-[13.5px] font-semibold">Activate a new key</span>
		<form
			method="POST"
			action="?/activate"
			class="flex flex-col gap-3"
			use:enhance={() => {
				phase = 'validating';
				error = '';
				return async ({ result, update }) => {
					if (result.type === 'failure') {
						error = String(result.data?.error ?? 'That key is not valid.');
						phase = 'idle';
						return;
					}
					if (result.type !== 'success') {
						phase = 'idle';
						return;
					}
					await update({ reset: false });
					key = '';
					phase = 'idle';
					toast.success('License activated — Business, unlimited responders, expires 2027-01-14.');
				};
			}}
		>
			<Field.Field class="gap-1.5 space-y-0">
				<Field.FieldLabel for="license-key" class="text-muted-foreground text-[13px] font-medium">License key</Field.FieldLabel>
				<Input
					id="license-key"
					name="key"
					class="font-mono"
					placeholder="OPSY-XXXX-XXXX-XXXX-XXXX"
					bind:value={key}
					aria-invalid={error ? 'true' : undefined}
					aria-describedby={error ? 'license-key-error' : undefined}
					oninput={() => (error = '')}
				/>
				{#if error}<Field.FieldError id="license-key-error" class="text-critical-ink text-xs">{error}</Field.FieldError>{/if}
			</Field.Field>
			<div class="flex items-center gap-2.5">
				<Button type="submit" disabled={!key.trim() || phase === 'validating'}>
					<KeyRoundIcon data-icon="inline-start" />
					{phase === 'validating' ? 'Validating…' : 'Activate license'}
				</Button>
				<button
					type="button"
					class="text-brand-foreground text-[12.5px] hover:underline"
					onclick={() => toast('The license portal is stubbed in this prototype.')}
				>
					Request a free small-tier license
				</button>
			</div>
		</form>
	</div>
</div>
