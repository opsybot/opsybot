<script lang="ts">
	import type { Component } from 'svelte';
	import { onDestroy } from 'svelte';
	import type { LucideProps } from '@lucide/svelte';
	import CheckIcon from '@lucide/svelte/icons/check';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import GitBranchIcon from '@lucide/svelte/icons/git-branch';
	import MinusIcon from '@lucide/svelte/icons/minus';
	import PlugIcon from '@lucide/svelte/icons/plug';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import RotateCwIcon from '@lucide/svelte/icons/rotate-cw';
	import SirenIcon from '@lucide/svelte/icons/siren';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import UserXIcon from '@lucide/svelte/icons/user-x';
	import { toast } from 'svelte-sonner';
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import Page from '$lib/components/layout/page.svelte';
	import { IMP_STEPS, isValidApiKey, type DecisionKind } from '$lib/importer';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const DECISION_ICON: Record<DecisionKind, Component<LucideProps>> = {
		user: UserXIcon,
		integration: PlugIcon,
		routing: GitBranchIcon
	};

	let step = $state(0);
	let apiKey = $state('');
	let region = $state('EU (api.eu.opsgenie.com)');
	let connecting = $state(false);
	let decisions = $state<Record<string, string>>({});
	let importing = $state(false);
	let imported = $state(false);
	let checked = $state<Record<number, boolean>>({});
	let testState = $state<'idle' | 'sending' | 'ok'>('idle');

	const allDecided = $derived(data.dryrun.decisions.every((decision) => decisions[decision.id]));
	const leftToResolve = $derived(data.dryrun.decisions.filter((decision) => !decisions[decision.id]).length);

	let connectTimer: ReturnType<typeof setTimeout>;
	let importTimer: ReturnType<typeof setTimeout>;
	let testTimer: ReturnType<typeof setTimeout>;

	function connect() {
		connecting = true;
		clearTimeout(connectTimer);
		connectTimer = setTimeout(() => {
			connecting = false;
			step = 1;
		}, 1600);
	}
	function runImport() {
		importing = true;
		clearTimeout(importTimer);
		importTimer = setTimeout(() => {
			importing = false;
			imported = true;
		}, 2400);
	}
	function runTest() {
		testState = 'sending';
		clearTimeout(testTimer);
		testTimer = setTimeout(() => (testState = 'ok'), 1800);
	}
	onDestroy(() => {
		clearTimeout(connectTimer);
		clearTimeout(importTimer);
		clearTimeout(testTimer);
	});

	const next = () => (step = Math.min(6, step + 1));
	const back = () => (step = Math.max(0, step - 1));

	function dotClass(index: number): string {
		if (index === step) return 'bg-brand-wash border-brand-edge text-brand-foreground border';
		if (index < step) return 'border-[var(--mint-500)] bg-[var(--mint-500)] text-[var(--text-inverse)] border';
		return 'bg-inset text-subtle-foreground border';
	}
</script>

