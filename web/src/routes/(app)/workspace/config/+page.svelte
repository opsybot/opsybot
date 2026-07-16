<script lang="ts">
	import type { Component } from 'svelte';
	import { onDestroy } from 'svelte';
	import type { LucideProps } from '@lucide/svelte';
	import CheckIcon from '@lucide/svelte/icons/check';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import MinusIcon from '@lucide/svelte/icons/minus';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import UploadIcon from '@lucide/svelte/icons/upload';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import * as Alert from '$lib/components/ui/alert';
	import { Badge } from '$lib/components/ui/badge';
	import { Button, buttonVariants } from '$lib/components/ui/button';
	import * as Select from '$lib/components/ui/select';
	import { DIFF_GROUPS, IMPORT_DECISIONS, type DiffKind } from '$lib/admin';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const DIFF_ICONS: Record<string, Component<LucideProps>> = {
		plus: PlusIcon,
		pencil: PencilIcon,
		'triangle-alert': TriangleAlertIcon,
		minus: MinusIcon
	};

	const dataUri = $derived(`data:text/yaml;charset=utf-8,${encodeURIComponent(data.exportYaml)}`);

	let stage = $state<'idle' | 'validating' | 'diff' | 'applied'>('idle');
	let decision = $state('');
	let timer: ReturnType<typeof setTimeout>;
	let applyForm: HTMLFormElement;

	function upload() {
		stage = 'validating';
		clearTimeout(timer);
		timer = setTimeout(() => (stage = 'diff'), 1600);
	}
	function reset() {
		stage = 'idle';
		decision = '';
	}
	onDestroy(() => clearTimeout(timer));

	const decisionLabel = (value: string) => IMPORT_DECISIONS.find((d) => d.value === value)?.label ?? 'Decide…';
</script>

