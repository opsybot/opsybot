<script lang="ts">
	import LockIcon from '@lucide/svelte/icons/lock';
	import MegaphoneIcon from '@lucide/svelte/icons/megaphone';
	import SirenIcon from '@lucide/svelte/icons/siren';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import Page from '$lib/components/layout/page.svelte';
	import SpTabs from '$lib/components/statuspages/sp-tabs.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Alert, AlertContent } from '$lib/components/ui/alert';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { Textarea } from '$lib/components/ui/textarea';
	import {
		COMPONENT_STATES,
		NEXT_UPDATE,
		PUBLISH_STAGES,
		TEMPLATES,
		type ComponentState,
		type PublishStage
	} from '$lib/statuspages';
	import { formatUtc } from '$lib/time';
	import { untrack } from 'svelte';
	import { SvelteMap, SvelteSet } from 'svelte/reactivity';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	const live = $derived((data.incident?.published.length ?? 0) > 0);

	const pages = new SvelteSet<string>(untrack(() => (data.incident ? [data.pages[0]?.id].filter(Boolean) : [])));
	const componentStates = new SvelteMap<string, ComponentState>([
		['Checkout', 'degraded'],
		['Payments API', 'degraded']
	]);
	let title = $state('');
	let firstText = $state(TEMPLATES.investigating);
	let nextIn = $state(NEXT_UPDATE[0]);

	let stage = $state<PublishStage>('identified');
	let updateText = $state(TEMPLATES.identified);

	function togglePage(id: string) {
		pages.has(id) ? pages.delete(id) : pages.add(id);
	}
	function toggleComponent(name: string) {
		componentStates.has(name) ? componentStates.delete(name) : componentStates.set(name, 'degraded');
	}

	const stateLabel = (value: ComponentState) =>
		COMPONENT_STATES.find((entry) => entry.value === value)?.label ?? 'Degraded performance';
</script>

