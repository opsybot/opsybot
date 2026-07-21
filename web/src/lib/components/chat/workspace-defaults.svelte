<script lang="ts">
	import { tick, untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { Switch } from '$lib/components/ui/switch';
	import { ANNOUNCE_CHANNELS, DEFAULT_DEFAULTS, previewChannelName, type Platform } from '$lib/chat';

	let { platform }: { platform: Platform } = $props();

	const initial = untrack(() => platform.connection?.defaults ?? DEFAULT_DEFAULTS);
	let naming = $state(initial.namingPattern);
	let announce = $state(initial.announceChannel);
	let archive = $state(initial.archiveOnResolve);

	const preview = $derived(previewChannelName(naming));

	let namingForm: HTMLFormElement;
	let announceForm: HTMLFormElement;
	let archiveForm: HTMLFormElement;

	async function changeAnnounce(value: string) {
		announce = value;
		await tick();
		announceForm.requestSubmit();
	}
</script>

<div class="flex flex-col gap-3.5 border-t px-4 py-[14px]">
	<div class="text-subtle-foreground -mb-1 text-[11px] tracking-[0.08em] uppercase">
		Defaults for this workspace
	</div>

	<div class="flex flex-wrap items-start gap-3">
		<form
			method="POST"
			action="?/saveNaming"
			bind:this={namingForm}
			class="w-[230px]"
			use:enhance={() => async ({ update }) => update({ reset: false })}
		>
			<input type="hidden" name="platform" value={platform.id} />
			<Field.Field class="gap-1.5 space-y-0">
				<Field.FieldLabel for="chan-{platform.id}" class="text-muted-foreground text-[13px] font-medium">
					Incident channel naming
				</Field.FieldLabel>
				<Input
					id="chan-{platform.id}"
					name="pattern"
					class="font-mono text-[12.5px]"
					bind:value={naming}
					onchange={() => namingForm.requestSubmit()}
				/>
				<Field.FieldDescription class="text-subtle-foreground text-xs">
					Preview: {preview}
				</Field.FieldDescription>
			</Field.Field>
		</form>

		<form
			method="POST"
			action="?/setAnnounce"
			bind:this={announceForm}
			class="w-[230px]"
			use:enhance={() => async ({ update }) => update({ reset: false })}
		>
			<input type="hidden" name="platform" value={platform.id} />
			<input type="hidden" name="channel" value={announce} />
			<Field.Field class="gap-1.5 space-y-0">
				<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
					Default announcement channel
				</Field.FieldLabel>
				<Select.Root type="single" value={announce} onValueChange={changeAnnounce}>
					<Select.Trigger class="w-full" aria-label="Default announcement channel">{announce}</Select.Trigger>
					<Select.Content>
						<Select.Group>
							{#each ANNOUNCE_CHANNELS as channel (channel)}
								<Select.Item value={channel} label={channel}>{channel}</Select.Item>
							{/each}
						</Select.Group>
					</Select.Content>
				</Select.Root>
			</Field.Field>
		</form>
	</div>

	<form
		method="POST"
		action="?/setArchive"
		bind:this={archiveForm}
		use:enhance={() =>
			async ({ result, update }) => {
				await update({ reset: false });
				if (result.type !== 'success') {
					archive = !archive;
					return;
				}
				toast(
					archive
						? 'Incident channels will be archived on resolve.'
						: 'Incident channels stay open after resolve.'
				);
			}}
	>
		<input type="hidden" name="platform" value={platform.id} />
		<input type="hidden" name="archive" value={String(archive)} />
		<div class="flex items-center gap-2.5">
			<Switch
				bind:checked={archive}
				aria-label="Archive the incident channel when the incident resolves"
				onCheckedChange={async () => {
					await tick();
					archiveForm.requestSubmit();
				}}
			/>
			<span class="text-muted-foreground text-[13px]">
				Archive the incident channel when the incident resolves
			</span>
		</div>
	</form>
</div>
