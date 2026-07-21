<script lang="ts">
	import BracesIcon from '@lucide/svelte/icons/braces';
	import CheckIcon from '@lucide/svelte/icons/check';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import EyeOffIcon from '@lucide/svelte/icons/eye-off';
	import XIcon from '@lucide/svelte/icons/x';
	import * as Select from '$lib/components/ui/select';
	import { AI_FEATURES, AI_PROMPTS, featureLabel, type AiFeatureId } from '$lib/ai';

	let feature = $state<AiFeatureId>('summaries');
	const prompt = $derived(AI_PROMPTS[feature]);
</script>

<div class="flex max-w-[760px] flex-col gap-3.5">
	<div class="flex flex-wrap items-center gap-2.5">
		<Select.Root type="single" value={feature} onValueChange={(value) => (feature = value as AiFeatureId)}>
			<Select.Trigger size="sm" class="w-[210px]" aria-label="Feature">{featureLabel(feature)}</Select.Trigger>
			<Select.Content>
				<Select.Group>
					{#each AI_FEATURES as option (option.id)}
						<Select.Item value={option.id} label={option.label}>{option.label}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
		<span class="text-subtle-foreground text-[12.5px]">
			Read-only: exactly what leaves for the model, nothing else.
		</span>
	</div>

	<div class="grid gap-3.5 sm:grid-cols-2">
		<div class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<EyeIcon class="text-subtle-foreground size-3.5" />
				<span class="text-[13px] font-semibold">Data included</span>
			</header>
			<ul class="m-0 list-none py-2 pl-0">
				{#each prompt.fields as field (field)}
					<li class="text-muted-foreground flex items-center gap-[9px] border-t px-4 py-[7px] text-[12.5px] first:border-t-0">
						<CheckIcon class="text-primary size-3 shrink-0" />
						{field}
					</li>
				{/each}
			</ul>
		</div>
		<div class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<EyeOffIcon class="text-subtle-foreground size-3.5" />
				<span class="text-[13px] font-semibold">Never included</span>
			</header>
			<ul class="m-0 list-none py-2 pl-0">
				{#each prompt.excluded as field (field)}
					<li class="text-muted-foreground flex items-center gap-[9px] border-t px-4 py-[7px] text-[12.5px] first:border-t-0">
						<XIcon class="size-3 shrink-0 text-[var(--critical)]" />
						{field}
					</li>
				{/each}
			</ul>
		</div>
	</div>

	<div class="bg-card overflow-hidden rounded-xl border">
		<header class="flex items-center gap-2 border-b px-4 py-3">
			<BracesIcon class="text-subtle-foreground size-3.5" />
			<span class="text-[13px] font-semibold">Prompt template</span>
			<span class="text-subtle-foreground ml-auto font-mono text-[10.5px]">
				{'{{fields}}'} are filled at run time
			</span>
		</header>
		<pre
			class="text-muted-foreground m-0 overflow-x-auto bg-[var(--ink-0)] px-4 py-4 font-mono text-[12px] leading-[1.7] whitespace-pre-wrap [overflow-wrap:anywhere]">{prompt.template}</pre>
	</div>
</div>
