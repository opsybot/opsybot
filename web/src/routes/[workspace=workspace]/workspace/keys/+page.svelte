<script lang="ts">
	import OctagonAlertIcon from '@lucide/svelte/icons/octagon-alert';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import CreateKeyDialog from '$lib/components/admin/create-key-dialog.svelte';
	import SegmentedToggle from '$lib/components/admin/segmented-toggle.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Table from '$lib/components/ui/table';
	import type { ApiKey, KeyKind } from '$lib/admin';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let kind = $state<KeyKind>('personal');
	let creating = $state(false);
	let revoking = $state(false);
	let target = $state<ApiKey | null>(null);

	const keys = $derived(kind === 'personal' ? data.keys.personal : data.keys.workspace);
	const note = $derived(
		kind === 'personal' ? 'act as you, die with your account' : 'act as the workspace, admin-managed'
	);
</script>

<div class="flex max-w-[860px] flex-col gap-3.5">
	<div class="flex flex-wrap items-center gap-2.5">
		<SegmentedToggle
			bind:value={kind}
			label="Key scope"
			options={[
				{ value: 'personal', label: 'Personal keys' },
				{ value: 'workspace', label: 'Workspace keys' }
			]}
		/>
		<span class="text-subtle-foreground text-[12px]">{note}</span>
		<div class="flex-1"></div>
		<Button size="sm" onclick={() => (creating = true)}>
			<PlusIcon data-icon="inline-start" />
			Create key
		</Button>
	</div>

	<section class="bg-card overflow-hidden rounded-xl border">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head class="pl-[18px]">Key</Table.Head>
					<Table.Head>Scopes</Table.Head>
					<Table.Head class="w-[180px]">Last used</Table.Head>
					<Table.Head class="w-[120px]">Created</Table.Head>
					<Table.Head class="w-[110px] pr-[18px]"></Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each keys as key (key.id)}
					<Table.Row data-key={key.id}>
						<Table.Cell class="text-foreground pl-[18px] font-mono text-[12.5px] font-medium">{key.name}</Table.Cell>
						<Table.Cell>
							<div class="flex flex-wrap gap-1.5">
								{#each key.scopes as scope (scope)}<Tag>{scope}</Tag>{/each}
							</div>
						</Table.Cell>
						<Table.Cell class="text-muted-foreground font-mono text-[12px]">{key.last}</Table.Cell>
						<Table.Cell class="text-muted-foreground font-mono text-[12px]">{key.created}</Table.Cell>
						<Table.Cell class="pr-[18px] text-right">
							<Button size="sm" variant="ghost" onclick={() => { target = key; revoking = true; }}>
								<Trash2Icon data-icon="inline-start" />
								Revoke
							</Button>
						</Table.Cell>
					</Table.Row>
				{:else}
					<Table.Row>
						<Table.Cell colspan={5} class="text-subtle-foreground py-10 text-center text-[13px]">
							No {kind} keys yet. Create one to call the API {kind === 'personal' ? 'as you' : 'as the workspace'}.
						</Table.Cell>
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</section>
</div>

<CreateKeyDialog bind:open={creating} {kind} />

<Dialog.Root bind:open={revoking}>
	<Dialog.Content class="sm:max-w-[440px]">
		<form
			method="POST"
			action="?/revoke"
			use:enhance={() => async ({ result, update }) => {
				await update({ reset: false });
				const name = target?.name;
				revoking = false;
				if (result.type === 'success' && name) toast(`${name} revoked.`);
				else if (result.type === 'failure') toast.error(String(result.data?.error ?? 'Could not revoke the key.'));
			}}
		>
			<input type="hidden" name="id" value={target?.id ?? ''} />
			<div class="flex flex-col gap-3 p-6">
				<div class="flex items-start gap-3">
					<span class="bg-critical-wash text-critical-ink flex size-[38px] shrink-0 items-center justify-center rounded-lg">
						<OctagonAlertIcon class="size-5" />
					</span>
					<div class="flex flex-1 flex-col gap-1">
						<Dialog.Title class="tracking-heading text-xl font-semibold">Revoke {target?.name}?</Dialog.Title>
						<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
							Requests with this key start failing immediately. This cannot be undone.
						</Dialog.Description>
					</div>
				</div>
			</div>
			<div class="flex justify-end gap-2 border-t bg-black/20 px-6 py-4">
				<Button type="button" variant="ghost" onclick={() => (revoking = false)}>Cancel</Button>
				<Button type="submit" variant="destructive">Revoke key</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
