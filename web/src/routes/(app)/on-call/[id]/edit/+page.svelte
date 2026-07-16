<script lang="ts">
	import Page from '$lib/components/layout/page.svelte';
	import ScheduleForm from '$lib/components/oncall/schedule-form.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<Page title="On-call" subtitle="Schedules, overrides, and who is next">
	<ScheduleForm
		data={data.form}
		heading="Edit schedule"
		submitLabel="Save schedule"
		back="/on-call/{data.id}"
		previewFrom={data.previewFrom}
	/>
</Page>
