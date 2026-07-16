<script lang="ts">
	import Page from '$lib/components/layout/page.svelte';
	import ScheduleForm from '$lib/components/oncall/schedule-form.svelte';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<Page title="On-call" subtitle="Schedules, overrides, and who is next">
	<ScheduleForm
		data={data.form}
		heading="New schedule"
		submitLabel="Create schedule"
		back={ws('/on-call')}
		previewFrom={data.previewFrom}
	/>
</Page>
