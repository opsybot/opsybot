<script lang="ts">
	import CheckIcon from '@lucide/svelte/icons/check';
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { WAIT_OPTIONS } from '$lib/escalation';

	let {
		wait,
		aboveId,
		editable = false,
		oninsert,
		onchangewait
	}: {
		wait: string;
		aboveId: string;
		editable?: boolean;
		oninsert?: (afterId: string) => void;
		onchangewait?: (id: string, wait: string) => void;
	} = $props();

	const chip =
		'bg-[var(--ink-2)] text-subtle-foreground inline-flex items-center gap-[5px] rounded-full border border-dashed border-border-strong px-2.5 py-[3px] font-mono text-[10.5px]';
</script>

<div class="group flex flex-col items-center">
	<span class="bg-border-strong h-[9px] w-[1.5px]"></span>
	<div class="relative flex items-center">
		{#if editable}
			<DropdownMenu.Root>
				<DropdownMenu.Trigger
					class="{chip} hover:text-muted-foreground hover:border-subtle-foreground cursor-pointer transition-colors data-[state=open]:border-primary data-[state=open]:text-muted-foreground"
					title="Change how long to wait for an ack"
				>
					<ClockIcon class="size-[11px]" />
					wait {wait}m for ack
					<ChevronDownIcon class="size-2.5 opacity-0 transition-opacity group-hover:opacity-80" />
				</DropdownMenu.Trigger>
				<DropdownMenu.Content class="min-w-[172px]" align="center">
					<DropdownMenu.Label
						class="text-subtle-foreground px-2 pt-1.5 pb-1.5 text-[10px] tracking-[0.06em] uppercase"
					>
						Wait for acknowledgement
					</DropdownMenu.Label>
					{#each WAIT_OPTIONS as option (option.value)}
						<DropdownMenu.Item class="gap-1.5 text-[12.5px]" onSelect={() => onchangewait?.(aboveId, option.value)}>
							<span class="flex w-[15px] shrink-0 justify-center">
								{#if String(wait) === option.value}
									<CheckIcon class="text-primary size-[13px]" />
								{/if}
							</span>
							{option.label}
						</DropdownMenu.Item>
					{/each}
				</DropdownMenu.Content>
			</DropdownMenu.Root>
			<button
				type="button"
				class="hover:border-brand-edge hover:text-brand-foreground absolute top-1/2 left-full ml-2 flex size-6 -translate-y-1/2 items-center justify-center rounded-full border border-border-strong bg-[var(--surface-raised)] text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100"
				aria-label="Insert level"
				title="Insert a level here"
				onclick={() => oninsert?.(aboveId)}
			>
				<PlusIcon class="size-[13px]" />
			</button>
		{:else}
			<span class={chip}>
				<ClockIcon class="size-[11px]" />
				wait {wait}m for ack
			</span>
		{/if}
	</div>
	<span class="bg-border-strong h-[9px] w-[1.5px]"></span>
	<ChevronDownIcon class="text-subtle-foreground -mt-1 size-3.5" />
</div>
