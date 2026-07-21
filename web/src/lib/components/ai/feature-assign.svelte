<script lang="ts">
	import { tick, untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import * as Select from '$lib/components/ui/select';
	import { USE_DEFAULT, type AiFeature, type Model } from '$lib/ai';

	let { feature, models, value }: { feature: AiFeature; models: Model[]; value: string } = $props();

	let assigned = $state(untrack(() => value));
	$effect(() => {
		assigned = value;
	});
	let form: HTMLFormElement;

	const label = (id: string) =>
		id === USE_DEFAULT ? 'Use default' : (models.find((model) => model.id === id)?.name ?? id);

	async function change(next: string) {
		assigned = next;
		await tick();
		form.requestSubmit();
	}
</script>

<div class="flex items-center gap-3 border-t px-4 py-3 first:border-t-0">
	<div class="min-w-0 flex-1">
		<div class="text-[13px] font-medium">{feature.label}</div>
		<div class="text-subtle-foreground mt-px text-[11.5px]">{feature.desc}</div>
	</div>
	<form
		method="POST"
		action="?/assign"
		bind:this={form}
		use:enhance={() => async ({ result, update }) => {
			await update({ reset: false });
			if (result.type !== 'success') {
				assigned = value;
				toast.error(String((result.type === 'failure' && result.data?.error) || 'Could not change the assignment.'));
			}
		}}
	>
		<input type="hidden" name="feature" value={feature.id} />
		<input type="hidden" name="model" value={assigned} />
		<Select.Root type="single" value={assigned} onValueChange={change}>
			<Select.Trigger size="sm" class="w-[210px]" aria-label="Model for {feature.label}">
				{label(assigned)}
			</Select.Trigger>
			<Select.Content>
				<Select.Group>
					<Select.Item value={USE_DEFAULT} label="Use default">Use default</Select.Item>
					{#each models as model (model.id)}
						<Select.Item value={model.id} label={model.name}>{model.name}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
	</form>
</div>
