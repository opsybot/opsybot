<script lang="ts">
	import SparklesIcon from '@lucide/svelte/icons/sparkles';
	import { tick, untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Textarea } from '$lib/components/ui/textarea';
	import type { SectionId } from '$lib/postmortems';

	let {
		section,
		title,
		guidance,
		value,
		readonly = false
	}: {
		section: SectionId;
		title: string;
		guidance?: string;
		value: string;
		readonly?: boolean;
	} = $props();

	let text = $state(untrack(() => value));
	let proposal = $state<string | null>(null);
	let drafting = $state(false);
	let field = $state<HTMLTextAreaElement | null>(null);
	let form = $state<HTMLFormElement | null>(null);

	async function save() {
		await tick();
		form?.requestSubmit();
	}

	function accept(andEdit: boolean) {
		text = proposal ?? '';
		proposal = null;
		if (andEdit) {
			field?.focus();
			toast.info('Draft inserted. Edit away. It saves when you click out.');
		} else {
			save();
			toast.success(`Draft accepted into “${title}”.`);
		}
	}
</script>

<section class="bg-card overflow-hidden rounded-xl border">
	<header class="flex items-center gap-2 border-b px-4 py-3">
		<span class="text-[13.5px] font-semibold">{title}</span>
		<div class="flex-1"></div>

		{#if !readonly && !proposal && !drafting}
			<form
				method="POST"
				action="?/draft"
				use:enhance={() => {
					drafting = true;
					return async ({ result }) => {
						drafting = false;
						if (result.type === 'success' && result.data?.text) {
							proposal = String(result.data.text);
						}
					};
				}}
			>
				<input type="hidden" name="section" value={section} />
				<Button type="submit" size="sm" variant="ghost">
					<SparklesIcon data-icon="inline-start" />
					Draft from timeline
				</Button>
			</form>
		{/if}
	</header>

	{#if guidance}
		<p class="bg-inset text-subtle-foreground m-0 border-b px-4 py-[9px] text-xs leading-[1.55]">
			{guidance}
		</p>
	{/if}

	<div class="p-3.5">
		<form
			method="POST"
			action="?/section"
			bind:this={form}
			use:enhance={() => async ({ update }) => update({ reset: false })}
		>
			<input type="hidden" name="section" value={section} />
			<Textarea
				name="text"
				bind:ref={field}
				bind:value={text}
				rows={3}
				{readonly}
				aria-label={title}
				onblur={save}
			/>
		</form>

		{#if drafting}
			<div class="mt-2.5 flex items-center gap-2.5">
				<span
					class="border-border border-t-primary motion-safe:animate-spin size-3.5 shrink-0 rounded-full border-2"
					aria-hidden="true"
				></span>
				<span class="text-muted-foreground text-[12.5px]">Reading the timeline…</span>
			</div>
		{/if}

		{#if proposal}
			<div class="bg-brand-wash border-brand-edge mt-2.5 rounded-md border p-3">
				<Badge tone="brand" size="sm">
					<SparklesIcon />
					Drafted from the timeline. Review before sharing
				</Badge>

				<p class="text-muted-foreground m-0 mt-[7px] text-[13px] leading-[1.6]">{proposal}</p>

				<div class="mt-2.5 flex gap-2">
					<Button size="sm" onclick={() => accept(false)}>Accept</Button>
					<Button size="sm" variant="secondary" onclick={() => accept(true)}>Accept and edit</Button>
					<Button size="sm" variant="ghost" onclick={() => (proposal = null)}>Discard</Button>
				</div>
			</div>
		{/if}
	</div>
</section>
