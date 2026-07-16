<script lang="ts">
	import { untrack } from 'svelte';
	import EyeOffIcon from '@lucide/svelte/icons/eye-off';
	import SirenIcon from '@lucide/svelte/icons/siren';
	import { enhance } from '$app/forms';
	import Tag from '$lib/components/tag.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as RadioGroup from '$lib/components/ui/radio-group';
	import * as Select from '$lib/components/ui/select';
	import type { Severity } from '$lib/dashboard';
	import { PEOPLE, SERVICES, SEVERITIES, type LinkedAlert } from '$lib/incidents';

	let {
		open = $bindable(false),
		openAlerts
	}: {
		open?: boolean;
		openAlerts: LinkedAlert[];
	} = $props();

	let name = $state('');
	let severity = $state<Severity>('SEV2');
	let lead = $state(PEOPLE[0]);
	let services = $state(new Set<string>());
	let linked = $state(new Set<string>(untrack(() => openAlerts).map((alert) => alert.id)));

	function toggle(set: Set<string>, value: string) {
		if (set.has(value)) set.delete(value);
		else set.add(value);
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[560px]">
		<form
			method="POST"
			action="/incidents?/declare"
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

					<div>
						<div class="text-subtle-foreground tracking-label mb-2 text-[11px] uppercase">
							Affected services
						</div>
						<div class="flex flex-wrap gap-1.5">
							{#each SERVICES as service (service)}
								<Tag selected={services.has(service)} onclick={() => toggle(services, service)}>
									{service}
								</Tag>
							{/each}
						</div>
						{#each services as service (service)}
							<input type="hidden" name="services" value={service} />
						{/each}
					</div>

					<Field.Field class="max-w-[260px] gap-1.5 space-y-0">
						<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
							Incident lead
						</Field.FieldLabel>
						<Select.Root type="single" bind:value={lead} name="lead">
							<Select.Trigger>{lead}</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each PEOPLE as person (person)}
										<Select.Item value={person} label={person}>{person}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
						<Field.FieldDescription class="text-subtle-foreground text-xs">
							Defaults to you. Hand over any time.
						</Field.FieldDescription>
					</Field.Field>

					{#if openAlerts.length}
						<div>
							<div class="text-subtle-foreground tracking-label mb-2 text-[11px] uppercase">
								Link open alerts (optional)
							</div>
							<div class="flex flex-col gap-2">
								{#each openAlerts as alert (alert.id)}
									<Field.Field orientation="horizontal" class="items-start gap-2.5 space-y-0">
										<Checkbox
											id={alert.id}
											checked={linked.has(alert.id)}
											onCheckedChange={() => toggle(linked, alert.id)}
											class="mt-0.5"
										/>
										<Field.FieldContent class="gap-0.5">
											<Field.FieldLabel for={alert.id} class="text-foreground text-sm font-normal">
												{alert.title}
											</Field.FieldLabel>
											<Field.FieldDescription class="text-subtle-foreground text-[13px]">
												{alert.severity} · {alert.status}
											</Field.FieldDescription>
										</Field.FieldContent>
									</Field.Field>
								{/each}
							</div>
							{#each linked as alertId (alertId)}
								<input type="hidden" name="alerts" value={alertId} />
							{/each}
						</div>
					{/if}

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
