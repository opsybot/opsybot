<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import ArrowUpRightIcon from '@lucide/svelte/icons/arrow-up-right';
	import CheckIcon from '@lucide/svelte/icons/check';
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import ChevronUpIcon from '@lucide/svelte/icons/chevron-up';
	import CopyIcon from '@lucide/svelte/icons/copy';
	import EyeOffIcon from '@lucide/svelte/icons/eye-off';
	import ImageIcon from '@lucide/svelte/icons/image';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import RotateCwIcon from '@lucide/svelte/icons/rotate-cw';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import Page from '$lib/components/layout/page.svelte';
	import ConfirmDialog from '$lib/components/statuspages/confirm-dialog.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as RadioGroup from '$lib/components/ui/radio-group';
	import { Switch } from '$lib/components/ui/switch';
	import { Textarea } from '$lib/components/ui/textarea';
	import { ws } from '$lib/navigation';
	import { ACCENTS, VISIBILITIES } from '$lib/statuspages';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	const page = $derived(data.page);

	// Seed once from saved values so edits are not reset by re-renders
	let accent = $state(untrack(() => page.accent));
	let visibility = $state(untrack(() => page.visibility));
	let utcDefault = $state(untrack(() => page.utcDefault));
	let showUptime = $state(untrack(() => page.showUptime));
	let allowIndexing = $state(untrack(() => page.allowIndexing));

	let unpublishing = $state(false);
	let deleting = $state(false);

	$effect(() => {
		if (form?.error) toast.error(form.error);
	});
	$effect(() => {
		if (form?.saved) toast.success('Settings saved.');
	});

	const DISPLAY = [
		{
			key: 'utcDefault',
			label: 'Times in UTC by default',
			hint: 'Viewers can toggle to their local time.',
			get: () => utcDefault,
			set: (value: boolean) => (utcDefault = value)
		},
		{
			key: 'showUptime',
			label: 'Show uptime',
			hint: '90-day uptime bars per component.',
			get: () => showUptime,
			set: (value: boolean) => (showUptime = value)
		},
		{
			key: 'allowIndexing',
			label: 'Allow search-engine indexing',
			hint: 'Off keeps the page out of search results.',
			get: () => allowIndexing,
			set: (value: boolean) => (allowIndexing = value)
		}
	];
</script>

