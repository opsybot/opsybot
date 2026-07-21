<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import RotateCwIcon from '@lucide/svelte/icons/rotate-cw';
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import IncidentTabs from '$lib/components/incidents/incident-tabs.svelte';
	import ResolveDialog from '$lib/components/incidents/resolve-dialog.svelte';
	import StatusStepper from '$lib/components/incidents/status-stepper.svelte';
	import UpdateCountdown from '$lib/components/incidents/update-countdown.svelte';
	import Page from '$lib/components/layout/page.svelte';
	import Tag from '$lib/components/tag.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { SEVERITY_TONE } from '$lib/dashboard';
	import { PEOPLE, SEVERITIES } from '$lib/incidents';
	import { ws } from '$lib/navigation';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();

	const incident = $derived(data.incident);
	const resolved = $derived(incident.status === 'resolved');
	const base = $derived(ws(`/incidents/${incident.id}`));

	let renaming = $state(false);
	let draft = $state('');
	let resolving = $state(false);

	const tab = $derived(page.url.pathname.slice(base.length));

	let form: HTMLFormElement | undefined = $state();
	let roleForms: Record<string, HTMLFormElement | undefined> = $state({});
</script>

<Page title="Incidents" subtitle="From alert to postmortem">
	<div class="flex flex-col gap-3.5">
		<a
			href={ws('/incidents')}
			class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[12.5px]"
		>
			<ArrowLeftIcon class="size-3.5" />
			All incidents
		</a>

		<div class="flex flex-wrap items-center gap-2.5">
			<Badge tone={SEVERITY_TONE[incident.severity]}>{incident.severity}</Badge>

			{#if renaming}
				<form
					method="POST"
					action="{base}?/rename"
					use:enhance={() => async ({ update }) => {
						await update({ reset: false });
						renaming = false;
					}}
					class="flex min-w-[260px] flex-1 items-center gap-2"
				>
					<!-- svelte-ignore a11y_autofocus -->
					<Input
						name="name"
						bind:value={draft}
						autofocus
						class="h-[34px] flex-1 text-[13px]"
						onkeydown={(event) => {
							if (event.key === 'Escape') renaming = false;
						}}
					/>
					<Button type="submit" size="sm">Save</Button>
				</form>
			{:else}
				<h2 class="tracking-heading m-0 min-w-0 text-[19px] font-semibold">
					{incident.name}
					<Button
						variant="ghost"
						size="icon-sm"
						aria-label="Rename"
						class="ml-1.5 align-middle"
						onclick={() => {
							draft = incident.name;
							renaming = true;
						}}
					>
						<PencilIcon />
					</Button>
				</h2>
			{/if}

			<span class="text-subtle-foreground font-mono text-xs">{incident.id}</span>
			<div class="flex-1"></div>

			<form method="POST" action="{base}?/severity" use:enhance bind:this={form}>
				<Select.Root
					type="single"
					name="severity"
					value={incident.severity}
					onValueChange={() => form?.requestSubmit()}
				>
					<Select.Trigger size="sm" class="w-24">{incident.severity}</Select.Trigger>
					<Select.Content>
						<Select.Group>
							{#each SEVERITIES as level (level.id)}
								<Select.Item value={level.id} label={level.id}>{level.id}</Select.Item>
							{/each}
						</Select.Group>
					</Select.Content>
				</Select.Root>
			</form>

			<a
				href={ws('/chat')}
				class="text-muted-foreground hover:text-brand-foreground hover:border-brand-edge border-input inline-flex items-center gap-1.5 rounded-full border px-[11px] py-[5px] font-mono text-xs"
			>
				<MessageSquareIcon class="size-3.5" />
				#inc-{incident.id.split('-')[1].toLowerCase()}
			</a>
		</div>

		<StatusStepper
			stage={incident.status}
			action="{base}?/status"
			onresolve={() => (resolving = true)}
		/>

		<div class="bg-card grid gap-6 rounded-xl border px-3.5 py-3 min-[900px]:grid-cols-[auto_1fr_auto]">
			<div class="flex flex-col gap-2">
				{#each [['lead', 'Lead', incident.lead], ['comms', 'Comms', incident.comms]] as [role, label, person] (role)}
					<div class="flex items-center gap-2">
						<span class="text-subtle-foreground w-13 font-mono text-[10.5px] tracking-[0.06em] uppercase">
							{label}
						</span>
						<form method="POST" action="{base}?/role" use:enhance bind:this={roleForms[role]}>
							<input type="hidden" name="role" value={role} />
							<Select.Root
								type="single"
								name="person"
								value={person}
								onValueChange={() => roleForms[role]?.requestSubmit()}
							>
								<Select.Trigger size="sm" class="w-[150px]">{person}</Select.Trigger>
								<Select.Content>
									<Select.Group>
										{#each PEOPLE as candidate (candidate)}
											<Select.Item value={candidate} label={candidate}>{candidate}</Select.Item>
										{/each}
									</Select.Group>
								</Select.Content>
							</Select.Root>
						</form>
					</div>
				{/each}
			</div>

			<div class="flex flex-col gap-2">
				<span class="text-subtle-foreground font-mono text-[10.5px] tracking-[0.06em] uppercase">
					Affected services
				</span>
				<div class="flex flex-wrap gap-1.5">
					{#each incident.services as service (service)}
						<Tag>{service}</Tag>
					{/each}
				</div>
			</div>

			<div class="flex flex-col items-start gap-2">
				<span class="text-subtle-foreground font-mono text-[10.5px] tracking-[0.06em] uppercase">
					Next update
				</span>
				{#if resolved || !incident.nextUpdateAt}
					<span class="text-subtle-foreground text-[12.5px]">Resolved: no updates due.</span>
				{:else}
					<form method="POST" action="{base}?/post-update" use:enhance>
						<UpdateCountdown dueAt={incident.nextUpdateAt} now={data.now} />
					</form>
				{/if}
			</div>
		</div>

		{#if resolved}
			<div class="flex items-center gap-3">
				<Alert.Root tone="success" class="flex-1">
					<CircleCheckIcon />
					<Alert.Content>
						<Alert.Title>Resolved</Alert.Title>
						<Alert.Description>
							Reopening is recorded on the timeline and restarts update reminders.
						</Alert.Description>
					</Alert.Content>
				</Alert.Root>
				<form method="POST" action="{base}?/reopen" use:enhance>
					<Button type="submit" variant="secondary" size="sm" class="shrink-0">
						<RotateCwIcon data-icon="inline-start" />
						Reopen
					</Button>
				</form>
			</div>
		{/if}

		<IncidentTabs {base} current={tab} />

		{@render children()}
	</div>
</Page>

<ResolveDialog
	bind:open={resolving}
	incidentId={incident.id}
	linkedAlerts={incident.alerts.filter((alert) => alert.status !== 'resolved').length}
/>
