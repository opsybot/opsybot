<script lang="ts">
	import CopyIcon from '@lucide/svelte/icons/copy';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import EyeOffIcon from '@lucide/svelte/icons/eye-off';
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';

	let {
		label,
		value,
		secret = false
	}: { label: string; value: string; secret?: boolean } = $props();

	let revealed = $state(false);
	const shown = $derived(secret && !revealed ? value.slice(0, 6) + '••••••••••••' : value);

	async function copy() {
		try {
			await navigator.clipboard?.writeText(value);
			toast.success(`${label} copied.`);
		} catch {
			toast.error(`Could not copy. Select and copy the ${label.toLowerCase()} manually.`);
		}
	}
</script>

<div>
	<div class="text-subtle-foreground mb-[7px] text-[11px] tracking-[0.08em] uppercase">{label}</div>
	<div class="flex items-center gap-2">
		<code
			class="bg-inset text-foreground flex-1 rounded-md border px-[11px] py-[9px] font-mono text-[12px] [overflow-wrap:anywhere]"
		>
			{shown}
		</code>
		{#if secret}
			<Button variant="ghost" size="sm" onclick={() => (revealed = !revealed)}>
				{#if revealed}
					<EyeOffIcon data-icon="inline-start" />
					Hide
				{:else}
					<EyeIcon data-icon="inline-start" />
					Reveal
				{/if}
			</Button>
		{/if}
		<Button variant="secondary" size="sm" onclick={copy}>
			<CopyIcon data-icon="inline-start" />
			Copy
		</Button>
	</div>
</div>
