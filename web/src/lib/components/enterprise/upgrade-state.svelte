<script lang="ts">
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';
	import LockIcon from '@lucide/svelte/icons/lock';
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';

	let { pitch }: { pitch: string } = $props();
</script>

<div
	data-upgrade
	class="mx-auto flex max-w-[640px] flex-col items-center gap-3 rounded-xl border px-5 py-16 text-center"
	style="background: radial-gradient(420px 220px at 50% 0%, rgba(0, 229, 172, 0.06), transparent 70%), var(--ink-3)"
>
	<span class="bg-brand-wash border-brand-edge flex size-[46px] items-center justify-center rounded-full border">
		<LockIcon class="text-brand-foreground size-[18px]" />
	</span>
	<h2 class="m-0 text-[19px] font-light tracking-[-0.02em]">Part of the enterprise license</h2>
	<p class="text-muted-foreground m-0 max-w-[460px] text-[13.5px] leading-[1.65]">{pitch}</p>
	<p class="text-subtle-foreground m-0 text-[12.5px]">
		Everything else in Opsybot stays fully functional without it. €99/month per workspace, flat.
	</p>
	<Button onclick={() => toast('Activating the enterprise license opens the billing upgrade.')}>
		<KeyRoundIcon data-icon="inline-start" />
		Activate license
	</Button>
</div>
