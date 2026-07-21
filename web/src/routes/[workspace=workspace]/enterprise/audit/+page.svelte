<script lang="ts">
	import { onDestroy, tick, untrack } from 'svelte';
	import CheckIcon from '@lucide/svelte/icons/check';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import FilterIcon from '@lucide/svelte/icons/filter';
	import SendIcon from '@lucide/svelte/icons/send';
	import { toast } from 'svelte-sonner';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import UpgradeState from '$lib/components/enterprise/upgrade-state.svelte';
	import { ENT_PITCH, FORMAT_OPTIONS, RETENTION_OPTIONS } from '$lib/enterprise';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let endpoint = $state(untrack(() => data.audit?.streamEndpoint ?? ''));
	let format = $state(untrack(() => data.audit?.format ?? 'JSON lines'));
	let retention = $state(untrack(() => data.audit?.retention ?? '7 years'));

	let streamTest = $state<'idle' | 'running' | 'ok'>('idle');
	let testTimer: ReturnType<typeof setTimeout>;
	let okAlert = $state<HTMLElement | null>(null);
	$effect(() => {
		void [endpoint, format];
		untrack(() => {
			if (streamTest !== 'idle') {
				clearTimeout(testTimer);
				streamTest = 'idle';
			}
		});
	});
	function runTest() {
		if (streamTest === 'running') return;
		streamTest = 'running';
		clearTimeout(testTimer);
		testTimer = setTimeout(async () => {
			streamTest = 'ok';
			toast.success('Test event delivered: 200 in 0.4 s.');
			await tick();
			okAlert?.focus();
		}, 1500);
	}
	onDestroy(() => clearTimeout(testTimer));
</script>

{#snippet labeledSelect(label: string, current: string, options: string[], onpick: (value: string) => void, width: string)}
	<div class="flex flex-col gap-1.5" style="width:{width}">
		<span class="text-muted-foreground text-[13px] font-medium">{label}</span>
		<Select.Root type="single" value={current} onValueChange={onpick}>
			<Select.Trigger size="sm" aria-label={label}>{current}</Select.Trigger>
			<Select.Content>
				<Select.Group>
					{#each options as option (option)}
						<Select.Item value={option} label={option}>{option}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
	</div>
{/snippet}

{#if !data.licensed}
	<UpgradeState pitch={ENT_PITCH.audit} />
{:else if data.audit}
	<div class="flex max-w-[760px] flex-col gap-3.5">
		<div class="bg-card flex flex-col gap-3 rounded-xl border p-4">
			<div class="flex flex-wrap items-center gap-2">
				<span class="text-[13.5px] font-semibold">Saved filters</span>
				<span class="text-subtle-foreground text-[11.5px]">one click back to a saved audit view</span>
				<div class="flex-1"></div>
				<Button size="sm" variant="ghost" onclick={() => toast.success('Export started: filtered events as JSONL.')}>
					<DownloadIcon data-icon="inline-start" />
					Export JSONL
				</Button>
				<Button size="sm" variant="ghost" onclick={() => toast.success('Export started: filtered events as CSV.')}>
					<DownloadIcon data-icon="inline-start" />
					Export CSV
				</Button>
			</div>
			<div class="flex flex-wrap gap-1.5">
				{#each data.audit.savedFilters as filter (filter.name)}
					<a
						href={ws('/workspace/audit')}
						class="bg-inset text-muted-foreground hover:text-brand-foreground hover:border-brand-edge inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-[12px] transition-colors"
					>
						<FilterIcon class="size-[11px]" />
						{filter.name}
						<span class="text-subtle-foreground font-mono text-[10px]">{filter.q}</span>
					</a>
				{/each}
			</div>
			<p class="text-subtle-foreground m-0 text-[11.5px]">
				The event list itself lives in
				<a href={ws('/workspace/audit')} class="text-brand-foreground hover:underline">Workspace admin → Audit log</a>. These
				tools extend it.
			</p>
		</div>

		<div class="bg-card flex flex-col gap-3.5 rounded-xl border p-4">
			<span class="text-[13.5px] font-semibold">Streaming destination</span>
			<div class="flex flex-wrap items-end gap-2.5">
				<Field.Field class="min-w-[260px] flex-1 gap-1.5 space-y-0">
					<Field.FieldLabel for="stream-endpoint" class="text-muted-foreground text-[13px] font-medium">Endpoint</Field.FieldLabel>
					<Input id="stream-endpoint" class="font-mono text-[12px]" bind:value={endpoint} />
				</Field.Field>
				{@render labeledSelect('Format', format, FORMAT_OPTIONS, (value) => (format = value), '150px')}
			</div>
			{#if streamTest === 'ok'}
				<Alert.Root tone="success" bind:ref={okAlert} tabindex={-1} class="outline-none">
					<CheckIcon />
					<Alert.Content>
						<Alert.Description>
							Streaming live: every audit event ships within seconds. Failures retry with backoff for 24 h.
						</Alert.Description>
					</Alert.Content>
				</Alert.Root>
			{:else}
				<div class="flex items-center gap-2.5" role="status" aria-live="polite">
					<Button
						size="sm"
						variant="secondary"
						aria-disabled={streamTest === 'running'}
						class="aria-disabled:pointer-events-none aria-disabled:opacity-60"
						onclick={runTest}
					>
						<SendIcon data-icon="inline-start" />
						{streamTest === 'running' ? 'Delivering test event…' : 'Send test delivery'}
					</Button>
					<span class="text-subtle-foreground text-[11.5px]">Required before streaming turns on.</span>
				</div>
			{/if}
			<div class="flex items-center gap-2.5">
				{@render labeledSelect(
					'Retention',
					retention,
					RETENTION_OPTIONS,
					(value) => {
						retention = value;
						toast(`Audit retention set to ${value}.`);
					},
					'140px'
				)}
				<span class="text-subtle-foreground pt-[22px] text-[11.5px]">Streamed copies are yours regardless of retention here.</span>
			</div>
		</div>
	</div>
{/if}
