<script lang="ts">
	import ActivityIcon from '@lucide/svelte/icons/activity';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import HistoryIcon from '@lucide/svelte/icons/history';
	import ImageIcon from '@lucide/svelte/icons/image';
	import LinkIcon from '@lucide/svelte/icons/link';
	import MegaphoneIcon from '@lucide/svelte/icons/megaphone';
	import PaperclipIcon from '@lucide/svelte/icons/paperclip';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import ScaleIcon from '@lucide/svelte/icons/scale';
	import ScrollTextIcon from '@lucide/svelte/icons/scroll-text';
	import WrenchIcon from '@lucide/svelte/icons/wrench';
	import XIcon from '@lucide/svelte/icons/x';
	import { enhance } from '$app/forms';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import {
		ENTRY_TYPES,
		type Attachment,
		type AttachmentKind,
		type EntryType,
		type TimelineEntry,
		type TimelineRevision
	} from '$lib/incidents';
	import { formatUtc, formatUtcDate, formatUtcTime } from '$lib/time';

	let {
		entry,
		last,
		showDate,
		attachmentBase,
		revisions = []
	}: {
		entry: TimelineEntry;
		last: boolean;
		showDate: boolean;
		attachmentBase: string;
		revisions?: TimelineRevision[];
	} = $props();

	const ICON: Record<EntryType, typeof ActivityIcon> = {
		status: ActivityIcon,
		communication: MegaphoneIcon,
		action: WrenchIcon,
		observation: EyeIcon,
		decision: ScaleIcon
	};

	const ATTACHMENT_ICON: Record<AttachmentKind, typeof LinkIcon> = {
		link: LinkIcon,
		log: ScrollTextIcon,
		image: ImageIcon
	};

	const EVIDENCE: AttachmentKind[] = ['link', 'log', 'image'];

	const label = $derived(ENTRY_TYPES.find((type) => type.id === entry.type)?.label ?? entry.type);
	const Icon = $derived(ICON[entry.type]);

	const local = $derived(
		new Date(entry.at).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
	);

	const attachments = $derived(entry.attachments ?? []);

	let editing = $state(false);
	let attaching = $state(false);
	let showRevisions = $state(false);
	let evidence = $state<AttachmentKind>('link');
	let draft = $state('');
	let draftType = $state<EntryType>('observation');

	function startEdit() {
		draft = entry.text;
		draftType = entry.type;
		editing = true;
	}
</script>