<Page title="Status pages" subtitle="Tell customers before they ask">
	<div class="flex max-w-[720px] flex-col gap-3.5">
		<a
			href={ws('/status-pages')}
			class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[12.5px]"
		>
			<ArrowLeftIcon class="size-3.5" />
			Status pages
		</a>

		<div class="flex flex-wrap items-center gap-2.5">
			<h2 class="m-0 font-mono text-[18px] font-semibold">{page.domain}</h2>
			<Badge tone={page.visibility === 'public' ? 'info' : 'neutral'} size="sm">
				{page.visibility}
			</Badge>
			{#if !page.published}
				<Badge tone="neutral" size="sm">unpublished</Badge>
			{/if}
			<div class="flex-1"></div>
			{#if page.published}
				<Button
					size="sm"
					variant="secondary"
					href="https://{page.domain}"
					target="_blank"
					rel="noopener"
				>
					<ArrowUpRightIcon data-icon="inline-start" />
					View page
				</Button>
			{:else}
				<form method="POST" action="?/publish" use:enhance>
					<Button type="submit" size="sm">Republish</Button>
				</form>
			{/if}
		</div>

		<form
			method="POST"
			action="?/save"
			use:enhance={() => async ({ update }) => update({ reset: false })}
			class="flex flex-col gap-3.5"
		>
			<input type="hidden" name="accent" value={accent} />
			<input type="hidden" name="visibility" value={visibility} />
			{#each DISPLAY as toggle (toggle.key)}
				{#if toggle.get()}<input type="hidden" name={toggle.key} value="on" />{/if}
			{/each}

			<section class="bg-card overflow-hidden rounded-xl border">
				<header class="border-b px-4 py-3">
					<span class="text-[13.5px] font-semibold">General</span>
				</header>
				<div class="flex flex-col gap-3.5 p-4">
					<div class="flex flex-wrap gap-2.5">
						<Field.Field class="min-w-[200px] flex-1 gap-1.5 space-y-0">
							<Field.FieldLabel for="name" class="text-muted-foreground text-[13px] font-medium">
								Name
							</Field.FieldLabel>
							<Input id="name" name="name" value={page.name} />
						</Field.Field>
						<Field.Field class="min-w-[200px] flex-1 gap-1.5 space-y-0">
							<Field.FieldLabel for="pageTitle" class="text-muted-foreground text-[13px] font-medium">
								Page title
							</Field.FieldLabel>
							<Input id="pageTitle" name="pageTitle" value={page.pageTitle} />
							<Field.FieldDescription class="text-subtle-foreground text-xs">
								Browser tab and search results.
							</Field.FieldDescription>
						</Field.Field>
					</div>

					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel for="description" class="text-muted-foreground text-[13px] font-medium">
							Description
						</Field.FieldLabel>
						<Textarea id="description" name="description" value={page.description} rows={2} />
					</Field.Field>

					<div class="flex flex-wrap items-start gap-5">
						<div>
							<div class="text-subtle-foreground tracking-label mb-2 text-[11px] uppercase">Logo</div>
							<button
								type="button"
								onclick={() => toast.info('Logo upload is a later part of this build.')}
								class="bg-inset border-border-strong text-subtle-foreground hover:text-muted-foreground hover:border-brand-edge inline-flex items-center gap-2 rounded-md border border-dashed px-4 py-3 text-[12.5px]"
							>
								<ImageIcon class="size-[15px]" />
								Upload — SVG or PNG
							</button>
						</div>
						<div>
							<div class="text-subtle-foreground tracking-label mb-2 text-[11px] uppercase">
								Accent color
							</div>
							<div class="flex gap-2">
								{#each ACCENTS as swatch (swatch.id)}
									<button
										type="button"
										aria-label={swatch.id}
										aria-pressed={accent === swatch.id}
										onclick={() => (accent = swatch.id)}
										class="size-[26px] rounded-full border-2 {accent === swatch.id
											? 'border-foreground'
											: 'border-transparent'}"
										style="background: {swatch.color}{accent === swatch.id
											? '; box-shadow: var(--focus-ring)'
											: ''}"
									></button>
								{/each}
							</div>
							<div class="text-subtle-foreground mt-1.5 text-[11.5px]">
								Constrained set — status colors stay reserved for status.
							</div>
						</div>
					</div>
				</div>
			</section>

			<section class="bg-card overflow-hidden rounded-xl border">
				<header class="border-b px-4 py-3">
					<span class="text-[13.5px] font-semibold">Visibility</span>
				</header>
				<div class="flex flex-col gap-3.5 p-4">
					<RadioGroup.Root bind:value={visibility} class="gap-2.5">
						{#each VISIBILITIES as option (option.value)}
							<div class="flex items-start gap-2.5">
								<RadioGroup.Item value={option.value} id="vis-{option.value}" class="mt-0.5" />
								<Label for="vis-{option.value}" class="flex flex-col items-start gap-0.5 font-normal">
									<span class="text-foreground text-sm leading-[1.3]">{option.label}</span>
									<span class="text-subtle-foreground text-[13px] leading-[1.4]">
										{option.description}
									</span>
								</Label>
							</div>
						{/each}
					</RadioGroup.Root>

					{#if visibility === 'token'}
						<div class="flex flex-wrap items-center gap-2">
							<code
								class="bg-inset text-foreground min-w-0 flex-1 rounded-md border px-2.5 py-2 font-mono text-xs [overflow-wrap:anywhere]"
							>
								https://{page.domain}/?t={form?.token ?? page.token}
							</code>
							<Button
								type="button"
								size="sm"
								variant="secondary"
								onclick={() => {
									navigator.clipboard.writeText(`https://${page.domain}/?t=${form?.token ?? page.token}`);
									toast.success('Tokenized link copied.');
								}}
							>
								<CopyIcon data-icon="inline-start" />
								Copy
							</Button>
							<Button type="submit" formaction="?/rotateToken" size="sm" variant="ghost">
								<RotateCwIcon data-icon="inline-start" />
								Rotate
							</Button>
						</div>
					{/if}
				</div>
			</section>

			<section class="bg-card overflow-hidden rounded-xl border">
				<header class="border-b px-4 py-3">
					<span class="text-[13.5px] font-semibold">Custom domain</span>
				</header>
				<div class="flex flex-col gap-2.5 p-4">
					<div class="flex flex-wrap items-start gap-3">
						<Field.Field class="min-w-[220px] flex-1 gap-1.5 space-y-0">
							<Field.FieldLabel for="domain" class="text-muted-foreground text-[13px] font-medium">
								Domain
							</Field.FieldLabel>
							<Input id="domain" name="domain" value={page.domain} class="h-[34px] font-mono text-[12.5px]" />
						</Field.Field>
						<div class="flex flex-col gap-1.5 pt-6">
							{#if page.domainVerified}
								<Badge tone="success" size="sm" dot>CNAME verified</Badge>
								<Badge tone="success" size="sm" dot>certificate issued · renews {page.certRenews}</Badge>
							{:else}
								<Badge tone="warning" size="sm" dot>CNAME not verified yet</Badge>
								<Badge tone="neutral" size="sm">certificate issues once the CNAME resolves</Badge>
							{/if}
						</div>
					</div>
					<p class="text-subtle-foreground m-0 text-[11.5px]">
						Point a CNAME at pages.opsy.bot — verification and certificates are automatic.
					</p>
				</div>
			</section>

			<section class="bg-card overflow-hidden rounded-xl border">
				<header class="border-b px-4 py-3">
					<span class="text-[13.5px] font-semibold">Display</span>
				</header>
				<div class="flex flex-col gap-3.5 p-4">
					{#each DISPLAY as toggle (toggle.key)}
						<div class="flex items-center gap-3">
							<div class="flex-1">
								<div class="text-[13px] font-medium">{toggle.label}</div>
								<div class="text-subtle-foreground mt-px text-[11.5px]">{toggle.hint}</div>
							</div>
							<Switch
								checked={toggle.get()}
								aria-label={toggle.label}
								onCheckedChange={(next) => toggle.set(next)}
							/>
						</div>
					{/each}
				</div>
			</section>

			<Button type="submit" class="self-start">
				<CheckIcon data-icon="inline-start" />
				Save settings
			</Button>
		</form>

		<!-- Outside the settings form, forms cannot nest -->
		<section class="bg-card overflow-hidden rounded-xl border">
			<header class="flex flex-wrap items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Components</span>
				<span class="text-subtle-foreground text-[11.5px]">
					what visitors see, mapped from services
				</span>
				<div class="flex-1"></div>
				<Button
					size="sm"
					variant="ghost"
					onclick={() =>
						toast.info('A component maps one or more catalog services to a public name.')}
				>
					<PlusIcon data-icon="inline-start" />
					Add component
				</Button>
			</header>

			{#each page.components as component, index (component.id)}
				<div class="flex items-center gap-3 border-t px-4 py-2.5">
					<span class="text-subtle-foreground w-[60px] shrink-0 font-mono text-[10.5px]">
						{component.group}
					</span>
					<span class="w-[120px] shrink-0 text-[13px] font-medium">{component.name}</span>
					<div class="flex flex-1 flex-wrap gap-[5px]">
						{#each component.services as service (service)}
							<Tag href={ws(`/catalog/${service}`)}>{service}</Tag>
						{:else}
							<span class="text-subtle-foreground text-[11.5px]">manual only</span>
						{/each}
					</div>
					<Badge tone={component.stateTone} size="sm">{component.stateLabel}</Badge>
					<form method="POST" action="?/moveComponent" use:enhance>
						<input type="hidden" name="id" value={component.id} />
						<input type="hidden" name="by" value="up" />
						<Button
							type="submit"
							variant="ghost"
							size="icon-sm"
							aria-label="Move {component.name} up"
							disabled={index === 0}
						>
							<ChevronUpIcon />
						</Button>
					</form>
					<form method="POST" action="?/moveComponent" use:enhance>
						<input type="hidden" name="id" value={component.id} />
						<input type="hidden" name="by" value="down" />
						<Button
							type="submit"
							variant="ghost"
							size="icon-sm"
							aria-label="Move {component.name} down"
							disabled={index === page.components.length - 1}
						>
							<ChevronDownIcon />
						</Button>
					</form>
				</div>
			{/each}
		</section>

		<section class="border-critical-edge bg-card overflow-hidden rounded-xl border">
			<header class="border-b px-4 py-3">
				<span class="text-critical-ink text-[13.5px] font-semibold">Danger zone</span>
			</header>
			<div class="flex flex-wrap gap-2.5 p-4">
				<Button size="sm" variant="secondary" onclick={() => (unpublishing = true)}>
					<EyeOffIcon data-icon="inline-start" />
					Unpublish page
				</Button>
				<Button size="sm" variant="destructive" onclick={() => (deleting = true)}>
					<Trash2Icon data-icon="inline-start" />
					Delete page
				</Button>
			</div>
		</section>
	</div>
</Page>

<ConfirmDialog
	bind:open={unpublishing}
	tone="warning"
	title="Unpublish {page.domain}?"
	action="?/unpublish"
	confirmLabel="Unpublish"
>
	The page returns 404 to visitors. Subscribers stay; nothing is deleted. Republish any time.
</ConfirmDialog>

<ConfirmDialog
	bind:open={deleting}
	tone="critical"
	title="Delete {page.domain}?"
	action="?/remove"
	confirmLabel="Delete permanently"
>
	Deletes the page, its update history, and its subscribers. This cannot be undone.
</ConfirmDialog>
