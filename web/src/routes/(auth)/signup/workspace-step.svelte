<script lang="ts">
	import { untrack } from 'svelte';
	import { superForm, type Infer, type SuperValidated } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import TextField from '$lib/components/text-field.svelte';
	import TimezoneSelect from '$lib/components/timezone-select.svelte';
	import { Button } from '$lib/components/ui/button';
	import { workspaceSchema } from '$lib/schemas/auth';

	let { data }: { data: SuperValidated<Infer<typeof workspaceSchema>> } = $props();

	const form = superForm(untrack(() => data), { validators: zod4Client(workspaceSchema) });
</script>

<form method="POST" action="?/workspace" use:form.enhance class="flex flex-col gap-4">
	<TextField
		{form}
		name="workspace"
		label="Workspace name"
		placeholder="Acme Corp"
		hint="Usually your company or team name."
	/>
	<TimezoneSelect {form} name="timezone" label="Workspace timezone" />

	<Button type="submit" class="w-full">Create workspace</Button>

	<a href="/signup" class="text-muted-foreground hover:text-brand-foreground mx-auto block text-[13px]">
		Back to account details
	</a>
</form>
