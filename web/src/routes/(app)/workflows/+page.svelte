<script lang="ts">
	import HistoryIcon from '@lucide/svelte/icons/history';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import WorkflowIcon from '@lucide/svelte/icons/workflow';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import Page from '$lib/components/layout/page.svelte';
	import RunRow from '$lib/components/workflows/run-row.svelte';
	import WfTabs from '$lib/components/workflows/wf-tabs.svelte';
	import WorkflowToggle from '$lib/components/workflows/workflow-toggle.svelte';
	import { ICON } from '$lib/components/workflows/icons';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import { formatUtc } from '$lib/time';
	import { describeTrigger, lastRun } from '$lib/workflows';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<Page title="Workflows" subtitle="Automate the boring half of incident response">
	<WfTabs current="workflows" />

	<div class="mt-3.5 flex flex-col gap-3.5">
		<div class="flex items-center">
			<span class="text-subtle-foreground text-[13px]">
				{data.workflows.length}
				{data.workflows.length === 1 ? 'workflow' : 'workflows'}
			</span>
			<div class="flex-1"></div>
			<Button size="sm" href="/workflows/new">
				<PlusIcon data-icon="inline-start" />
				New workflow
			</Button>
		</div>

		{#each data.workflows as workflow (workflow.id)}
			{@const run = lastRun(workflow)}
			<div data-workflow={workflow.id} class="bg-card overflow-hidden rounded-xl border">
				<div class="flex items-center gap-3 px-4 py-[13px]">
					<span
						class="bg-inset text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-sm border"
					>
						<WorkflowIcon class="size-[15px]" />
					</span>
					<div class="min-w-0 flex-1">
						<div class="flex flex-wrap items-center gap-2">
							<span class="text-[13.5px] font-semibold">{workflow.name}</span>
							<Badge tone="neutral" size="sm">
								{workflow.actions.length}
								{workflow.actions.length === 1 ? 'action' : 'actions'}
							</Badge>
							{#if run && !run.ok}
								<Badge tone="critical" size="sm" dot>last run failed</Badge>
							{/if}
						</div>
						<div class="text-subtle-foreground mt-[3px] font-mono text-[11px]">
							when {describeTrigger(workflow.trigger, workflow.conditions)}{run
								? ` · last ran ${formatUtc(run.at)} (${run.incident})`
								: ' · never run'}
						</div>
					</div>
					<Button size="sm" variant="ghost" href="/workflows/{workflow.id}">
						<PencilIcon data-icon="inline-start" />
						Edit
					</Button>
					<WorkflowToggle id={workflow.id} name={workflow.name} enabled={workflow.enabled} />
				</div>

				{#if workflow.history.length}
					<Collapsible.Root>
						<Collapsible.Trigger
							class="text-muted-foreground hover:text-foreground group/wf inline-flex items-center gap-2 rounded-sm border px-2.5 py-[7px] text-sm font-medium transition-colors"
						>
							<ChevronRightIcon
								class="size-[15px] transition-transform duration-150 group-data-[state=open]/wf:rotate-90"
							/>
							<span class="inline-flex items-center gap-[7px] text-[12.5px] font-semibold">
								<HistoryIcon class="size-3" />
								Run history
								<Badge tone="neutral" size="sm">{workflow.history.length}</Badge>
							</span>
						</Collapsible.Trigger>
						<Collapsible.Content>
							<div class="px-3.5 pb-3">
								{#each workflow.history as entry (entry.id)}
									<RunRow workflowId={workflow.id} run={entry} />
								{/each}
							</div>
						</Collapsible.Content>
					</Collapsible.Root>
				{/if}
			</div>
		{:else}
			<div
				class="text-muted-foreground flex flex-col items-center gap-2.5 rounded-xl border border-dashed px-5 py-14"
			>
				<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
					<WorkflowIcon class="text-subtle-foreground size-5" />
				</span>
				<div class="text-[15px] font-medium">No workflows yet</div>
				<p class="text-subtle-foreground m-0 max-w-[420px] text-center text-[13px] leading-[1.55]">
					A workflow runs an action when something happens — a declaration, a severity change, an
					update that has gone overdue. Start from a template below.
				</p>
			</div>
		{/each}

		<div>
			<div class="text-subtle-foreground mt-2 mb-2.5 text-[11px] tracking-[0.08em] uppercase">
				Templates
			</div>
			<div class="grid gap-2.5 [grid-template-columns:repeat(auto-fill,minmax(230px,1fr))]">
				{#each data.templates as template (template.id)}
					{@const Icon = ICON[template.icon]}
					<div
						class="border-border-strong bg-card flex flex-col gap-2 rounded-xl border border-dashed p-3.5"
					>
						<span
							class="bg-inset text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-sm border"
						>
							<Icon class="size-[15px]" />
						</span>
						<div class="text-[13px] font-semibold">{template.name}</div>
						<p class="text-subtle-foreground m-0 flex-1 text-[12px] leading-[1.5]">
							{template.description}
						</p>
						<Button
							size="sm"
							variant="secondary"
							class="self-start"
							href="/workflows/new?template={template.id}"
						>
							Use template
						</Button>
					</div>
				{/each}
			</div>
		</div>
	</div>
</Page>
