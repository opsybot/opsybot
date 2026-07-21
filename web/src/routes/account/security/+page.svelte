<script lang="ts">
	import { untrack } from 'svelte';
	import ArrowRightIcon from '@lucide/svelte/icons/arrow-right';
	import CheckCircle2Icon from '@lucide/svelte/icons/circle-check';
	import CopyIcon from '@lucide/svelte/icons/copy';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';
	import LockIcon from '@lucide/svelte/icons/lock';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import ShieldOffIcon from '@lucide/svelte/icons/shield-off';
	import { toast } from 'svelte-sonner';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Input } from '$lib/components/ui/input';
	import * as InputOTP from '$lib/components/ui/input-otp';
	import { changePasswordSchema, totpSchema } from '$lib/schemas/auth';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let pwOpen = $state(false);
	let setupOpen = $state(false);
	let setupStep = $state<'scan' | 'codes'>('scan');
	let setupCodes = $state<string[]>([]);
	let saved = $state(false);
	let regenOpen = $state(false);
	let regenCodes = $state<string[]>([]);
	let disableOpen = $state(false);

	const pw = superForm(untrack(() => data.passwordForm), {
		validators: zod4Client(changePasswordSchema),
		onUpdated: ({ form }) => {
			if (form.message === 'changed') {
				pwOpen = false;
				toast.success('Password updated', {
					description: 'Use your new password next time you sign in.'
				});
			}
		}
	});
	const { form: pwData, message: pwMsg, errors: pwErrors, enhance: pwEnhance, reset: pwReset } = pw;

	const vf = superForm(untrack(() => data.verifyForm), {
		validators: zod4Client(totpSchema),
		onResult: ({ result }) => {
			if (result.type === 'success' && result.data?.recoveryCodes) {
				setupCodes = result.data.recoveryCodes;
				setupStep = 'codes';
			}
		}
	});
	const { form: vfData, message: vfMsg, enhance: vfEnhance, reset: vfReset } = vf;

	const rf = superForm(untrack(() => data.regenerateForm), {
		validators: zod4Client(totpSchema),
		onResult: ({ result }) => {
			if (result.type === 'success' && result.data?.recoveryCodes) {
				regenCodes = result.data.recoveryCodes;
				toast.success('Recovery codes regenerated', {
					description: 'Your previous codes no longer work. Save the new set.'
				});
			}
		}
	});
	const { form: rfData, message: rfMsg, enhance: rfEnhance, reset: rfReset } = rf;

	const df = superForm(untrack(() => data.disableForm), { validators: zod4Client(totpSchema) });
	const { form: dfData, message: dfMsg, enhance: dfEnhance, reset: dfReset } = df;

	function copy(text: string, msg: string) {
		navigator.clipboard?.writeText(text);
		toast.success(msg);
	}
	function download(codes: string[]) {
		const blob = new Blob([codes.join('\n') + '\n'], { type: 'text/plain' });
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.href = url;
		link.download = 'opsybot-recovery-codes.txt';
		link.click();
		URL.revokeObjectURL(url);
	}

	function openSetup() {
		vfReset();
		setupStep = 'scan';
		setupCodes = [];
		saved = false;
		setupOpen = true;
	}
	function openRegen() {
		rfReset();
		regenCodes = [];
		regenOpen = true;
	}
	function openDisable() {
		dfReset();
		disableOpen = true;
	}
	function openPassword() {
		pwReset();
		pwOpen = true;
	}
</script>

