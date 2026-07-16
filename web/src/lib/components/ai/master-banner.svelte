<script lang="ts">
	import { tick, untrack } from 'svelte';
	import SparklesIcon from '@lucide/svelte/icons/sparkles';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { Switch } from '$lib/components/ui/switch';

	let { enabled }: { enabled: boolean } = $props();

	let checked = $state(untrack(() => enabled));
	$effect(() => {
		checked = enabled;
	});
	let form: HTMLFormElement;
</script>

<div
	class="flex items-center gap-3 rounded-xl border px-4 py-3.5 transition-colors {checked
		? 'bg-brand-wash border-brand-edge'
		: 'bg-card'}"
>
	<SparklesIcon class="size-4 shrink-0 {checked ? 'text-primary' : 'text-subtle-foreground'}" />
	<div class="min-w-0 flex-1">
		<div class="text-[13.5px] font-semibold">
			{checked ? 'AI features are on' : 'AI features are off'}
		</div>
		<div class="text-subtle-foreground mt-px text-[12px] leading-[1.5]">
			{checked
				? 'Summaries, drafts, and correlation run against the models below. Incident data goes only to these endpoints.'
				: 'Nothing is sent anywhere. Every AI surface shows its off state.'}
		</div>
	</div>
	<form
		method="POST"
		action="?/toggle"
		bind:this={form}
		use:enhance={() => async ({ result, update }) => {
			await update({ reset: false });
			if (result.type !== 'success') {
				checked = enabled;
				toast.error(String((result.type === 'failure' && result.data?.error) || 'Could not change AI.'));
				return;
			}
			if (checked) toast.success('AI features on.');
			else toast('AI features off. No data leaves your infrastructure.');
		}}
	>
		<input type="hidden" name="enabled" value={String(checked)} />
		<Switch
			bind:checked
			aria-label={checked ? 'Turn AI features off' : 'Turn AI features on'}
			onCheckedChange={async () => {
				await tick();
				form.requestSubmit();
			}}
		/>
	</form>
</div>
