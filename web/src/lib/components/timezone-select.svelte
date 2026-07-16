<script lang="ts" module>
	function timezones(): string[] {
		const supported = Intl.supportedValuesOf?.('timeZone');
		return supported?.length ? supported : ['UTC'];
	}
</script>

<script lang="ts" generics="T extends Record<string, unknown>, U extends FormPath<T>">
	import { untrack } from 'svelte';
	import CheckIcon from '@lucide/svelte/icons/check';
	import ChevronsUpDownIcon from '@lucide/svelte/icons/chevrons-up-down';
	import type { FormPath, SuperForm } from 'sveltekit-superforms';
	import * as Command from '$lib/components/ui/command';
	import * as Form from '$lib/components/ui/form';
	import * as Popover from '$lib/components/ui/popover';
	import { cn } from '$lib/utils';

	let {
		form,
		name,
		label
	}: {
		form: SuperForm<T>;
		name: U;
		label: string;
	} = $props();

	const { form: formData } = untrack(() => form);

	let open = $state(false);
	const options = timezones();

	function select(zone: string) {
		($formData[name] as unknown) = zone;
		open = false;
	}
</script>

<Form.Field {form} {name} class="flex flex-col gap-1.5 space-y-0">
	<Form.Control>
		{#snippet children({ props })}
			<Form.Label class="text-muted-foreground text-[13px] font-medium">{label}</Form.Label>

			<input type="hidden" {...props} value={$formData[name]} />

			<Popover.Root bind:open>
				<Popover.Trigger
					class={cn(
						'bg-inset border-border-strong text-foreground flex h-10 w-full items-center gap-2 rounded-sm border px-3 text-sm outline-none transition-[border-color,box-shadow] duration-[120ms] ease-out',
						'data-[state=open]:border-primary data-[state=open]:shadow-[var(--focus-ring)] focus-visible:border-primary focus-visible:shadow-[var(--focus-ring)]'
					)}
				>
					<span class="flex-1 truncate text-left">{$formData[name]}</span>
					<ChevronsUpDownIcon class="text-subtle-foreground size-[15px] shrink-0" />
				</Popover.Trigger>

				<Popover.Content
					align="start"
					sideOffset={6}
					class="border-input shadow-pop w-(--bits-popover-anchor-width) rounded-md border p-[5px] ring-0"
				>
					<Command.Root class="rounded-none bg-transparent p-0">
						<Command.Input placeholder="Search timezones" />
						<Command.List class="max-h-[220px]">
							<Command.Empty class="text-subtle-foreground px-2.5 py-2.5 text-left text-[13px]">
								No results.
							</Command.Empty>
							<Command.Group>
								{#each options as zone (zone)}
									<Command.Item
										value={zone}
										onSelect={() => select(zone)}
										class="rounded-xs gap-2 px-2.5 py-2 text-sm"
									>
										<span class="flex w-4 shrink-0 items-center">
											{#if zone === $formData[name]}
												<CheckIcon class="text-primary size-[15px]" />
											{/if}
										</span>
										{zone}
									</Command.Item>
								{/each}
							</Command.Group>
						</Command.List>
					</Command.Root>
				</Popover.Content>
			</Popover.Root>
		{/snippet}
	</Form.Control>

	<Form.FieldErrors class="text-critical-ink text-xs font-normal" />
</Form.Field>