<div style="display: flex; flex-direction: column; gap: 18px">
	<div class="acct-note">
		<ShieldIcon size={15} style="color: var(--text-tertiary); flex-shrink: 0; margin-top: 1px" />
		<span>
			These protect <strong>your account</strong> everywhere it signs in. A workspace can also
			<em>require</em> two-factor: if it does, you'll be prompted at next login.
		</span>
	</div>

	<div class="acct-card">
		<div class="acct-sec">
			<span class="acct-sec-ic"><LockIcon size={18} /></span>
			<div style="flex: 1; min-width: 0">
				<div class="acct-sec-title">Password</div>
				<div class="acct-sec-desc">Used with two-factor to sign in to your account.</div>
			</div>
			<Button size="sm" variant="secondary" onclick={openPassword}>Change password</Button>
		</div>
	</div>

	<div class="acct-card">
		<div class="acct-sec">
			<span class="acct-sec-ic {data.enabled ? 'is-on' : ''}"><ShieldCheckIcon size={18} /></span>
			<div style="flex: 1; min-width: 0">
				<div class="acct-sec-title">
					Two-factor authentication
					<Badge tone={data.enabled ? 'success' : 'neutral'} size="sm">{data.enabled ? 'On' : 'Off'}</Badge>
				</div>
				<div class="acct-sec-desc">
					A time-based code from an authenticator app, required at sign-in. Recommended.
				</div>
			</div>
			{#if !data.enabled && !data.unavailable}
				<Button size="sm" onclick={openSetup}>
					<ShieldCheckIcon class="size-3.5" />
					Set up
				</Button>
			{/if}
		</div>
		{#if data.enabled}
			<div class="acct-sec-body">
				<div class="acct-on">
					<CheckCircle2Icon size={15} style="color: var(--mint-500)" />
					<span>Enabled · authenticator app · recovery codes ready if you lose your device.</span>
				</div>
				<div style="display: flex; gap: 10px">
					<Button size="sm" variant="secondary" onclick={openRegen}>
						<KeyRoundIcon class="size-3.5" />
						Recovery codes
					</Button>
					<Button size="sm" variant="ghost" class="text-critical" onclick={openDisable}>
						<ShieldOffIcon class="size-3.5" />
						Turn off
					</Button>
				</div>
			</div>
		{:else if data.unavailable}
			<div class="acct-sec-body">
				<div class="acct-on">
					<span>
						{data.unavailableDetail ||
							'Two-factor is unavailable. This instance has no auth secret key configured. Ask an admin to set it.'}
					</span>
				</div>
			</div>
		{/if}
	</div>
</div>

<Dialog.Root open={setupOpen} onOpenChange={(v) => (v ? undefined : (setupOpen = false))}>
	<Dialog.Content class="sm:max-w-[460px]">
		<div class="flex flex-col gap-4 p-6">
			<div class="flex items-start gap-3">
				<span class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg">
					{#if setupStep === 'codes'}<KeyRoundIcon class="size-5" />{:else}<ShieldCheckIcon class="size-5" />{/if}
				</span>
				<div class="flex flex-1 flex-col gap-1">
					<Dialog.Title class="tracking-heading text-lg font-semibold">
						{setupStep === 'codes' ? 'Save your recovery codes' : 'Set up two-factor authentication'}
					</Dialog.Title>
					<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
						{setupStep === 'codes'
							? 'Store these somewhere safe. Each one signs you in once if you lose your authenticator.'
							: 'Scan the code with any TOTP app: 1Password, Google Authenticator, Authy, then enter a code to confirm.'}
					</Dialog.Description>
				</div>
			</div>

			{#if setupStep === 'scan'}
				<div style="display: flex; flex-direction: column; gap: 18px; margin-top: 4px">
					<div class="acct-qr-row">
						<div class="acct-qr">{@html data.qr}</div>
						<div style="min-width: 0; flex: 1">
							<div class="acct-detail-label">Can't scan? Enter this key</div>
							<code class="acct-secret">{data.groupedSecret}</code>
							<button type="button" class="acct-linkbtn" onclick={() => copy(data.secret, 'Setup key copied.')}>Copy key</button>
						</div>
					</div>
					<form method="POST" action="?/verify" use:vfEnhance>
						<div class="acct-detail-label">Enter the 6-digit code from your app</div>
						<InputOTP.Root maxlength={6} bind:value={$vfData.code} name="code">
							{#snippet children({ cells })}
								<InputOTP.Group>
									{#each cells.slice(0, 3) as cell, i (i)}<InputOTP.Slot {cell} class="mr-2 last:mr-0" />{/each}
								</InputOTP.Group>
								<InputOTP.Separator />
								<InputOTP.Group>
									{#each cells.slice(3, 6) as cell, i (i)}<InputOTP.Slot {cell} class="mr-2 last:mr-0" />{/each}
								</InputOTP.Group>
							{/snippet}
						</InputOTP.Root>
						{#if $vfMsg === 'wrong'}
							<div class="acct-err">That code didn't match. Codes rotate every 30 s. Enter the current one.</div>
						{/if}
						<div style="display: flex; justify-content: flex-end; gap: 8px; margin-top: 16px">
							<Button type="button" variant="ghost" onclick={() => (setupOpen = false)}>Cancel</Button>
							<Button type="submit">
								<ArrowRightIcon class="size-4" />
								Verify and continue
							</Button>
						</div>
					</form>
				</div>
			{:else}
				<div style="margin-top: 4px">
					<div class="acct-codes" role="list">
						{#each setupCodes as code (code)}<span role="listitem" class="acct-code">{code}</span>{/each}
					</div>
					<div style="display: flex; gap: 10px; margin-top: 14px">
						<Button size="sm" variant="secondary" onclick={() => copy(setupCodes.join('\n'), 'Recovery codes copied.')}>
							<CopyIcon class="size-3.5" />
							Copy codes
						</Button>
						<Button size="sm" variant="secondary" onclick={() => download(setupCodes)}>
							<DownloadIcon class="size-3.5" />
							Download .txt
						</Button>
					</div>
					<div style="margin-top: 16px; display: flex; align-items: flex-start; gap: 10px">
						<Checkbox id="acct-saved" bind:checked={saved} class="mt-0.5" />
						<label for="acct-saved" class="text-foreground text-sm">
							I saved my recovery codes
							<span class="text-subtle-foreground block text-[13px]">You won't see them again after this step.</span>
						</label>
					</div>
					<div style="display: flex; justify-content: flex-end; margin-top: 16px">
						<Button disabled={!saved} onclick={() => (setupOpen = false)}>
							<ShieldCheckIcon class="size-4" />
							Done
						</Button>
					</div>
				</div>
			{/if}
		</div>
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root open={regenOpen} onOpenChange={(v) => (v ? undefined : (regenOpen = false))}>
	<Dialog.Content class="sm:max-w-[460px]">
		<div class="flex flex-col gap-4 p-6">
			<div class="flex items-start gap-3">
				<span class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg">
					<KeyRoundIcon class="size-5" />
				</span>
				<div class="flex flex-1 flex-col gap-1">
					<Dialog.Title class="tracking-heading text-lg font-semibold">Recovery codes</Dialog.Title>
					<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
						{regenCodes.length
							? 'Save this new set. Each code works once; your previous codes no longer work.'
							: 'Enter a current authenticator code to generate a fresh set. The old codes stop working immediately.'}
					</Dialog.Description>
				</div>
			</div>

			{#if regenCodes.length}
				<div style="margin-top: 4px">
					<div class="acct-codes" role="list">
						{#each regenCodes as code (code)}<span role="listitem" class="acct-code">{code}</span>{/each}
					</div>
					<div style="display: flex; gap: 10px; margin-top: 14px">
						<Button size="sm" variant="secondary" onclick={() => copy(regenCodes.join('\n'), 'Recovery codes copied.')}>
							<CopyIcon class="size-3.5" />
							Copy codes
						</Button>
						<Button size="sm" variant="secondary" onclick={() => download(regenCodes)}>
							<DownloadIcon class="size-3.5" />
							Download .txt
						</Button>
					</div>
					<div style="display: flex; justify-content: flex-end; margin-top: 16px">
						<Button onclick={() => (regenOpen = false)}>Done</Button>
					</div>
				</div>
			{:else}
				<form method="POST" action="?/regenerate" use:rfEnhance style="margin-top: 4px">
					<div class="acct-detail-label">Enter the 6-digit code from your app</div>
					<InputOTP.Root maxlength={6} bind:value={$rfData.code} name="code">
						{#snippet children({ cells })}
							<InputOTP.Group>
								{#each cells.slice(0, 3) as cell, i (i)}<InputOTP.Slot {cell} class="mr-2 last:mr-0" />{/each}
							</InputOTP.Group>
							<InputOTP.Separator />
							<InputOTP.Group>
								{#each cells.slice(3, 6) as cell, i (i)}<InputOTP.Slot {cell} class="mr-2 last:mr-0" />{/each}
							</InputOTP.Group>
						{/snippet}
					</InputOTP.Root>
					{#if $rfMsg === 'wrong'}<div class="acct-err">That code didn't match. Enter the current one.</div>{/if}
					<div style="display: flex; justify-content: flex-end; gap: 8px; margin-top: 16px">
						<Button type="button" variant="ghost" onclick={() => (regenOpen = false)}>Cancel</Button>
						<Button type="submit">Regenerate codes</Button>
					</div>
				</form>
			{/if}
		</div>
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root open={disableOpen} onOpenChange={(v) => (v ? undefined : (disableOpen = false))}>
	<Dialog.Content class="sm:max-w-[440px]">
		<div class="flex flex-col gap-4 p-6">
			<div class="flex items-start gap-3">
				<span class="bg-warning-wash text-warning-ink flex size-[38px] shrink-0 items-center justify-center rounded-lg">
					<ShieldOffIcon class="size-5" />
				</span>
				<div class="flex flex-1 flex-col gap-1">
					<Dialog.Title class="tracking-heading text-lg font-semibold">Turn off two-factor?</Dialog.Title>
					<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
						Enter a current authenticator code to confirm. Your account will then be protected by your
						password alone.
					</Dialog.Description>
				</div>
			</div>
			<form method="POST" action="?/disable" use:dfEnhance>
				<div class="acct-detail-label">Enter the 6-digit code from your app</div>
				<InputOTP.Root maxlength={6} bind:value={$dfData.code} name="code">
					{#snippet children({ cells })}
						<InputOTP.Group>
							{#each cells.slice(0, 3) as cell, i (i)}<InputOTP.Slot {cell} class="mr-2 last:mr-0" />{/each}
						</InputOTP.Group>
						<InputOTP.Separator />
						<InputOTP.Group>
							{#each cells.slice(3, 6) as cell, i (i)}<InputOTP.Slot {cell} class="mr-2 last:mr-0" />{/each}
						</InputOTP.Group>
					{/snippet}
				</InputOTP.Root>
				{#if $dfMsg === 'wrong'}<div class="acct-err">That code didn't match. Enter the current one.</div>{/if}
				<div style="display: flex; justify-content: flex-end; gap: 8px; margin-top: 16px">
					<Button type="button" variant="ghost" onclick={() => (disableOpen = false)}>Keep it on</Button>
					<Button type="submit" variant="destructive">Turn it off</Button>
				</div>
			</form>
		</div>
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root open={pwOpen} onOpenChange={(v) => (v ? undefined : (pwOpen = false))}>
	<Dialog.Content class="sm:max-w-[440px]">
		<div class="flex flex-col gap-4 p-6">
			<div class="flex items-start gap-3">
				<span class="bg-brand-wash text-brand-foreground flex size-[38px] shrink-0 items-center justify-center rounded-lg">
					<LockIcon class="size-5" />
				</span>
				<div class="flex flex-1 flex-col gap-1">
					<Dialog.Title class="tracking-heading text-lg font-semibold">Change password</Dialog.Title>
					<Dialog.Description class="text-muted-foreground text-sm leading-[1.55]">
						Enter your current password, then choose a new one. Your other sessions stay signed in.
					</Dialog.Description>
				</div>
			</div>
			<form method="POST" action="?/changePassword" use:pwEnhance style="display: flex; flex-direction: column; gap: 14px">
				{#if $pwMsg === 'wrong'}
					<div class="acct-err">That current password isn't right, or the new one is too weak.</div>
				{/if}
				<div>
					<label class="acct-field-label" for="pw-cur">Current password</label>
					<Input id="pw-cur" name="currentPassword" type="password" bind:value={$pwData.currentPassword} autocomplete="current-password" />
				</div>
				<div>
					<label class="acct-field-label" for="pw-new">New password</label>
					<Input id="pw-new" name="newPassword" type="password" placeholder="12+ characters" bind:value={$pwData.newPassword} autocomplete="new-password" />
					{#if $pwErrors.newPassword}<div class="acct-err">{$pwErrors.newPassword}</div>{/if}
				</div>
				<div>
					<label class="acct-field-label" for="pw-confirm">Confirm new password</label>
					<Input id="pw-confirm" name="confirm" type="password" bind:value={$pwData.confirm} autocomplete="new-password" />
					{#if $pwErrors.confirm}<div class="acct-err">{$pwErrors.confirm}</div>{/if}
				</div>
				<div style="display: flex; justify-content: flex-end; gap: 8px; margin-top: 4px">
					<Button type="button" variant="ghost" onclick={() => (pwOpen = false)}>Cancel</Button>
					<Button type="submit">Update password</Button>
				</div>
			</form>
		</div>
	</Dialog.Content>
</Dialog.Root>
