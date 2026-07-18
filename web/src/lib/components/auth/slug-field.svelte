<script lang="ts" generics="T extends Record<string, unknown>, U extends FormPath<T>">
	import { untrack } from 'svelte';
	import CheckIcon from '@lucide/svelte/icons/check';
	import LoaderIcon from '@lucide/svelte/icons/loader-circle';
	import type { FormPath, SuperForm } from 'sveltekit-superforms';
	import TextField from '$lib/components/text-field.svelte';
	import { SLUG_RE, slugify } from '$lib/slug';

	let {
		form,
		name,
		workspace
	}: {
		form: SuperForm<T>;
		name: U;
		workspace: string;
	} = $props();

	const { form: formData } = untrack(() => form);

	let touched = $state(false);
	let status = $state<'idle' | 'checking' | 'available' | 'taken' | 'invalid'>('idle');
	let suggestion = $state('');
	let timer: ReturnType<typeof setTimeout> | undefined;

	function current(): string {
		return ($formData[name] as string) ?? '';
	}

	$effect(() => {
		const wsName = workspace;
		if (touched) return;
		untrack(() => {
			const next = wsName.trim() ? slugify(wsName) : '';
			if (current() !== next) {
				($formData[name] as unknown) = next;
				schedule(next);
			}
		});
	});

	function schedule(slug: string) {
		clearTimeout(timer);
		suggestion = '';
		if (!slug) {
			status = 'idle';
			return;
		}
		if (!SLUG_RE.test(slug)) {
			status = 'invalid';
			return;
		}
		status = 'checking';
		timer = setTimeout(() => check(slug), 400);
	}

	async function check(slug: string) {
		try {
			const res = await fetch(`/slug-available?slug=${encodeURIComponent(slug)}`);
			const data = (await res.json()) as { checked: boolean; available?: boolean; suggestion?: string };
			if (current() !== slug) return;
			if (!data.checked) {
				status = 'idle';
			} else if (data.available) {
				status = 'available';
				suggestion = '';
			} else {
				status = 'taken';
				suggestion = data.suggestion ?? '';
			}
		} catch {
			status = 'idle';
		}
	}

	function onInput(event: Event) {
		touched = true;
		schedule((event.currentTarget as HTMLInputElement).value);
	}

	function useSuggestion() {
		if (!suggestion) return;
		touched = true;
		($formData[name] as unknown) = suggestion;
		status = 'available';
		suggestion = '';
	}
</script>

<div class="flex flex-col gap-1.5">
	<TextField
		{form}
		{name}
		label="Workspace URL"
		placeholder="acme"
		autocapitalize="none"
		autocorrect="off"
		spellcheck={false}
		oninput={onInput}
	/>

	{#if status === 'checking'}
		<p class="text-subtle-foreground flex items-center gap-1.5 text-xs">
			<LoaderIcon class="size-3 animate-spin" aria-hidden="true" />
			Checking availability…
		</p>
	{:else if status === 'available'}
		<p class="text-success-ink flex items-center gap-1.5 text-xs">
			<CheckIcon class="size-3" aria-hidden="true" />
			<span><span class="font-mono">{current()}</span> is available.</span>
		</p>
	{:else if status === 'taken'}
		<p class="text-critical-ink text-xs">
			<span class="font-mono">{current()}</span> is taken.
			{#if suggestion}
				<button
					type="button"
					onclick={useSuggestion}
					class="text-brand-foreground font-medium hover:underline"
				>
					Try {suggestion}
				</button>
			{/if}
		</p>
	{:else if status === 'invalid'}
		<p class="text-critical-ink text-xs">Lowercase letters, numbers, and hyphens; start with a letter.</p>
	{/if}
</div>