<div class="flex max-w-[780px] flex-col gap-3.5">
	<div class="bg-card flex flex-wrap items-center gap-3.5 rounded-xl border p-4">
		<div class="min-w-[240px] flex-1">
			<div class="text-[13.5px] font-semibold">Export workspace configuration</div>
			<div class="text-subtle-foreground mt-0.5 text-[12px] leading-[1.5]">
				Schedules, policies, routing, workflows, severities, custom fields — one YAML file.
				<button
					type="button"
					class="text-brand-foreground ml-1.5 hover:underline"
					onclick={() => toast('Format docs are stubbed in this prototype.')}
				>
					Format reference
				</button>
			</div>
		</div>
		<a
			class={buttonVariants({ variant: 'secondary', size: 'sm' })}
			href={dataUri}
			download="acme-corp.opsybot.yaml"
			onclick={() => toast.success('Configuration exported.')}
		>
			<DownloadIcon data-icon="inline-start" />
			Download config
		</a>
	</div>

	<div class="bg-card flex flex-col gap-3 rounded-xl border p-4">
		<div class="text-[13.5px] font-semibold">Import configuration</div>
		{#if stage === 'idle'}
			<div class="flex flex-wrap items-center gap-3">
				<button
					type="button"
					onclick={upload}
					class="border-border-strong bg-inset text-muted-foreground hover:text-foreground hover:border-brand-edge inline-flex items-center gap-[9px] rounded-md border border-dashed px-[18px] py-3 text-[12.5px] transition-colors"
				>
					<UploadIcon class="size-[15px]" />
					Choose file — .yaml or .json
				</button>
				<span class="text-subtle-foreground text-[12px]">Nothing changes until you confirm the dry-run diff.</span>
			</div>
		{/if}
		<div role="status" aria-live="polite" class="contents">
			{#if stage === 'validating'}
				<div class="text-muted-foreground flex items-center gap-2.5">
					<span
						class="border-border border-t-primary size-4 shrink-0 animate-spin rounded-full border-2 [animation-duration:0.8s] motion-reduce:animate-none"
						aria-hidden="true"
					></span>
					<span class="text-[13px]">Validating acme-corp.opsybot.yaml — schema, references, permissions…</span>
				</div>
			{/if}
			{#if stage === 'diff' || stage === 'applied'}
				<Alert.Root tone={stage === 'applied' ? 'success' : 'info'}>
					{#if stage === 'applied'}<CheckIcon />{:else}<TriangleAlertIcon />{/if}
					<Alert.Content>
						<Alert.Title>{stage === 'applied' ? 'Applied' : 'Validation passed — dry run below'}</Alert.Title>
						<Alert.Description>
							{stage === 'applied'
								? 'The configuration is live. Re-import any time; unchanged items are skipped.'
								: 'acme-corp.opsybot.yaml · 6 resources · schema v2 · no permission problems. Review the diff, resolve decisions, then apply.'}
						</Alert.Description>
					</Alert.Content>
				</Alert.Root>
			{/if}
		</div>
	</div>

	{#if stage === 'diff' || stage === 'applied'}
		{#each DIFF_GROUPS as group (group.kind)}
			{@const items = data.diff[group.kind as DiffKind]}
			{@const Icon = DIFF_ICONS[group.icon]}
			{#if items.length}
				<div class="bg-card overflow-hidden rounded-xl border">
					<header class="flex items-center gap-2 border-b px-4 py-3">
						<Icon
							class="size-3.5 {group.tone === 'success'
								? 'text-[var(--success)]'
								: group.tone === 'warning'
									? 'text-[var(--warning)]'
									: group.tone === 'info'
										? 'text-[var(--info)]'
										: 'text-subtle-foreground'}"
						/>
						<span class="text-[13px] font-semibold">{group.label}</span>
						<Badge tone={group.tone} size="sm">{items.length}</Badge>
					</header>
					{#each items as item (item.path)}
						<div class="flex items-start gap-2.5 border-t px-4 py-2.5 first:border-t-0">
							<span class="text-foreground w-[240px] shrink-0 font-mono text-[12px]">{item.path}</span>
							<span class="text-muted-foreground flex-1 text-[12.5px] leading-[1.5]">{item.note}</span>
							{#if group.kind === 'decision'}
								{#if stage === 'applied'}
									<Badge tone="success" size="sm">{decision === 'skip' ? 'skipped' : 'reassigned'}</Badge>
								{:else}
									<Select.Root type="single" value={decision} onValueChange={(value) => (decision = value)}>
										<Select.Trigger size="sm" class="w-[190px]" aria-label="Decision for {item.path}">
											{decisionLabel(decision)}
										</Select.Trigger>
										<Select.Content>
											<Select.Group>
												{#each IMPORT_DECISIONS as option (option.value)}
													<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
												{/each}
											</Select.Group>
										</Select.Content>
									</Select.Root>
								{/if}
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		{/each}

		{#if stage === 'diff'}
			<div class="flex gap-2.5">
				<Button disabled={!decision} onclick={() => applyForm.requestSubmit()}>
					<CheckIcon data-icon="inline-start" />
					Apply configuration
				</Button>
				<Button variant="ghost" onclick={reset}>Cancel</Button>
			</div>
		{:else}
			<Button variant="secondary" class="self-start" onclick={reset}>Import another file</Button>
		{/if}
	{/if}
</div>

<form
	bind:this={applyForm}
	method="POST"
	action="?/apply"
	class="hidden"
	use:enhance={() => async ({ result, update }) => {
		if (result.type === 'failure') {
			toast.error(String(result.data?.error ?? 'Could not apply.'));
			return;
		}
		if (result.type !== 'success') return;
		await update({ reset: false });
		stage = 'applied';
		toast.success(
			`Configuration applied — 2 created, 2 changed, 1 ${decision === 'skip' ? 'skipped' : 'reassigned'}, 1 unchanged. Recorded in the audit log.`
		);
	}}
>
	<input type="hidden" name="decision" value={decision} />
</form>