{#snippet spinner()}
	<span
		class="border-border border-t-primary size-4 shrink-0 animate-spin rounded-full border-2 [animation-duration:0.8s] motion-reduce:animate-none"
		aria-hidden="true"
	></span>
{/snippet}

<Page title="Import from Opsgenie" subtitle="Dry-run first, cut over when it matches">
	<div class="flex max-w-[780px] flex-col gap-[18px]">
		<ol class="flex flex-wrap items-center gap-[5px]" aria-label="Import progress">
			{#each IMP_STEPS as label, i (label)}
				<li
					aria-current={i === step ? 'step' : undefined}
					class="flex items-center gap-[5px] text-[11.5px] {i === step
						? 'text-foreground font-semibold'
						: i < step
							? 'text-muted-foreground'
							: 'text-subtle-foreground'}"
				>
					{#if i > 0}
						<span class="h-px w-[14px] shrink-0 {i <= step ? 'bg-[var(--mint-500)]' : 'bg-border'}" aria-hidden="true"></span>
					{/if}
					<span class="inline-flex items-center gap-1.5">
						<span class="flex size-[19px] shrink-0 items-center justify-center rounded-full font-mono text-[10px] {dotClass(i)}" aria-hidden="true">
							{#if i < step}<CheckIcon class="size-2.5" />{:else}{i + 1}{/if}
						</span>
						<span class="whitespace-nowrap">{label}</span>
						{#if i < step}<span class="sr-only">, completed</span>{/if}
					</span>
				</li>
			{/each}
		</ol>

		{#if step === 0}
			<div class="bg-card flex flex-col gap-3.5 rounded-xl border p-[18px]">
				<h2 class="m-0 text-[17px] font-semibold">Connect Opsgenie</h2>
				<p class="text-muted-foreground m-0 text-[13px] leading-[1.6]">
					Paste a read-only Opsgenie API key. We only read your config: schedules, policies, teams, users, services.
					Nothing in Opsgenie is changed.
				</p>
				<Field.Field class="gap-1.5 space-y-0">
					<Field.FieldLabel for="og-key" class="text-muted-foreground text-[13px] font-medium">Opsgenie API key</Field.FieldLabel>
					<Input id="og-key" type="password" class="font-mono text-[12.5px]" placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" bind:value={apiKey} />
				</Field.Field>
				<Field.Field class="max-w-[300px] gap-1.5 space-y-0">
					<Field.FieldLabel for="og-region" class="text-muted-foreground text-[13px] font-medium">Region</Field.FieldLabel>
					<Input id="og-region" bind:value={region} />
				</Field.Field>
				<Button class="self-start" disabled={!isValidApiKey(apiKey) || connecting} onclick={connect}>
					<PlugIcon data-icon="inline-start" />
					{connecting ? 'Connecting…' : 'Connect & scan'}
				</Button>
				<div role="status" aria-live="polite" class="sr-only">
					{connecting ? 'Connecting to Opsgenie and scanning your configuration…' : ''}
				</div>
			</div>
		{:else if step === 1}
			<Alert.Root tone="info">
				<EyeIcon />
				<Alert.Content>
					<Alert.Description>Dry run: nothing is written yet. This is what a real import would do.</Alert.Description>
				</Alert.Content>
			</Alert.Root>

			<div class="bg-card overflow-hidden rounded-xl border">
				<header class="flex items-center gap-2 border-b px-4 py-[11px]">
					<PlusIcon class="size-[13px] text-[var(--success)]" />
					<span class="text-[13px] font-semibold">Will be created</span>
				</header>
				<div>
					{#each data.dryrun.created as created (created.kind)}
						<div class="flex items-center gap-3 border-t px-4 py-2.5 first:border-t-0" data-created={created.kind}>
							<span class="text-subtle-foreground w-10 font-mono text-[14px] font-semibold">{created.n}</span>
							<span class="w-[150px] text-[13px] font-medium">{created.kind}</span>
							<span class="text-subtle-foreground flex-1 text-[12px]">{created.note}</span>
						</div>
					{/each}
				</div>
			</div>

			<div class="bg-card overflow-hidden rounded-xl border">
				<header class="flex items-center gap-2 border-b px-4 py-[11px]">
					<TriangleAlertIcon class="size-[13px] text-[var(--warning)]" />
					<span class="text-[13px] font-semibold">Needs a decision</span>
					<Badge tone="warning" size="sm">{data.dryrun.decisions.length}</Badge>
				</header>
				<div>
					{#each data.dryrun.decisions as decision (decision.id)}
						<div class="flex items-center gap-3 border-t px-4 py-2.5 first:border-t-0">
							<span class="text-foreground text-[13px]">{decision.title}</span>
						</div>
					{/each}
				</div>
			</div>

			<div class="bg-card overflow-hidden rounded-xl border">
				<header class="flex items-center gap-2 border-b px-4 py-[11px]">
					<MinusIcon class="text-subtle-foreground size-[13px]" />
					<span class="text-[13px] font-semibold">Skipped</span>
				</header>
				<div>
					{#each data.dryrun.skipped as skip (skip.title)}
						<div class="flex items-start gap-3 border-t px-4 py-2.5 first:border-t-0">
							<span class="w-[260px] shrink-0 text-[13px]">{skip.title}</span>
							<span class="text-subtle-foreground flex-1 text-[12px] leading-[1.5]">{skip.reason}</span>
						</div>
					{/each}
				</div>
			</div>

			<div class="flex gap-2.5">
				<Button onclick={next}>Resolve {data.dryrun.decisions.length} decisions</Button>
				<Button variant="ghost" onclick={back}>Back</Button>
			</div>
		{:else if step === 2}
			<p class="text-muted-foreground m-0 text-[13px]">
				Resolve each item. Still nothing written. These choices feed the import.
			</p>
			{#each data.dryrun.decisions as decision (decision.id)}
				{@const Icon = DECISION_ICON[decision.kind]}
				<div class="bg-card rounded-xl border {decisions[decision.id] ? 'border-brand-edge' : ''}" data-decision={decision.id}>
					<div class="p-3.5">
						<div class="mb-1 flex items-center gap-2">
							<Icon class="size-[14px] text-[var(--warning)]" />
							<span class="text-[13.5px] font-semibold">{decision.title}</span>
							{#if decisions[decision.id]}<Badge tone="success" size="sm" class="ml-auto">resolved</Badge>{/if}
						</div>
						<p class="text-subtle-foreground m-0 mb-2.5 ml-[22px] text-[12.5px] leading-[1.5]">{decision.detail}</p>
						<div class="ml-[22px] flex flex-wrap gap-2" role="group" aria-label="Resolution for: {decision.title}">
							{#each decision.choices as choice (choice.value)}
								<button
									type="button"
									aria-pressed={decisions[decision.id] === choice.value}
									onclick={() => (decisions = { ...decisions, [decision.id]: choice.value })}
									class="rounded-full border px-3 py-1.5 text-[12px] transition-colors {decisions[decision.id] === choice.value
										? 'bg-brand-wash border-brand-edge text-brand-foreground'
										: 'bg-inset text-muted-foreground hover:text-foreground hover:border-border-strong'}"
								>
									{choice.label}
								</button>
							{/each}
						</div>
					</div>
				</div>
			{/each}
			<div class="flex gap-2.5">
				<Button disabled={!allDecided} onclick={next}>
					{allDecided ? 'Continue to import' : `${leftToResolve} left to resolve`}
				</Button>
				<Button variant="ghost" onclick={back}>Back</Button>
			</div>
		{:else if step === 3}
			<div class="bg-card flex flex-col gap-4 rounded-xl border p-5">
				{#if !imported}
					<h2 class="m-0 text-[17px] font-semibold">Ready to import</h2>
					<Alert.Root tone="info">
						<RotateCwIcon />
						<Alert.Content>
							<Alert.Description>
								This import is idempotent: safe to re-run. Re-running updates changed items and skips identical ones,
								so you can import again after fixing anything in Opsgenie.
							</Alert.Description>
						</Alert.Content>
					</Alert.Root>
					{#if importing}
						<div class="flex items-center gap-3" role="status" aria-live="polite">
							{@render spinner()}
							<span class="text-muted-foreground text-[13px]">Importing: 42 users, 6 schedules, 4 policies, 5 teams, 11 services…</span>
						</div>
					{:else}
						<div class="flex gap-2.5">
							<Button onclick={runImport}>
								<DownloadIcon data-icon="inline-start" />
								Run import
							</Button>
							<Button variant="ghost" onclick={back}>Back</Button>
						</div>
					{/if}
				{:else}
					<div class="flex items-center gap-3" role="status" aria-live="polite">
						<span class="flex size-[34px] shrink-0 items-center justify-center rounded-full bg-[var(--mint-500)] shadow-[var(--glow-brand)]">
							<CheckIcon class="size-[18px] text-[var(--text-inverse)]" />
						</span>
						<div>
							<h2 class="m-0 text-[17px] font-semibold">Import complete</h2>
							<p class="text-subtle-foreground m-0 mt-0.5 text-[12.5px]">68 objects created, 0 errors. Verify the result before cutting over.</p>
						</div>
					</div>
					<Button class="self-start" onclick={next}>Verify on-call matches</Button>
				{/if}
			</div>
		{:else if step === 4}
			<p class="text-muted-foreground m-0 text-[13px]">
				Who's on call right now, in both systems. They should match before you cut over.
			</p>
			<div class="bg-card overflow-hidden rounded-xl border">
				<div class="text-subtle-foreground flex items-center gap-3 border-b px-4 py-[9px] text-[10.5px] tracking-[0.06em] uppercase">
					<span class="flex-1">Schedule</span>
					<span class="w-[150px]">Opsybot</span>
					<span class="w-[150px]">Opsgenie</span>
					<span class="w-6"></span>
				</div>
				<div>
					{#each data.compare as row (row.schedule)}
						<div class="flex items-center gap-3 border-t px-4 py-2.5 first:border-t-0" data-compare={row.schedule}>
							<span class="text-foreground flex-1 font-mono text-[12.5px]">{row.schedule}</span>
							<span class="flex w-[150px] items-center gap-[7px]">
								<UserAvatar name={row.opsy} size="xs" onCall />
								<span class="text-[12.5px]">{row.opsy}</span>
							</span>
							<span class="text-muted-foreground w-[150px] text-[12.5px]">{row.og}</span>
							<span class="w-6"><CheckIcon class="size-[15px] text-[var(--mint-500)]" /></span>
						</div>
					{/each}
				</div>
			</div>
			<Alert.Root tone="success">
				<CheckIcon />
				<Alert.Content>
					<Alert.Description>All 5 schedules match. Overrides and next-handover times line up too.</Alert.Description>
				</Alert.Content>
			</Alert.Root>
			<div class="flex gap-2.5">
				<Button onclick={next}>Send a test page</Button>
				<Button variant="ghost" onclick={back}>Back</Button>
			</div>
		{:else if step === 5}
			<div class="bg-card flex flex-col gap-4 rounded-xl border p-5">
				<h2 class="m-0 text-[17px] font-semibold">Send a real test page</h2>
				<p class="text-muted-foreground m-0 text-[13px] leading-[1.6]">
					Pages the current on-call for payments-primary through Opsybot: the real path, real device. Confirm it
					arrives before you re-point any monitoring.
				</p>
				{#if testState === 'idle'}
					<Button class="self-start" onclick={runTest}>
						<SirenIcon data-icon="inline-start" />
						Page Priya Nair (test)
					</Button>
				{:else if testState === 'sending'}
					<div class="flex items-center gap-3" role="status" aria-live="polite">
						{@render spinner()}
						<span class="text-muted-foreground text-[13px]">Paging through push + SMS…</span>
					</div>
				{:else}
					<Alert.Root tone="success">
						<CheckIcon />
						<Alert.Content>
							<Alert.Title>Test page delivered and acknowledged</Alert.Title>
							<Alert.Description>
								Priya's push at 09:44:02 UTC, acknowledged 09:44:19 UTC. The full delivery path works.
							</Alert.Description>
						</Alert.Content>
					</Alert.Root>
					<Button class="self-start" onclick={next}>Get the cutover guide</Button>
				{/if}
				{#if testState !== 'ok'}
					<Button variant="ghost" class="self-start" onclick={back}>Back</Button>
				{/if}
			</div>
		{:else if step === 6}
			<div class="flex items-center gap-2.5">
				<h2 class="m-0 flex-1 text-[17px] font-semibold">Cutover checklist</h2>
				<Button size="sm" variant="secondary" onclick={() => toast('Opens a printable checklist.')}>
					<PrinterIcon data-icon="inline-start" />
					Print
				</Button>
			</div>
			<p class="text-muted-foreground m-0 text-[13px] leading-[1.6]">
				Re-point each source from Opsgenie to Opsybot. Do them one at a time: alerts flow to both until you flip each
				one. Tick as you go.
			</p>
			<div class="bg-card overflow-hidden rounded-xl border">
				{#each data.cutover as source, i (source.source)}
					<label class="flex cursor-pointer items-center gap-3 border-t px-4 py-3 first:border-t-0" data-cutover={i}>
						<Checkbox checked={!!checked[i]} onCheckedChange={(value) => (checked = { ...checked, [i]: value === true })} />
						<div class="min-w-0 flex-1 {checked[i] ? 'line-through opacity-55' : ''}">
							<div class="text-[13px] font-medium">{source.source}</div>
							<div class="text-subtle-foreground mt-0.5 font-mono text-[11px]">{source.from} → {source.to}</div>
						</div>
					</label>
				{/each}
			</div>
			<Alert.Root tone="info">
				<ClockIcon />
				<Alert.Content>
					<Alert.Description>
						Leave Opsgenie active for a week as a safety net. Once Opsybot has paged real incidents cleanly, cancel
						Opsgenie.
					</Alert.Description>
				</Alert.Content>
			</Alert.Root>
			<Button
				variant="secondary"
				class="self-start"
				onclick={() => toast.success('Migration marked complete. Opsgenie import can be re-run any time.')}
			>
				Mark migration complete
			</Button>
		{/if}
	</div>
</Page>
