<script lang="ts">
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Field from '$lib/components/ui/field';
	import { Textarea } from '$lib/components/ui/textarea';

	let {
		open = $bindable(false),
		incidentId,
		linkedAlerts
	}: {
		open?: boolean;
		incidentId: string;
		linkedAlerts: number;
	} = $props();

	let summary = $state('');
	let alsoAlerts = $state(true);
	let schedulePostmortem = $state(true);
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[520px]">
		<form
			method="POST"
			action="?/resolve"
			use:enhance={() => async ({ update }) => {
				await update();
				open = false;
				summary = '';
			}}
		>
			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span
						class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg"
					>
						<CircleCheckIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">
							Resolve {incidentId}
						</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							Resolving stops update reminders and notifies responders. It doesn't publish anything.
						</Dialog.Description>
					</div>
				</div>

				<div class="mt-1 flex flex-col gap-3.5">
					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel for="summary" class="text-muted-foreground text-[13px] font-medium">
							Resolution summary
						</Field.FieldLabel>
						<Textarea
							id="summary"
							name="summary"
							rows={3}
							bind:value={summary}
							placeholder="What fixed it, and how do we know it's fixed?"
							required
						/>
						<Field.FieldDescription class="text-subtle-foreground text-xs">
							Required. Lands on the timeline and in the postmortem.
						</Field.FieldDescription>
					</Field.Field>

					{#if linkedAlerts > 0}
						<Field.Field orientation="horizontal" class="items-center gap-2.5 space-y-0">
							<Checkbox id="alerts" name="alerts" bind:checked={alsoAlerts} value="on" />
							<Field.FieldLabel for="alerts" class="text-foreground text-sm font-normal">
								Resolve the {linkedAlerts} linked alerts too
							</Field.FieldLabel>
						</Field.Field>
					{/if}

					<Field.Field orientation="horizontal" class="items-start gap-2.5 space-y-0">
						<Checkbox
							id="postmortem"
							name="postmortem"
							bind:checked={schedulePostmortem}
							value="on"
							class="mt-0.5"
						/>
						<Field.FieldContent class="gap-0.5">
							<Field.FieldLabel for="postmortem" class="text-foreground text-sm font-normal">
								Schedule the postmortem
							</Field.FieldLabel>
							<Field.FieldDescription class="text-subtle-foreground text-[13px]">
								Draft due in 3 working days, assigned to the lead.
							</Field.FieldDescription>
						</Field.FieldContent>
					</Field.Field>
				</div>
			</div>

			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (open = false)}>Cancel</Button>
				<Button type="submit" disabled={!summary.trim()}>Resolve incident</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
