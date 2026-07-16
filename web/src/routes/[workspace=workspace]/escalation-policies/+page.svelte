<script lang="ts">
	import ArrowUpRightIcon from '@lucide/svelte/icons/arrow-up-right';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import Tag from '$lib/components/tag.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<div class="flex flex-col gap-3.5">
	<div class="flex items-center">
		<span class="text-subtle-foreground text-[13px]">
			{data.policies.length}
			{data.policies.length === 1 ? 'policy' : 'policies'}
		</span>
		<div class="flex-1"></div>
		<Button size="sm" href={ws('/escalation-policies/new')}>
			<PlusIcon data-icon="inline-start" />
			New policy
		</Button>
	</div>

	{#if data.policies.length === 0}
		<div
			class="text-muted-foreground flex flex-col items-center gap-2.5 rounded-xl border border-dashed px-5 py-14"
		>
			<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
				<ArrowUpRightIcon class="text-subtle-foreground size-5" />
			</span>
			<div class="text-[15px] font-medium">No escalation policies</div>
			<p class="text-subtle-foreground m-0 max-w-[420px] text-center text-[13px] leading-[1.55]">
				A policy decides who gets paged, in what order, and how long to wait between steps. Alerts
				route to a policy; the policy does the chasing.
			</p>
			<Button size="sm" variant="secondary" href={ws('/escalation-policies/new')}>
				<PlusIcon data-icon="inline-start" />
				Create your first policy
			</Button>
		</div>
	{:else}
		<div class="bg-card overflow-hidden rounded-xl border">
			{#each data.policies as policy (policy.id)}
				<a
					href={ws(`/escalation-policies/${policy.id}`)}
					data-policy={policy.id}
					class="hover:bg-accent flex items-center gap-[18px] border-t px-4 py-3.5 first:border-t-0"
				>
					<div class="min-w-0 flex-1">
						<div class="flex flex-wrap items-center gap-2">
							<span class="font-mono text-[13.5px] font-medium">{policy.name}</span>
							<Tag>{policy.team}</Tag>
							{#if policy.branch}
								<Badge tone="info" size="sm" dot>branches by {policy.branch}</Badge>
							{/if}
							{#if policy.warning}
								<Badge tone="warning" size="sm" dot>{policy.warning}</Badge>
							{/if}
						</div>
						<div class="text-muted-foreground mt-1.5 flex items-center gap-1.5 overflow-hidden">
							{#each policy.summary as part, index (index)}
								{#if index > 0}
									<ChevronRightIcon class="text-subtle-foreground size-[11px] shrink-0" />
								{/if}
								<span
									class="text-[11.5px] whitespace-nowrap {part.kind === 'wait'
										? 'font-mono'
										: ''} {part.kind === 'branch' ? 'text-brand-foreground' : ''}"
								>
									{part.text}
								</span>
							{/each}
						</div>
					</div>
					<div class="shrink-0 text-right">
						<div class="font-mono text-[15px] font-semibold">{policy.routed}</div>
						<div class="text-subtle-foreground text-[10.5px]">alerts routed</div>
					</div>
					<ChevronRightIcon class="text-subtle-foreground size-4 shrink-0" />
				</a>
			{/each}
		</div>
	{/if}
</div>
