<script lang="ts">
	import PlusIcon from '@lucide/svelte/icons/plus';
	import SparklesIcon from '@lucide/svelte/icons/sparkles';
	import { Button } from '$lib/components/ui/button';
	import AddModelDialog from '$lib/components/ai/add-model-dialog.svelte';
	import FeatureAssign from '$lib/components/ai/feature-assign.svelte';
	import MasterBanner from '$lib/components/ai/master-banner.svelte';
	import ModelRow from '$lib/components/ai/model-row.svelte';
	import { AI_FEATURES } from '$lib/ai';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let addOpen = $state(false);
</script>

<div class="flex max-w-[760px] flex-col gap-3.5">
	<MasterBanner enabled={data.enabled} />

	{#if data.models.length === 0}
		<div class="flex flex-col items-center gap-2.5 rounded-xl border border-dashed px-5 py-14 text-center">
			<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
				<SparklesIcon class="text-subtle-foreground size-5" />
			</span>
			<div class="text-[15px] font-medium">No model configured</div>
			<p class="text-subtle-foreground m-0 max-w-[440px] text-[13px] leading-[1.6]">
				AI stays off until you connect a model. Bring your own — a self-hosted Ollama keeps incident
				data entirely on your infrastructure. Data goes only to the endpoint you configure.
			</p>
			<Button size="sm" variant="secondary" onclick={() => (addOpen = true)}>
				<PlusIcon data-icon="inline-start" />
				Connect a model
			</Button>
		</div>
	{:else}
		<div class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Model endpoints</span>
				<div class="flex-1"></div>
				<Button size="sm" onclick={() => (addOpen = true)}>
					<PlusIcon data-icon="inline-start" />
					Add model
				</Button>
			</header>
			<div>
				{#each data.models as model (model.id)}
					<ModelRow {model} isDefault={model.id === data.defaultModelId} />
				{/each}
			</div>
		</div>

		<div class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Per-feature assignment</span>
				<span class="text-subtle-foreground text-[11.5px]">
					route heavy work to the big model, quick work to the fast one
				</span>
			</header>
			<div>
				{#each AI_FEATURES as feature (feature.id)}
					<FeatureAssign {feature} models={data.models} value={data.assignments[feature.id]} />
				{/each}
			</div>
		</div>
	{/if}
</div>

<AddModelDialog bind:open={addOpen} />
