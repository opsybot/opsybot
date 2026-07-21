<script lang="ts">
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import ChevronUpIcon from '@lucide/svelte/icons/chevron-up';
	import XIcon from '@lucide/svelte/icons/x';
	import { Button } from '$lib/components/ui/button';
	import * as Select from '$lib/components/ui/select';
	import {
		CHANNEL_TYPES,
		DELAY_OPTIONS,
		MAX_STEPS,
		channelLabel,
		normalizeSteps,
		uid,
		type RuleStep
	} from '$lib/notifications';

	let { steps = $bindable(), label }: { steps: RuleStep[]; label: string } = $props();

	const delayLabel = (value: string) => DELAY_OPTIONS.find((delay) => delay.value === value)?.label ?? value;
	const atCap = $derived(steps.length >= MAX_STEPS);

	function commit(next: RuleStep[]) {
		steps = normalizeSteps(next);
	}
	function update(index: number, patch: Partial<RuleStep>) {
		commit(steps.map((step, other) => (other === index ? { ...step, ...patch } : step)));
	}
	function move(index: number, dir: -1 | 1) {
		const to = index + dir;
		if (to < 0 || to >= steps.length) return;
		const next = steps.slice();
		[next[index], next[to]] = [next[to], next[index]];
		commit(next);
	}
	function remove(index: number) {
		commit(steps.filter((_, other) => other !== index));
	}
	function add() {
		if (atCap) return;
		commit([...steps, { id: uid(), channel: 'email', delay: '5' }]);
	}
</script>

<div class="flex flex-col gap-2">
	{#each steps as step, index (step.id)}
		<div class="bg-inset flex items-center gap-2.5 rounded-md border px-2.5 py-2">
			<span
				class="bg-brand-wash border-brand-edge text-brand-foreground flex size-5 shrink-0 items-center justify-center rounded-full border font-mono text-[10.5px] font-semibold"
			>
				{index + 1}
			</span>
			<Select.Root
				type="single"
				value={step.channel}
				onValueChange={(value) => update(index, { channel: value as RuleStep['channel'] })}
			>
				<Select.Trigger size="sm" class="w-[170px]" aria-label="{label} step {index + 1} channel">
					{channelLabel(step.channel)}
				</Select.Trigger>
				<Select.Content>
					<Select.Group>
						{#each CHANNEL_TYPES as channel (channel.id)}
							<Select.Item value={channel.id} label={channel.label}>{channel.label}</Select.Item>
						{/each}
					</Select.Group>
				</Select.Content>
			</Select.Root>
			{#if index === 0}
				<span class="text-subtle-foreground text-[12.5px]">immediately</span>
			{:else}
				<Select.Root
					type="single"
					value={step.delay}
					onValueChange={(value) => update(index, { delay: value })}
				>
					<Select.Trigger size="sm" class="w-[210px]" aria-label="{label} step {index + 1} delay">
						{delayLabel(step.delay)} (after step {index})
					</Select.Trigger>
					<Select.Content>
						<Select.Group>
							{#each DELAY_OPTIONS.slice(1) as delay (delay.value)}
								<Select.Item value={delay.value} label="{delay.label} (after step {index})">
									{delay.label} (after step {index})
								</Select.Item>
							{/each}
						</Select.Group>
					</Select.Content>
				</Select.Root>
			{/if}
			<div class="flex-1"></div>
			<Button
				variant="ghost"
				size="icon-sm"
				aria-label="Move step {index + 1} up"
				disabled={index === 0}
				onclick={() => move(index, -1)}
			>
				<ChevronUpIcon />
			</Button>
			<Button
				variant="ghost"
				size="icon-sm"
				aria-label="Move step {index + 1} down"
				disabled={index === steps.length - 1}
				onclick={() => move(index, 1)}
			>
				<ChevronDownIcon />
			</Button>
			<Button
				variant="ghost"
				size="icon-sm"
				aria-label="Remove step {index + 1}"
				disabled={steps.length === 1}
				onclick={() => remove(index)}
			>
				<XIcon />
			</Button>
		</div>
	{/each}
	<button
		type="button"
		onclick={add}
		disabled={atCap}
		class="text-muted-foreground hover:text-brand-foreground self-start text-[12.5px] transition-colors disabled:pointer-events-none disabled:opacity-45"
	>
		+ Add step
	</button>
</div>
