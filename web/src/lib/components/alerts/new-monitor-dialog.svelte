<script lang="ts">
	import CheckIcon from '@lucide/svelte/icons/check';
	import CopyIcon from '@lucide/svelte/icons/copy';
	import HeartPulseIcon from '@lucide/svelte/icons/heart-pulse';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import PolicyField from '$lib/components/alertsources/policy-field.svelte';
	import { GRACE_PERIODS, INTERVALS } from '$lib/alerts';
	import { ws } from '$lib/navigation';

	let {
		open = $bindable(false),
		knownPolicies = [],
		defaultPolicy
	}: { open?: boolean; knownPolicies?: string[]; defaultPolicy: string } = $props();

	let name = $state('');
	let interval = $state('300');
	let grace = $state('120');
	let policy = $state('');

	let checkInUrl = $state<string | null>(null);
	let copied = $state(false);

	$effect(() => {
		if (open && !policy) policy = defaultPolicy;
	});

	const label = (options: { value: string; label: string }[], value: string) =>
		options.find((option) => option.value === value)?.label ?? value;

	async function copy() {
		if (!checkInUrl) return;
		await navigator.clipboard.writeText(checkInUrl);
		copied = true;
	}

	function reset() {
		open = false;
		checkInUrl = null;
		copied = false;
		name = '';
		policy = '';
	}
</script>

<Dialog.Root
	bind:open
	onOpenChangeComplete={(next) => {
		if (!next) reset();
	}}
>
	<Dialog.Content class="sm:max-w-[480px]">
		{#if checkInUrl}
			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-success-wash text-success-ink flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						<CheckIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">
							Monitor created
						</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							Have the job call this URL every time it finishes.
						</Dialog.Description>
					</div>
				</div>

				<div class="mt-1 flex flex-col gap-2">
					<div class="bg-inset flex items-center gap-2 rounded-lg border p-2.5 pl-3">
						<code class="min-w-0 flex-1 truncate font-mono text-[12.5px]">{checkInUrl}</code>
						<Button variant="secondary" size="sm" onclick={copy}>
							{#if copied}
								<CheckIcon data-icon="inline-start" />
								Copied
							{:else}
								<CopyIcon data-icon="inline-start" />
								Copy
							{/if}
						</Button>
					</div>
					<p class="text-subtle-foreground m-0 text-xs">
						A GET or POST counts as a check-in. The URL is shown once. Copy it now.
					</p>
				</div>
			</div>

			<div class="flex justify-end border-t bg-black/20 px-6 py-4">
				<Button onclick={reset}>Done</Button>
			</div>
		{:else}
			<form
				method="POST"
				action="{ws('/alerts/heartbeats')}?/create"
				use:enhance={() => async ({ result, update }) => {
					await update({ reset: false, invalidateAll: true });
					if (result.type === 'failure') {
						toast.error(String(result.data?.error ?? 'Could not create that monitor.'));
						return;
					}
					if (result.type === 'success' && result.data?.url) {
						checkInUrl = String(result.data.url);
					}
				}}
			>
				<div class="flex flex-col gap-3 p-6">
					<div class="flex items-start gap-3">
						<span
							class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
						>
							<HeartPulseIcon class="size-5" />
						</span>
						<div class="flex flex-1 flex-col gap-1">
							<Dialog.Title class="tracking-heading text-xl font-semibold">
								New heartbeat monitor
							</Dialog.Title>
							<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
								Miss the interval plus the grace period and it pages, like any other alert.
							</Dialog.Description>
						</div>
					</div>

					<div class="mt-1 flex flex-col gap-4">
						<Field.Field class="gap-1.5 space-y-0">
							<Field.FieldLabel for="name" class="text-muted-foreground text-[13px] font-medium">
								Name
							</Field.FieldLabel>
							<Input id="name" name="name" bind:value={name} placeholder="nightly-backup" required />
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
												<Select.Item value={entry.value} label={entry.label}>
													{entry.label}
												</Select.Item>
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
												<Select.Item value={entry.value} label={entry.label}>
													{entry.label}
												</Select.Item>
											{/each}
										</Select.Group>
									</Select.Content>
								</Select.Root>
							</Field.Field>
						</div>

						<PolicyField
							id="monitor-policy"
							label="Route to"
							known={knownPolicies}
							bind:value={policy}
							description="A missed check-in pages whoever this policy names. Routing rules still run first."
						/>
						<input type="hidden" name="policy" value={policy} />
					</div>
				</div>

				<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
					<Button type="button" variant="ghost" onclick={() => (open = false)}>Cancel</Button>
					<Button type="submit" disabled={!name.trim() || !policy.trim()}>Create monitor</Button>
				</div>
			</form>
		{/if}
	</Dialog.Content>
</Dialog.Root>
