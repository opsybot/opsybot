<script lang="ts">
	import ArrowUpRightIcon from '@lucide/svelte/icons/arrow-up-right';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import WebhookIcon from '@lucide/svelte/icons/webhook';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import Tag from '$lib/components/tag.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let hookName = $state('');
	let hookUrl = $state('');
	let hookSecret = $state('');
</script>

<div class="flex flex-col gap-3.5">
	<div class="flex items-center">
		<span class="text-subtle-foreground text-[13px]">
			{data.policies.length}
			{data.policies.length === 1 ? 'policy' : 'policies'}
		</span>
		<div class="flex-1"></div>
		<Button size="sm" href={ws('/escalation-policies/new')}>
			<PlusIcon data-icon="inline-start" />
			New policy
		</Button>
	</div>

	{#if data.policies.length === 0}
		<div
			class="text-muted-foreground flex flex-col items-center gap-2.5 rounded-xl border border-dashed px-5 py-14"
		>
			<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
				<ArrowUpRightIcon class="text-subtle-foreground size-5" />
			</span>
			<div class="text-[15px] font-medium">No escalation policies</div>
			<p class="text-subtle-foreground m-0 max-w-[420px] text-center text-[13px] leading-[1.55]">
				A policy decides who gets paged, in what order, and how long to wait between steps. Alerts
				route to a policy; the policy does the chasing.
			</p>
			<Button size="sm" variant="secondary" href={ws('/escalation-policies/new')}>
				<PlusIcon data-icon="inline-start" />
				Create your first policy
			</Button>
		</div>
	{:else}
		<div class="bg-card overflow-hidden rounded-xl border">
			{#each data.policies as policy (policy.id)}
				<a
					href={ws(`/escalation-policies/${policy.id}`)}
					data-policy={policy.id}
					class="hover:bg-accent flex items-center gap-[18px] border-t px-4 py-3.5 first:border-t-0"
				>
					<div class="min-w-0 flex-1">
						<div class="flex flex-wrap items-center gap-2">
							<span class="font-mono text-[13.5px] font-medium">{policy.name}</span>
							<Tag>{policy.team}</Tag>
							{#if policy.branch}
								<Badge tone="info" size="sm" dot>branches by {policy.branch}</Badge>
							{/if}
							{#if policy.warning}
								<Badge tone="warning" size="sm" dot>{policy.warning}</Badge>
							{/if}
						</div>
						<div class="text-muted-foreground mt-1.5 flex items-center gap-1.5 overflow-hidden">
							{#each policy.summary as part, index (index)}
								{#if index > 0}
									<ChevronRightIcon class="text-subtle-foreground size-[11px] shrink-0" />
								{/if}
								<span
									class="text-[11.5px] whitespace-nowrap {part.kind === 'wait'
										? 'font-mono'
										: ''} {part.kind === 'branch' ? 'text-brand-foreground' : ''}"
								>
									{part.text}
								</span>
							{/each}
						</div>
					</div>
					<div class="shrink-0 text-right">
						<div class="font-mono text-[15px] font-semibold">{policy.routed}</div>
						<div class="text-subtle-foreground text-[10.5px]">alerts routed</div>
					</div>
					<ChevronRightIcon class="text-subtle-foreground size-4 shrink-0" />
				</a>
			{/each}
		</div>
	{/if}
</div>

<div class="bg-card mt-3.5 overflow-hidden rounded-xl border">
	<header class="flex items-center gap-2 border-b px-4 py-3">
		<WebhookIcon class="text-subtle-foreground size-3.5" />
		<span class="text-[13.5px] font-semibold">Webhook targets</span>
		<span class="text-subtle-foreground text-[12px]">Levels can page these instead of a person.</span>
	</header>
	<div class="flex flex-col gap-3 p-[14px]">
		{#each data.webhooks as hook (hook.slug)}
			<div class="bg-inset flex items-center gap-3 rounded-lg border px-3 py-2">
				<span class="font-mono text-[12.5px] font-medium">{hook.slug}</span>
				<span class="text-subtle-foreground min-w-0 flex-1 truncate font-mono text-[11.5px]">{hook.url}</span>
				{#if hook.hasSecret}
					<Badge tone="neutral" size="sm">signed</Badge>
				{/if}
				<form
					method="POST"
					action="?/deleteWebhook"
					use:enhance={() => async ({ result, update }) => {
						await update({ invalidateAll: true });
						if (result.type === 'failure') toast.error(String(result.data?.error ?? 'Could not delete the webhook.'));
					}}
				>
					<input type="hidden" name="slug" value={hook.slug} />
					<Button type="submit" size="icon-sm" variant="ghost" aria-label="Delete webhook {hook.slug}">
						<Trash2Icon />
					</Button>
				</form>
			</div>
		{:else}
			<p class="text-subtle-foreground m-0 text-[12.5px]">
				No webhook targets yet. Add one and pick it inside a policy level.
			</p>
		{/each}

		<form
			method="POST"
			action="?/addWebhook"
			class="flex flex-wrap items-end gap-2"
			use:enhance={() => async ({ result, update }) => {
				await update({ invalidateAll: true });
				if (result.type === 'failure') {
					toast.error(String(result.data?.error ?? 'Could not create the webhook.'));
					return;
				}
				if (result.type === 'success') {
					hookName = '';
					hookUrl = '';
					hookSecret = '';
					toast.success('Webhook target added.');
				}
			}}
		>
			<Input name="name" bind:value={hookName} placeholder="ops-bridge" class="w-[160px] font-mono" aria-label="Webhook name" />
			<Input name="url" bind:value={hookUrl} placeholder="https://bridge.example.com/hook" class="min-w-[240px] flex-1 font-mono" aria-label="Webhook URL" />
			<Input name="secret" bind:value={hookSecret} placeholder="signing secret (optional)" class="w-[210px] font-mono" aria-label="Webhook signing secret" />
			<Button type="submit" size="sm" variant="secondary" disabled={!hookName.trim() || !hookUrl.trim()}>
				<PlusIcon data-icon="inline-start" />
				Add webhook
			</Button>
		</form>
	</div>
</div>
