<script lang="ts">
	import { enhance } from '$app/forms';
	import CopyIcon from '@lucide/svelte/icons/copy';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import { toast } from 'svelte-sonner';
	import AuthShell from '$lib/components/auth/auth-shell.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Field from '$lib/components/ui/field';
	import CodeForm from '../code-form.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let saved = $state(false);

	async function copySecret() {
		await navigator.clipboard.writeText(data.secret);
		toast.success('Setup key copied.');
	}

	async function copyCodes() {
		await navigator.clipboard.writeText(data.recoveryCodes.join('\n'));
		toast.success('Recovery codes copied.');
	}

	function downloadCodes() {
		const blob = new Blob([data.recoveryCodes.join('\n') + '\n'], { type: 'text/plain' });
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.href = url;
		link.download = 'opsybot-recovery-codes.txt';
		link.click();
		URL.revokeObjectURL(url);
	}
</script>

{#if data.step === 'recovery'}
	<AuthShell
		title="Save your recovery codes"
		subtitle="Each code works once if you lose your authenticator. Store them somewhere safe — a password manager works."
	>
		<div class="flex flex-col gap-5">
			<div class="grid grid-cols-2 gap-2" role="list">
				{#each data.recoveryCodes as recoveryCode (recoveryCode)}
					<span
						role="listitem"
						class="bg-inset text-foreground rounded-md border px-3 py-[9px] text-center font-mono text-[13px] tracking-[0.04em]"
					>
						{recoveryCode}
					</span>
				{/each}
			</div>

			<div class="flex gap-2.5">
				<Button variant="secondary" onclick={copyCodes}>
					<CopyIcon data-icon="inline-start" />
					Copy codes
				</Button>
				<Button variant="secondary" onclick={downloadCodes}>
					<DownloadIcon data-icon="inline-start" />
					Download .txt
				</Button>
			</div>

			<Field.Field orientation="horizontal" class="items-start gap-2.5">
				<Checkbox id="saved" bind:checked={saved} class="mt-0.5" />
				<Field.FieldContent class="gap-0.5">
					<Field.FieldLabel for="saved" class="text-foreground text-sm font-normal">
						I saved these codes
					</Field.FieldLabel>
					<Field.FieldDescription class="text-subtle-foreground text-[13px]">
						You won't see them again after this step.
					</Field.FieldDescription>
				</Field.FieldContent>
			</Field.Field>

			<form method="POST" action="?/finish" use:enhance>
				<Button type="submit" class="w-full" disabled={!saved}>Finish setup</Button>
			</form>
		</div>
	</AuthShell>
{:else}
	<AuthShell
		title="Set up two-factor authentication"
		subtitle="Scan the code with any TOTP app — 1Password, Google Authenticator, Authy — then confirm a code to turn it on."
		width={440}
	>
		<div class="flex flex-col gap-5">
			<div class="flex items-start gap-[18px] max-[560px]:flex-col">
				<div class="shrink-0 overflow-hidden rounded-md [&>svg]:block [&>svg]:size-[148px]">
					{@html data.qr}
				</div>

				<div class="flex min-w-0 flex-col gap-2">
					<div class="text-subtle-foreground tracking-label text-[11px] uppercase">
						Can't scan? Enter this key
					</div>
					<code
						class="bg-inset text-foreground rounded-md border px-3 py-2.5 font-mono text-[13px] tracking-[0.05em] [overflow-wrap:anywhere]"
					>
						{data.groupedSecret}
					</code>
					<button
						type="button"
						onclick={copySecret}
						class="text-muted-foreground hover:text-brand-foreground self-start text-[13px]"
					>
						Copy key
					</button>
				</div>
			</div>

			<div class="flex flex-col gap-3">
				<div class="text-subtle-foreground tracking-label text-[11px] uppercase">
					Enter the 6-digit code from your app
				</div>
				<CodeForm data={data.form} action="?/verify" submitLabel="Verify and continue" />
			</div>
		</div>
	</AuthShell>
{/if}
