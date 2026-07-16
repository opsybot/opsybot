<script lang="ts">
	import { untrack } from 'svelte';
	import { superForm, type SuperValidated } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import TextField from '$lib/components/text-field.svelte';
	import { Button } from '$lib/components/ui/button';
	import { signupAccountSchema } from '$lib/schemas/auth';
	import type { Infer } from 'sveltekit-superforms';

	let { data }: { data: SuperValidated<Infer<typeof signupAccountSchema>> } = $props();

	const form = superForm(untrack(() => data), { validators: zod4Client(signupAccountSchema) });
</script>

<form method="POST" action="?/account" use:form.enhance class="flex flex-col gap-4">
	<TextField {form} name="name" label="Name" placeholder="Maya Chen" autocomplete="name" />
	<TextField
		{form}
		name="email"
		label="Work email"
		type="email"
		placeholder="you@company.com"
		autocomplete="email"
	/>
	<TextField
		{form}
		name="password"
		label="Password"
		type="password"
		placeholder="12+ characters"
		hint="12 characters minimum."
		autocomplete="new-password"
	/>
	<Button type="submit" class="w-full">Continue</Button>
</form>