<div class="grid grid-cols-[92px_26px_minmax(0,1fr)] gap-x-2.5">
	<div class="flex flex-col gap-px pt-[3px] text-right">
		{#if showDate}
			<span class="text-subtle-foreground font-mono text-[10px]">{formatUtcDate(entry.at)}</span>
		{/if}
		<span class="text-muted-foreground font-mono text-xs">
			{formatUtcTime(entry.at).replace(' UTC', '')} UTC
		</span>
		<span class="text-subtle-foreground font-mono text-[10px]">{local} local</span>
	</div>

	<div class="flex flex-col items-center">
		<span
			title={label}
			class="bg-popover text-muted-foreground flex size-[22px] shrink-0 items-center justify-center rounded-full border"
		>
			<Icon class="size-[11px]" />
		</span>
		{#if !last}
			<span class="bg-border min-h-2 w-px flex-1"></span>
		{/if}
	</div>

	<div class="group min-w-0 pt-0.5 pb-4">
		<div class="mb-[3px] flex flex-wrap items-center gap-[7px]">
			<span class="text-[12.5px] font-semibold">{entry.actor}</span>
			{#if entry.ai}
				<Badge tone="brand" size="sm">Opsybot</Badge>
			{/if}
			<Badge tone="neutral" size="sm">{label}</Badge>
			{#if entry.alertId}
				<Badge tone="info" size="sm">from alert</Badge>
			{/if}
			{#if entry.result}
				<Badge tone="neutral" size="sm">{entry.result}</Badge>
			{/if}
			{#if entry.retro}
				<Badge tone="info" size="sm">retroactive</Badge>
			{/if}
			{#if entry.edited}
				<form method="POST" action="?/revisions" use:enhance>
					<input type="hidden" name="entryId" value={entry.id} />
					<button
						type="submit"
						class="text-subtle-foreground hover:text-foreground font-mono text-[10px] underline-offset-2 hover:underline"
						onclick={() => (showRevisions = true)}
					>
						edited
					</button>
				</form>
			{/if}
			{#if entry.editable}
				<div class="ml-auto flex items-center gap-1 opacity-0 focus-within:opacity-100 group-hover:opacity-100">
					<Button variant="ghost" size="icon-sm" onclick={startEdit} aria-label="Edit entry">
						<PencilIcon class="size-3.5" />
					</Button>
					<Button
						variant="ghost"
						size="icon-sm"
						onclick={() => (attaching = !attaching)}
						aria-label="Attach evidence"
					>
						<PaperclipIcon class="size-3.5" />
					</Button>
				</div>
			{/if}
		</div>

		{#if editing}
			<form
				method="POST"
				action="?/editEntry"
				class="mb-2 flex flex-col gap-2"
				use:enhance={() =>
					async ({ update }) => {
						editing = false;
						await update({ reset: false });
					}}
			>
				<input type="hidden" name="entryId" value={entry.id} />
				<Textarea name="text" bind:value={draft} rows={3} aria-label="Entry text" />
				<div class="flex flex-wrap items-center gap-1.5">
					{#each ENTRY_TYPES as entryType (entryType.id)}
						<label
							class="inline-flex h-6 cursor-pointer items-center rounded-md border px-2.5 text-xs font-medium has-[:focus-visible]:border-brand-edge has-[:focus-visible]:ring-brand-edge/50 has-[:focus-visible]:ring-2 {draftType ===
							entryType.id
								? 'bg-brand-wash border-brand-edge text-brand-foreground'
								: 'bg-popover border-input text-muted-foreground'}"
						>
							<input
								type="radio"
								name="category"
								value={entryType.id}
								class="sr-only"
								bind:group={draftType}
							/>
							{entryType.label}
						</label>
					{/each}
					<div class="flex-1"></div>
					<Button variant="ghost" size="sm" onclick={() => (editing = false)}>Cancel</Button>
					<Button type="submit" size="sm">Save entry</Button>
				</div>
			</form>
		{:else}
			<div class="text-muted-foreground text-[13.5px] leading-[1.55] [overflow-wrap:anywhere]">
				{entry.text}
			</div>
		{/if}

		{#if attachments.length > 0}
			<div class="mt-2 flex flex-col gap-1.5">
				{#each attachments as attachment (attachment.id)}
					{@const AttachmentIcon = ATTACHMENT_ICON[attachment.kind]}
					<div class="bg-popover rounded-lg border px-2.5 py-2">
						<div class="flex items-center gap-2">
							<AttachmentIcon class="text-muted-foreground size-3 shrink-0" />
							{#if attachment.kind === 'link'}
								<a
									href={attachment.url}
									target="_blank"
									rel="noreferrer"
									class="hover:text-foreground min-w-0 flex-1 truncate text-[12.5px] underline-offset-2 hover:underline"
								>
									{attachment.label}
								</a>
							{:else}
								<span class="min-w-0 flex-1 truncate text-[12.5px]">{attachment.label}</span>
							{/if}
							{#if entry.editable}
								<form method="POST" action="?/detach" use:enhance>
									<input type="hidden" name="attachmentId" value={attachment.id} />
									<Button
										type="submit"
										variant="ghost"
										size="icon-sm"
										aria-label="Remove {attachment.label}"
									>
										<XIcon class="size-3.5" />
									</Button>
								</form>
							{/if}
						</div>
						{#if attachment.kind === 'log' && attachment.body}
							<pre
								class="text-subtle-foreground mt-1.5 max-h-52 overflow-auto rounded-md border p-2 font-mono text-[11px] leading-[1.5] whitespace-pre-wrap">{attachment.body}</pre>
						{/if}
						{#if attachment.kind === 'image'}
							<a href="{attachmentBase}/{attachment.id}" target="_blank" rel="noreferrer">
								<img
									src="{attachmentBase}/{attachment.id}"
									alt={attachment.label}
									class="mt-1.5 max-h-64 rounded-md border"
								/>
							</a>
						{/if}
					</div>
				{/each}
			</div>
		{/if}

		{#if attaching}
			<form
				method="POST"
				action="?/attach"
				enctype="multipart/form-data"
				class="bg-popover mt-2 flex flex-col gap-2 rounded-lg border p-2.5"
				use:enhance={() =>
					async ({ update }) => {
						attaching = false;
						await update({ reset: true });
					}}
			>
				<input type="hidden" name="entryId" value={entry.id} />
				<input type="hidden" name="kind" value={evidence} />
				<div class="flex flex-wrap items-center gap-1.5">
					{#each EVIDENCE as kind (kind)}
						<button
							type="button"
							onclick={() => (evidence = kind)}
							aria-pressed={evidence === kind}
							class="inline-flex h-6 items-center rounded-md border px-2.5 text-xs font-medium capitalize {evidence ===
							kind
								? 'bg-brand-wash border-brand-edge text-brand-foreground'
								: 'border-input text-muted-foreground'}"
						>
							{kind}
						</button>
					{/each}
				</div>
				<Input name="label" placeholder="Label" aria-label="Attachment label" />
				{#if evidence === 'link'}
					<Input
						name="url"
						type="url"
						placeholder="https://grafana.example.com/d/abc"
						aria-label="Link"
					/>
				{:else if evidence === 'log'}
					<Textarea
						name="body"
						rows={4}
						placeholder="Paste the log snippet"
						aria-label="Log snippet"
					/>
				{:else}
					<input
						name="file"
						type="file"
						accept="image/*"
						aria-label="Image file"
						class="text-muted-foreground text-[12.5px]"
					/>
				{/if}
				<div class="flex items-center gap-2">
					<div class="flex-1"></div>
					<Button variant="ghost" size="sm" onclick={() => (attaching = false)}>Cancel</Button>
					<Button type="submit" size="sm">Attach</Button>
				</div>
			</form>
		{/if}

		{#if showRevisions && revisions.length > 0}
			<div class="bg-popover mt-2 rounded-lg border p-2.5">
				<div class="mb-1.5 flex items-center gap-1.5">
					<HistoryIcon class="text-muted-foreground size-3" />
					<span class="text-[12px] font-semibold">Edit history</span>
					<div class="flex-1"></div>
					<Button variant="ghost" size="sm" onclick={() => (showRevisions = false)}>Close</Button>
				</div>
				{#each revisions as revision (revision.id)}
					<div class="border-t py-1.5 first:border-t-0">
						<div class="text-subtle-foreground font-mono text-[10px]">
							{formatUtc(revision.at)} · {revision.editor || 'Opsybot'} · was {revision.type}
						</div>
						<div class="text-muted-foreground text-[12.5px] [overflow-wrap:anywhere]">
							{revision.text}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>
