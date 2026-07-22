<script lang="ts">
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import PolicyField from '$lib/components/alertsources/policy-field.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { GRACE_PERIODS, INTERVALS, type Heartbeat } from '$lib/alerts';

	let {
		open = $bindable(false),
		monitor,
		knownPolicies = []
	}: { open?: boolean; monitor: Heartbeat | null; knownPolicies?: string[] } = $props();

	let name = $state('');
	let interval = $state('300');
	let grace = $state('120');
	let policy = $state('');

	$effect(() => {
		if (!open || !monitor) return;
		const current = monitor;
		untrack(() => {
			name = current.name;
			interval = String(current.intervalSeconds);
			grace = String(current.graceSeconds);
			policy = current.policy;
		});
	});

	const label = (options: { value: string; label: string }[], value: string) =>
		options.find((option) => option.value === value)?.label ?? `${value} s`;
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[460px]">
		<form
			method="POST"
			action="?/update"
			use:enhance={() => async ({ result, update }) => {
				await update({ invalidateAll: true });
				if (result.type === 'failure') {
					toast.error(String(result.data?.error ?? 'Could not save that monitor.'));
					return;
				}
				if (result.type === 'success') {
					open = false;
					toast.success(`${name} saved.`);
				}
			}}
		>
			<div class="flex flex-col gap-4 p-6">
				<div class="flex flex-col gap-1">
					<Dialog.Title class="tracking-heading text-xl font-semibold">Edit monitor</Dialog.Title>
					<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
						The check-in URL stays the same.
					</Dialog.Description>
				</div>

				<input type="hidden" name="id" value={monitor?.id ?? ''} />

				<Field.Field class="gap-1.5 space-y-0">
					<Field.FieldLabel for="edit-monitor-name" class="text-muted-foreground text-[13px] font-medium">
						Name
					</Field.FieldLabel>
					<Input id="edit-monitor-name" name="name" bind:value={name} required />
				</Field.Field>

				<div class="flex gap-2.5">
					<Field.Field class="flex-1 gap-1.5 space-y-0">
						<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
							Expected interval
						</Field.FieldLabel>
						<Select.Root type="single" name="interval" bind:value={interval}>
							<Select.Trigger>{label(INTERVALS, interval)}</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each INTERVALS as entry (entry.value)}
										<Select.Item value={entry.value} label={entry.label}>{entry.label}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</Field.Field>

					<Field.Field class="flex-1 gap-1.5 space-y-0">
						<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
							Grace period
						</Field.FieldLabel>
						<Select.Root type="single" name="grace" bind:value={grace}>
							<Select.Trigger>{label(GRACE_PERIODS, grace)}</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each GRACE_PERIODS as entry (entry.value)}
										<Select.Item value={entry.value} label={entry.label}>{entry.label}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</Field.Field>
				</div>

				<PolicyField id="edit-monitor-policy" label="Route to" known={knownPolicies} bind:value={policy} />
				<input type="hidden" name="policy" value={policy} />
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (open = false)}>Cancel</Button>
				<Button type="submit" disabled={!name.trim() || !policy.trim()}>Save monitor</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
