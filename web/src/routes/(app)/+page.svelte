<script lang="ts">
	import AlertsCard from '$lib/components/home/alerts-card.svelte';
	import EmptyIncidents from '$lib/components/home/empty-incidents.svelte';
	import HealthStrip from '$lib/components/home/health-strip.svelte';
	import IncidentsCard from '$lib/components/home/incidents-card.svelte';
	import OnCallBanner from '$lib/components/home/on-call-banner.svelte';
	import OnCallNowCard from '$lib/components/home/on-call-now-card.svelte';
	import OnboardingChecklist from '$lib/components/home/onboarding-checklist.svelte';
	import OverdueCard from '$lib/components/home/overdue-card.svelte';
	import ShiftsCard from '$lib/components/home/shifts-card.svelte';
	import Page from '$lib/components/layout/page.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const home = $derived(data.dashboard);

	const degraded = $derived(
		home.instance.selfHosted && home.instance.workersHealthy < home.instance.workersTotal
	);

	let dismissed = $state(false);
	const setup = $derived(home.onboarding && !home.onboarding.dismissed && !dismissed);

	const onCall = $derived(home.onCallNow.some((entry) => entry.you));
	const currentShift = $derived(
		onCall ? home.myShifts.find((shift) => home.now < Date.parse(shift.end)) : undefined
	);
</script>

<Page title="Home">
	{#if degraded}
		<HealthStrip instance={home.instance} />
	{/if}

	{#if home.onboarding}
		{#if setup}
			<OnboardingChecklist
				onboarding={home.onboarding}
				selfHosted={home.instance.selfHosted}
				ondismiss={() => (dismissed = true)}
			/>
		{/if}
		<EmptyIncidents />
	{:else}
		{#if currentShift}
			<OnCallBanner shift={currentShift} now={home.now} />
		{/if}

		<div
			class="grid grid-cols-1 items-start gap-4 min-[1100px]:grid-cols-[minmax(0,1fr)_312px]"
		>
			<div class="flex min-w-0 flex-col gap-4">
				<IncidentsCard incidents={home.incidents} now={home.now} />
				<AlertsCard alerts={home.alerts} volume={home.alertVolume} now={home.now} />
			</div>

			<div class="flex flex-col gap-4">
				<OnCallNowCard entries={home.onCallNow} now={home.now} />
				<ShiftsCard shifts={home.myShifts} now={home.now} />
				<OverdueCard items={home.overdue} now={home.now} />
			</div>
		</div>
	{/if}
</Page>
