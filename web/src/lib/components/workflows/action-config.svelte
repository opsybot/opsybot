<script lang="ts">
	import WfSelect from '$lib/components/workflows/wf-select.svelte';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import {
		FOLLOWUP_DUES,
		FOLLOWUP_OWNERS,
		POST_CHANNELS,
		STATUS_PAGE_OPTIONS,
		type WorkflowAction
	} from '$lib/workflows';

	// action is the builder's deep-reactive proxy; binds write into the array it serializes
	let { action, roleNames }: { action: WorkflowAction; roleNames: string[] } = $props();

	const labelClass = 'text-muted-foreground text-[13px] font-medium';
</script>

{#if action.type === 'post'}
	<div class="flex flex-col gap-2">
		<WfSelect
			size="sm"
			label="Channel"
			options={POST_CHANNELS}
			bind:value={action.config.channel}
			class="w-[220px]"
		/>
		<Field.Field class="gap-1.5 space-y-0">
			<Field.FieldLabel for="ac-{action.id}-text" class={labelClass}>Message</Field.FieldLabel>
			<Textarea
				id="ac-{action.id}-text"
				rows={2}
				bind:value={action.config.text}
				placeholder="SEV1 declared: {'{name}'}. Lead: {'{lead}'}."
			/>
			<Field.FieldDescription class="text-subtle-foreground text-xs">
				Placeholders: {'{id}'}
				{'{name}'}
				{'{lead}'}
				{'{channel}'}
				{'{severity}'}
			</Field.FieldDescription>
		</Field.Field>
	</div>
{:else if action.type === 'role'}
	<WfSelect
		size="sm"
		label="Role to prompt for"
		options={roleNames}
		bind:value={action.config.role}
		class="w-[220px]"
	/>
{:else if action.type === 'note'}
	<Field.Field class="gap-1.5 space-y-0">
		<Field.FieldLabel for="ac-{action.id}-note" class={labelClass}>Note text</Field.FieldLabel>
		<Input
			id="ac-{action.id}-note"
			bind:value={action.config.text}
			placeholder="Comms cadence active — updates every 15 min."
		/>
	</Field.Field>
{:else if action.type === 'webhook'}
	<div class="flex flex-col gap-2">
		<Field.Field class="gap-1.5 space-y-0">
			<Field.FieldLabel for="ac-{action.id}-url" class={labelClass}>URL</Field.FieldLabel>
			<Input
				id="ac-{action.id}-url"
				bind:value={action.config.url}
				placeholder="https://hooks.acme.dev/sec-pager"
				class="font-mono text-[12px]"
			/>
		</Field.Field>
		<Field.Field class="gap-1.5 space-y-0">
			<Field.FieldLabel for="ac-{action.id}-payload" class={labelClass}>Payload</Field.FieldLabel>
			<Textarea
				id="ac-{action.id}-payload"
				rows={2}
				bind:value={action.config.payload}
				class="font-mono text-[12px]"
			/>
		</Field.Field>
	</div>
{:else if action.type === 'followup'}
	<div class="flex flex-wrap gap-2">
		<Field.Field class="min-w-[200px] flex-1 gap-1.5 space-y-0">
			<Field.FieldLabel for="ac-{action.id}-title" class={labelClass}>Title</Field.FieldLabel>
			<Input
				id="ac-{action.id}-title"
				bind:value={action.config.title}
				placeholder="Write postmortem for {'{id}'}"
			/>
		</Field.Field>
		<WfSelect
			size="sm"
			label="Owner"
			options={FOLLOWUP_OWNERS}
			bind:value={action.config.owner}
			class="w-[160px]"
		/>
		<WfSelect
			size="sm"
			label="Due"
			options={FOLLOWUP_DUES}
			bind:value={action.config.due}
			class="w-[180px]"
		/>
	</div>
{:else if action.type === 'statuspage'}
	<WfSelect
		size="sm"
		label="Status page"
		options={STATUS_PAGE_OPTIONS}
		bind:value={action.config.page}
		class="w-[220px]"
	/>
{/if}
