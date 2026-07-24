<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import ArrowUpRightIcon from '@lucide/svelte/icons/arrow-up-right';
	import BellOffIcon from '@lucide/svelte/icons/bell-off';
	import LinkIcon from '@lucide/svelte/icons/link';
	import SirenIcon from '@lucide/svelte/icons/siren';
	import BookOpenIcon from '@lucide/svelte/icons/book-open';
	import BracesIcon from '@lucide/svelte/icons/braces';
	import ChartLineIcon from '@lucide/svelte/icons/chart-line';
	import CheckIcon from '@lucide/svelte/icons/check';
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import PlugIcon from '@lucide/svelte/icons/plug';
	import { enhance } from '$app/forms';
	import AlertStatus from '$lib/components/alerts/alert-status.svelte';
	import EscalationTimeline from '$lib/components/alerts/escalation-timeline.svelte';
	import Page from '$lib/components/layout/page.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import { SEVERITY_TONE } from '$lib/alerts';
	import { ws } from '$lib/navigation';
	import { formatSince, formatUtc } from '$lib/time';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const PENDING_INCIDENTS = 'Available once incidents ship.';

	const alert = $derived(data.alert);
	const LINK_ICON = { runbook: BookOpenIcon, dashboard: ChartLineIcon, source: PlugIcon };

</script>