<Page title="Status pages" subtitle="Tell customers before they ask">
	<SpTabs current="publish" />

	<div class="mt-3.5 flex max-w-[720px] flex-col gap-3.5">
		{#if !data.incident}
			<div
				class="text-muted-foreground flex flex-col items-center gap-2.5 rounded-xl border border-dashed px-5 py-14"
			>
				<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
					<MegaphoneIcon class="text-subtle-foreground size-5" />
				</span>
				<div class="text-[15px] font-medium">No incident to publish</div>
				<p class="text-subtle-foreground m-0 max-w-[420px] text-center text-[13px] leading-[1.55]">
					This flow is reached from an incident’s status-page tab. Declare an incident, and its
					update goes out to customers from here.
				</p>
				<Button size="sm" variant="secondary" href="/incidents?declare">
					<SirenIcon data-icon="inline-start" />
					Declare incident
				</Button>
			</div>
		{:else}
			<div class="flex flex-wrap items-baseline gap-2.5">
				<h2 class="tracking-heading m-0 text-[18px] font-semibold">
					Publish {data.incident.id} to a status page
				</h2>
				<span class="text-subtle-foreground font-mono text-[11.5px]">
					reached from the incident’s status-page tab
				</span>
			</div>

			{#if !live}
				<div class="bg-card flex flex-col gap-4 rounded-xl border p-4">
					<div>
						<div class="text-subtle-foreground tracking-label mb-2 text-[11px] uppercase">
							Publish to
						</div>
						<div class="flex flex-wrap gap-1.5">
							{#each data.pages as sp (sp.id)}
								<Tag selected={pages.has(sp.id)} onclick={() => togglePage(sp.id)}>{sp.id}</Tag>
							{/each}
						</div>
					</div>

					<div>
						<div class="text-subtle-foreground tracking-label mb-2 text-[11px] uppercase">
							Affected components and their public state
						</div>
						<div class="flex flex-col gap-2">
							{#each data.components as component (component.name)}
								<div class="flex items-center gap-2.5">
									<Checkbox
										checked={componentStates.has(component.name)}
										onCheckedChange={() => toggleComponent(component.name)}
										aria-label={component.name}
									/>
									<span class="text-[13px]">{component.name}</span>
									<div class="flex-1"></div>
									{#if componentStates.has(component.name)}
										<Select.Root
											type="single"
											value={componentStates.get(component.name)}
											onValueChange={(value) =>
												componentStates.set(component.name, value as ComponentState)}
										>
											<Select.Trigger size="sm" class="w-[190px]">
												{stateLabel(componentStates.get(component.name)!)}
											</Select.Trigger>
											<Select.Content>
												<Select.Group>
													{#each COMPONENT_STATES as option (option.value)}
														<Select.Item value={option.value} label={option.label}>
															{option.label}
														</Select.Item>
													{/each}
												</Select.Group>
											</Select.Content>
										</Select.Root>
									{/if}
								</div>
							{/each}
						</div>
					</div>

					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel for="title" class="text-muted-foreground text-[13px] font-medium">
							Public title
						</Field.FieldLabel>
						<Input
							id="title"
							bind:value={title}
							placeholder="Elevated checkout errors in Europe"
						/>
						{#if !title.trim()}
							<div class="mt-1 flex items-center gap-2">
								<span class="text-subtle-foreground text-xs">
									Internal name: “{data.incident.name}”
								</span>
								<button
									type="button"
									onclick={() => (title = data.incident!.name)}
									class="text-brand-foreground text-xs hover:underline"
								>
									Use it anyway
								</button>
							</div>
						{:else if title === data.incident.name}
							<Alert tone="warning" class="mt-1">
								<AlertContent>
									You are publishing the internal name verbatim. Make sure it reads like something a
									customer should see.
								</AlertContent>
							</Alert>
						{:else}
							<Field.FieldDescription class="text-subtle-foreground text-xs">
								What customers see. Symptom-first, no internal jargon.
							</Field.FieldDescription>
						{/if}
					</Field.Field>

					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel for="first" class="text-muted-foreground text-[13px] font-medium">
							First update
						</Field.FieldLabel>
						<Textarea id="first" bind:value={firstText} rows={3} />
						<Field.FieldDescription class="text-subtle-foreground text-xs">
							Impact, no speculation about cause, and a commitment to the next update.
						</Field.FieldDescription>
					</Field.Field>

					<Field.Field class="max-w-[160px] gap-1.5 space-y-0">
						<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
							Next update in
						</Field.FieldLabel>
						<Select.Root type="single" bind:value={nextIn}>
							<Select.Trigger>{nextIn}</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each NEXT_UPDATE as option (option)}
										<Select.Item value={option} label={option}>{option}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</Field.Field>
				</div>

				<form
					method="POST"
					action="?/publish"
					use:enhance={() =>
						async ({ result }) => {
							if (result.type === 'success') {
								toast.success(
									`${[...pages].join(', ')} now shows this incident. Next update promised in ${nextIn}.`
								);
							}
							await invalidateAll();
						}}
				>
					<input type="hidden" name="incident" value={data.incident.id} />
					{#each [...pages] as pageId (pageId)}
						<input type="hidden" name="page" value={pageId} />
					{/each}
					{#each [...componentStates] as [name, state] (name)}
						<input type="hidden" name="component" value={name} />
						<input type="hidden" name="state:{name}" value={state} />
					{/each}
					<input type="hidden" name="title" value={title} />
					<input type="hidden" name="text" value={firstText} />

					<Button
						type="submit"
						disabled={pages.size === 0 || componentStates.size === 0 || !title.trim()}
					>
						<MegaphoneIcon data-icon="inline-start" />
						Publish to {pages.size}
						{pages.size === 1 ? 'page' : 'pages'}
					</Button>
				</form>
			{:else}
				<Alert tone="info">
					<LockIcon />
					<AlertContent>
						Updates are append-only. To correct something, publish a new update — visitors and
						auditors see the full history.
					</AlertContent>
				</Alert>

				<form
					method="POST"
					action="?/update"
					use:enhance={() =>
						async ({ result, update }) => {
							await update({ reset: false });
							if (result.type === 'success') toast.success('Update published. Subscribers notified.');
						}}
					class="bg-card flex flex-col gap-3.5 rounded-xl border p-4"
				>
					<input type="hidden" name="incident" value={data.incident.id} />
					<div class="flex flex-wrap gap-2.5">
						<Field.Field class="w-[180px] gap-1.5 space-y-0">
							<Field.FieldLabel class="text-muted-foreground text-[13px] font-medium">
								Stage
							</Field.FieldLabel>
							<Select.Root
								type="single"
								name="stage"
								value={stage}
								onValueChange={(value) => {
									stage = value as PublishStage;
									updateText = TEMPLATES[stage];
								}}
							>
								<Select.Trigger>{stage}</Select.Trigger>
								<Select.Content>
									<Select.Group>
										{#each PUBLISH_STAGES as option (option)}
											<Select.Item value={option} label={option}>{option}</Select.Item>
										{/each}
									</Select.Group>
								</Select.Content>
							</Select.Root>
						</Field.Field>
					</div>

					<Field.Field class="gap-1.5 space-y-0">
						<Field.FieldLabel for="upd" class="text-muted-foreground text-[13px] font-medium">
							Update text
						</Field.FieldLabel>
						<Textarea id="upd" name="text" bind:value={updateText} rows={3} />
						<Field.FieldDescription class="text-subtle-foreground text-xs">
							Pre-filled with the stage-appropriate template — replace the brackets.
						</Field.FieldDescription>
					</Field.Field>

					<Button type="submit" class="self-start">
						<MegaphoneIcon data-icon="inline-start" />
						Publish update
					</Button>
				</form>

				<section class="bg-card overflow-hidden rounded-xl border">
					<header class="flex flex-wrap items-center gap-2.5 border-b px-4 py-3">
						<span class="text-[13.5px] font-semibold">Published so far</span>
						<Badge tone="neutral" size="sm">{data.incident.published.length}</Badge>
						{#if data.incident.title}
							<span class="text-subtle-foreground text-[12px]">
								· public title “{data.incident.title}”
							</span>
						{/if}
					</header>

					{#each data.incident.published as update (update.at)}
						<div class="flex items-start gap-2.5 border-t px-4 py-3 first:border-t-0">
							<Badge tone={update.tone} size="sm" class="shrink-0">{update.stage}</Badge>
							<div class="min-w-0 flex-1">
								<div class="text-muted-foreground text-[12.5px] leading-[1.55]">{update.text}</div>
								<div class="text-subtle-foreground mt-[3px] font-mono text-[10.5px]">
									{formatUtc(update.at)}
								</div>
							</div>
						</div>
					{/each}
				</section>
			{/if}
		{/if}
	</div>
</Page>
