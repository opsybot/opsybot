<script lang="ts">
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { toast } from 'svelte-sonner';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const BACKUP_CMD = `docker compose exec app \\
  opsybot backup create --dest s3://acme-opsybot-backups \\
  --retention 30d`;
	const RESTORE_CMD = `opsybot backup restore \\
  --from s3://acme-opsybot-backups/2026-07-12.tar.zst \\
  --verify`;
</script>

<div class="flex max-w-[760px] flex-col gap-3.5">
	<div class="bg-card flex items-center gap-3 rounded-xl border p-4">
		<span class="bg-inset flex size-[34px] shrink-0 items-center justify-center rounded-full border">
			{#if data.backup}
				<CircleCheckIcon class="size-4 text-[var(--success)]" />
			{:else}
				<DownloadIcon class="text-subtle-foreground size-4" />
			{/if}
		</span>
		<div class="flex-1">
			{#if data.backup}
				<div class="text-[13.5px] font-semibold">Last backup {data.backup.ago}</div>
				<div class="text-subtle-foreground mt-px font-mono text-[11px]">
					{data.backup.at} · {data.backup.size} · {data.backup.dest} · {data.backup.schedule}
				</div>
			{:else}
				<div class="text-[13.5px] font-semibold">No backups yet</div>
				<div class="text-subtle-foreground mt-px text-[11.5px]">Run the first one, then point nightly backups at a bucket below.</div>
			{/if}
		</div>
		<Button
			size="sm"
			variant="secondary"
			onclick={() => toast.success("On-demand backup started — you'll get a notification when it completes.")}
		>
			<DownloadIcon data-icon="inline-start" />
			Back up now
		</Button>
	</div>

	<div class="bg-card overflow-hidden rounded-xl border">
		<header class="flex items-center gap-2 border-b px-4 py-[11px]">
			<span class="text-[13.5px] font-semibold">Backup</span>
		</header>
		<div class="p-4">
			<p class="text-muted-foreground m-0 mb-2.5 text-[13px] leading-[1.6]">
				Nightly backups run from the app container. Point them at any S3-compatible bucket:
			</p>
			<pre class="text-muted-foreground m-0 overflow-x-auto rounded-md border bg-[var(--ink-0)] px-3.5 py-3 font-mono text-[12px] leading-[1.7] whitespace-pre">{BACKUP_CMD}</pre>
			<div class="mt-3 flex flex-wrap gap-2.5">
				<span class="text-subtle-foreground font-mono text-[11.5px]">Retention is set in</span>
				<a href="/workspace/settings" class="text-brand-foreground text-[12px] hover:underline">Workspace admin → data retention</a>
			</div>
		</div>
	</div>

	<div class="bg-card overflow-hidden rounded-xl border">
		<Collapsible.Root>
			<Collapsible.Trigger
				class="group hover:bg-accent flex w-full items-center gap-2 px-4 py-3 text-[13px] font-semibold transition-colors"
			>
				<RotateCcwIcon class="size-[13px]" />
				Restore procedure
				<ChevronDownIcon class="text-subtle-foreground ml-auto size-4 transition-transform duration-200 group-data-[state=open]:rotate-180" />
			</Collapsible.Trigger>
			<Collapsible.Content>
				<div class="px-4 pb-4">
					<p class="text-muted-foreground m-0 mb-2.5 text-[13px] leading-[1.6]">
						Restore into a fresh instance, then cut traffic over. Never restore over a running production database.
					</p>
					<pre class="text-muted-foreground m-0 overflow-x-auto rounded-md border bg-[var(--ink-0)] px-3.5 py-3 font-mono text-[12px] leading-[1.7] whitespace-pre">{RESTORE_CMD}</pre>
					<Alert.Root tone="warning" class="mt-3">
						<TriangleAlertIcon />
						<Alert.Content>
							<Alert.Description>
								After restore, rotate the SCIM and API tokens — backups include them. Paging stays paused until the
								scheduler confirms healthy.
							</Alert.Description>
						</Alert.Content>
					</Alert.Root>
				</div>
			</Collapsible.Content>
		</Collapsible.Root>
	</div>
</div>