<Page title="Alert" subtitle="Deduplicated signals from every connected source">
	<div class="flex flex-col gap-3.5">
		<a
			href={ws('/alerts')}
			class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[12.5px]"
		>
			<ArrowLeftIcon class="size-3.5" />
			All alerts
		</a>

		<section
			class="bg-card overflow-hidden rounded-xl border"
			style="box-shadow: inset 3px 0 0 var(--{alert.severity})"
		>
			<div class="flex flex-col gap-3.5 px-[18px] py-4">
				<div class="flex flex-wrap items-start gap-3">
					<Badge tone={SEVERITY_TONE[alert.severity]}>{alert.severity.toUpperCase()}</Badge>

					<div class="min-w-0 flex-1">
						<h2 class="tracking-heading m-0 text-[18px] font-semibold">{alert.title}</h2>
						<div class="text-subtle-foreground mt-[3px] font-mono text-[11.5px]">
							{alert.source} · dedup ×{alert.count} · first seen {formatUtc(alert.firstSeenAt)} ·
							last seen {formatUtc(alert.lastSeenAt)}
						</div>
					</div>

					<AlertStatus {alert} size="md" />
				</div>

				<p class="text-muted-foreground m-0 max-w-[720px] text-[13.5px] leading-[1.6]">
					{alert.description}
				</p>

				{#if alert.escalation}
					<div class="text-muted-foreground flex flex-wrap items-center gap-2 text-[12.5px]">
						<Badge
							tone={alert.escalation.state === 'exhausted'
								? 'critical'
								: alert.escalation.state === 'running'
									? 'brand'
									: alert.escalation.state === 'acked'
										? 'success'
										: 'neutral'}
							size="sm"
							dot
						>
							escalation {alert.escalation.state}
						</Badge>
						<span>
							level {Math.min(alert.escalation.stepIndex + (alert.escalation.state === 'running' ? 1 : 0), alert.escalation.totalSteps) || alert.escalation.stepIndex}
							of {alert.escalation.totalSteps}
							through
							<a href={ws(`/escalation-policies/${alert.escalation.policySlug}`)} class="text-brand-foreground font-mono hover:underline">
								{alert.escalation.policySlug}
							</a>
						</span>
						{#if alert.escalation.state === 'running' && alert.escalation.nextAt}
							<span class="text-subtle-foreground font-mono text-[11.5px]">
								next step {formatUtc(alert.escalation.nextAt)}
							</span>
						{:else if alert.escalation.state === 'acked' && alert.escalation.ackExpiresAt}
							<span class="text-subtle-foreground font-mono text-[11.5px]">
								resumes {formatUtc(alert.escalation.ackExpiresAt)} unless resolved
							</span>
						{/if}
					</div>
				{/if}

				<div class="flex flex-wrap gap-2">
					{#if alert.status === 'open'}
						<form method="POST" action="?/ack" use:enhance>
							<Button type="submit" size="sm">
								<CheckIcon data-icon="inline-start" />
								Acknowledge
							</Button>
						</form>
					{/if}

					{#if alert.status !== 'resolved'}
						<form method="POST" action="?/resolve" use:enhance>
							<Button
								type="submit"
								size="sm"
								variant={alert.status === 'open' ? 'secondary' : 'default'}
							>
								<CircleCheckIcon data-icon="inline-start" />
								Resolve
							</Button>
						</form>
					{/if}

					<Button size="sm" variant="secondary" href={ws(`/alerts/silences?source=${alert.source}`)}>
						<BellOffIcon data-icon="inline-start" />
						Silence source
					</Button>

					{#if alert.status === 'open' && alert.escalation && alert.escalation.state === 'running'}
						<form method="POST" action="?/escalate" use:enhance>
							<Button size="sm" variant="secondary" type="submit">
								<ArrowUpRightIcon data-icon="inline-start" />
								Escalate to next step
							</Button>
						</form>
					{/if}

					<form method="POST" action="?/declare" use:enhance>
						<Button size="sm" variant="destructive" type="submit">
							<SirenIcon data-icon="inline-start" />
							Promote to incident
						</Button>
					</form>

					<Button size="sm" variant="ghost" disabled title={PENDING_INCIDENTS}>
						<LinkIcon data-icon="inline-start" />
						Attach to incident
					</Button>
				</div>
			</div>

			<div class="bg-inset grid gap-4 border-t px-[18px] py-3.5 min-[860px]:grid-cols-3">
				<div>
					<div class="text-subtle-foreground tracking-label mb-[7px] text-[11px] uppercase">
						Service
					</div>
					<Tag>{alert.service}</Tag>
				</div>

				<div>
					<div class="text-subtle-foreground tracking-label mb-[7px] text-[11px] uppercase">
						Labels
					</div>
					<div class="flex flex-wrap gap-1.5">
						{#each alert.labels as label (label)}
							<Tag>{label}</Tag>
						{/each}
					</div>
				</div>

				<div>
					<div class="text-subtle-foreground tracking-label mb-[7px] text-[11px] uppercase">
						Links
					</div>
					<div class="flex flex-col gap-1.5">
						{#each alert.links as link (link.label)}
							{@const Icon = LINK_ICON[link.kind]}
							<a
								href={ws('/catalog')}
								class="text-brand-foreground inline-flex items-center gap-1.5 text-[12.5px] hover:underline"
							>
								<Icon class="size-3.5" />
								{link.label}
							</a>
						{/each}
					</div>
				</div>
			</div>
		</section>

		<section class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2.5 border-b px-4 py-3">
				<span class="text-sm font-semibold">Escalation timeline</span>
				<span class="text-subtle-foreground ml-auto font-mono text-[11px]">UTC + local</span>
			</header>
			<EscalationTimeline events={alert.timeline} />
		</section>

		<section class="bg-card overflow-hidden rounded-xl border">
			<Collapsible.Root>
				<Collapsible.Trigger
					class="hover:bg-accent text-foreground flex w-full items-center gap-2 px-4 py-3.5 text-[13.5px] font-semibold"
				>
					<BracesIcon class="size-3.5" />
					Raw payload
				</Collapsible.Trigger>
				<Collapsible.Content>
					<pre
						class="bg-inset text-muted-foreground mx-4 mb-3.5 overflow-x-auto rounded-md border px-3.5 py-3 font-mono text-xs leading-[1.65] whitespace-pre-wrap [overflow-wrap:anywhere]">{alert.payload}</pre>
				</Collapsible.Content>
			</Collapsible.Root>
		</section>
	</div>
</Page>

