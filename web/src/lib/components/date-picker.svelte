<script lang="ts">
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import { CalendarDate, parseDate, type DateValue } from '@internationalized/date';
	import { Calendar } from '$lib/components/ui/calendar';
	import * as Popover from '$lib/components/ui/popover';
	import { cn } from '$lib/utils';

	let {
		value = $bindable(''),
		name,
		id,
		label,
		invalid = false,
		size = 'default',
		class: className,
		onChange
	}: {
		value?: string;
		name?: string;
		id?: string;
		label?: string;
		invalid?: boolean;
		size?: 'default' | 'sm';
		class?: string;
		onChange?: (value: string) => void;
	} = $props();

	let open = $state(false);

	function toCalendarDate(iso: string): CalendarDate | undefined {
		try {
			const parsed = parseDate(iso);
			return new CalendarDate(parsed.year, parsed.month, parsed.day);
		} catch {
			return undefined;
		}
	}

	const selected = $derived(toCalendarDate(value));
	const display = $derived(selected ? selected.toString() : 'Pick a date');

	function pick(next: DateValue | undefined) {
		if (!next) return;
		value = new CalendarDate(next.year, next.month, next.day).toString();
		open = false;
		onChange?.(value);
	}
</script>

{#if name}
	<input type="hidden" {name} {value} />
{/if}

<Popover.Root bind:open>
	<Popover.Trigger
		{id}
		type="button"
		aria-label={label}
		aria-invalid={invalid ? 'true' : undefined}
		class={cn(
			'bg-inset border-border-strong text-foreground flex w-full items-center gap-2 rounded-sm border px-3 outline-none transition-[border-color,box-shadow] duration-[120ms] ease-out',
			'data-[state=open]:border-primary data-[state=open]:shadow-[var(--focus-ring)] focus-visible:border-primary focus-visible:shadow-[var(--focus-ring)]',
			'aria-[invalid=true]:border-critical',
			size === 'sm' ? 'h-[34px] text-[13px]' : 'h-10 text-sm',
			className
		)}
	>
		<span class={cn('flex-1 truncate text-left', !selected && 'text-subtle-foreground')}>
			{display}
		</span>
		<CalendarIcon class="text-subtle-foreground size-[15px] shrink-0" />
	</Popover.Trigger>

	<Popover.Content align="start" sideOffset={6} class="border-input shadow-pop w-auto rounded-md border p-0">
		<Calendar type="single" value={selected} onValueChange={pick} captionLayout="dropdown" />
	</Popover.Content>
</Popover.Root>
