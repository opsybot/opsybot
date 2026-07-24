<script lang="ts">
	import EyeOffIcon from '@lucide/svelte/icons/eye-off';
	import SirenIcon from '@lucide/svelte/icons/siren';
	import { enhance } from '$app/forms';
	import Tag from '$lib/components/tag.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as RadioGroup from '$lib/components/ui/radio-group';
	import * as Select from '$lib/components/ui/select';
	import type { Severity } from '$lib/dashboard';
	import { SEVERITIES } from '$lib/incidents';
	import { ws } from '$lib/navigation';

	let {
		open = $bindable(false),
		services = [],
		members = []
	}: {
		open?: boolean;
		services?: { id: string; name: string }[];
		members?: { id: string; name: string }[];
	} = $props();

	let name = $state('');
	let severity = $state<Severity>('SEV2');
	let lead = $state('');
	let selected = $state(new Set<string>());

	const leadName = $derived(members.find((member) => member.id === lead)?.name ?? 'Unassigned');

	function toggle(value: string) {
		const next = new Set(selected);
		if (next.has(value)) next.delete(value);
		else next.add(value);
		selected = next;
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[560px]">
		<form
			method="POST"
			action="{ws('/incidents')}?/declare"
			use:enhance={() => async ({ update }) => {
				await update();
				open = false;
				name = '';
			}}
		>
			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-critical-wash text-critical-ink flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						<SirenIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">
							Declare an incident
						</Dialog.Title>
						<Dialog.Description class="sr-only">
							Declaring pages the responders and opens a channel. It publishes nothing.
						</Dialog.Description>
					</div>
				</div>

				<div class="mt-1 flex flex-col gap-4">
					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel for="incident-name" class="text-muted-foreground text-[13px] font-medium">
							What's happening?
						</Field.FieldLabel>
						<Input
							id="incident-name"
							name="name"
							bind:value={name}
							placeholder="Checkout returning errors in EU"
							required
						/>
					</Field.Field>

					<div>
						<div class="text-subtle-foreground tracking-label mb-2 text-[11px] uppercase">
							Severity
						</div>
						<RadioGroup.Root bind:value={severity} name="severity" class="gap-2.5">
							{#each SEVERITIES as level (level.id)}
								<Field.Field orientation="horizontal" class="items-start gap-2.5 space-y-0">
									<RadioGroup.Item value={level.id} id={level.id} class="mt-0.5" />
									<Field.FieldContent class="gap-0.5">
										<Field.FieldLabel for={level.id} class="text-foreground text-sm font-normal">
											{level.id}
										</Field.FieldLabel>
										<Field.FieldDescription class="text-subtle-foreground text-[13px]">
											{level.definition}
										</Field.FieldDescription>
									</Field.FieldContent>
								</Field.Field>
							{/each}
						</RadioGroup.Root>
					</div>

					{#if services.length}
						<div>
							<div class="text-subtle-foreground tracking-label mb-2 text-[11px] uppercase">
								Affected services
							</div>
							<div class="flex flex-wrap gap-1.5">
								{#each services as service (service.id)}
									<Tag selected={selected.has(service.id)} onclick={() => toggle(service.id)}>
										{service.name}
									</Tag>
								{/each}
							</div>
							{#each selected as serviceId (serviceId)}
								<input type="hidden" name="services" value={serviceId} />
							{/each}
						</div>
					{/if}

					<Field.Field class="max-w-[260px] gap-1.5 space-y-0">
						<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
							Incident lead
						</Field.FieldLabel>
						<Select.Root type="single" bind:value={lead} name="lead">
							<Select.Trigger>{leadName}</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each members as member (member.id)}
										<Select.Item value={member.id} label={member.name}>{member.name}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
						<Field.FieldDescription class="text-subtle-foreground text-xs">
							Optional. Hand over any time.
						</Field.FieldDescription>
					</Field.Field>

					<Alert.Root tone="info">
						<EyeOffIcon />
						<Alert.Content>
							<Alert.Description>
								Declaring never publishes anything publicly. Status pages only change when you
								explicitly publish.
							</Alert.Description>
						</Alert.Content>
					</Alert.Root>
				</div>
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (open = false)}>Cancel</Button>
				<Button type="submit" variant="destructive" disabled={!name.trim()}>
					Declare incident
				</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
