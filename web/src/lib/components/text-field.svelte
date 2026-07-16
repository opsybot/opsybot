<script lang="ts" generics="T extends Record<string, unknown>, U extends FormPath<T>">
	import { untrack } from 'svelte';
	import type { HTMLInputAttributes, HTMLInputTypeAttribute } from 'svelte/elements';
	import type { FormPath, SuperForm } from 'sveltekit-superforms';
	import * as Form from '$lib/components/ui/form';
	import { Input } from '$lib/components/ui/input';

	type Props = Omit<HTMLInputAttributes, 'type' | 'files' | 'form' | 'name'> & {
		form: SuperForm<T>;
		name: U;
		label: string;
		hint?: string;
		type?: Exclude<HTMLInputTypeAttribute, 'file'>;
	};

	let { form, name, label, hint, type, ...rest }: Props = $props();

	const { form: formData } = untrack(() => form);
</script>

<Form.Field {form} {name} class="flex flex-col gap-1.5 space-y-0">
	{#snippet children({ errors })}
		<Form.Control>
			{#snippet children({ props })}
				<Form.Label class="text-muted-foreground text-[13px] font-medium">{label}</Form.Label>
				<Input {...props} {type} bind:value={$formData[name]} {...rest} />
			{/snippet}
		</Form.Control>

		{#if hint && errors.length === 0}
			<Form.Description class="text-subtle-foreground text-xs">{hint}</Form.Description>
		{/if}

		<Form.FieldErrors class="text-critical-ink text-xs font-normal" />
	{/snippet}
</Form.Field>
