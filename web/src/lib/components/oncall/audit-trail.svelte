<script lang="ts">
	import HistoryIcon from '@lucide/svelte/icons/history';
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import type { AuditEntry } from '$lib/oncall';
	import { formatUtc } from '$lib/time';

	let { entries }: { entries: AuditEntry[] } = $props();
</script>

<section class="bg-card overflow-hidden rounded-xl border">
	<Collapsible.Root>
		<Collapsible.Trigger
			class="hover:bg-accent text-foreground flex w-full items-center gap-2 px-4 py-3.5 text-[13.5px] font-semibold"
		>
			<HistoryIcon class="size-3.5" />
			Audit trail
			<Badge tone="neutral" size="sm">{entries.length}</Badge>
		</Collapsible.Trigger>

		<Collapsible.Content>
			{#each entries as entry (entry.id)}
				<div class="flex items-center gap-2.5 border-t px-3.5 py-2.5">
					<UserAvatar name={entry.by} size="xs" />
					<span class="min-w-0 flex-1 text-[12.5px]">{entry.what}</span>
					<span class="text-subtle-foreground shrink-0 font-mono text-[10.5px]">
						{formatUtc(entry.at)} · {entry.by.split(' ')[0]}
					</span>
				</div>
			{/each}
		</Collapsible.Content>
	</Collapsible.Root>
</section>
