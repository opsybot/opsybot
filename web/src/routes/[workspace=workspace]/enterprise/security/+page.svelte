<script lang="ts">
	import { untrack } from 'svelte';
	import CheckIcon from '@lucide/svelte/icons/check';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import * as Alert from '$lib/components/ui/alert';
	import { Button } from '$lib/components/ui/button';
	import * as Select from '$lib/components/ui/select';
	import { Switch } from '$lib/components/ui/switch';
	import { Textarea } from '$lib/components/ui/textarea';
	import UpgradeState from '$lib/components/enterprise/upgrade-state.svelte';
	import { ENT_PITCH, SESSION_OPTIONS } from '$lib/enterprise';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let sessionLifetime = $state(untrack(() => data.security?.sessionLifetime ?? '14 days'));
	let ipAllowlist = $state(untrack(() => data.security?.ipAllowlist ?? ''));
	let enforce2fa = $state(untrack(() => data.security?.enforce2fa ?? false));
	let ssoEnforced = $state(untrack(() => data.security?.ssoEnforced ?? false));

	function changeSession(value: string) {
		sessionLifetime = value;
		toast(`Session lifetime set to ${value}. Applies at next login.`);
	}
	function toggle2fa(value: boolean) {
		enforce2fa = value;
		if (value) toast.success('2FA enforced. 2 members will be prompted.');
		else toast.warning('2FA enforcement off.');
	}
	function toggleSso(value: boolean) {
		ssoEnforced = value;
		if (value) toast.warning('SSO enforced. Password login is now disabled.');
		else toast('SSO enforcement off: password login allowed again.');
	}
</script>

{#if !data.licensed}
	<UpgradeState pitch={ENT_PITCH.security} />
{:else}
	<div class="flex max-w-[680px] flex-col gap-3.5">
		<div class="bg-card flex flex-col gap-4 rounded-xl border p-4">
			<div class="flex max-w-[240px] flex-col gap-1.5">
				<span class="text-muted-foreground text-[13px] font-medium">Session lifetime</span>
				<Select.Root type="single" value={sessionLifetime} onValueChange={changeSession}>
					<Select.Trigger size="sm" aria-label="Session lifetime">{sessionLifetime}</Select.Trigger>
					<Select.Content>
						<Select.Group>
							{#each SESSION_OPTIONS as option (option)}
								<Select.Item value={option} label={option}>{option}</Select.Item>
							{/each}
						</Select.Group>
					</Select.Content>
				</Select.Root>
				<span class="text-subtle-foreground text-[11.5px]">Idle sessions end sooner: 24 h of inactivity.</span>
			</div>

			<div>
				<div class="text-subtle-foreground mb-[7px] text-[11px] tracking-[0.08em] uppercase">
					IP allowlist for admin actions
				</div>
				<Textarea rows={3} class="font-mono text-[12px]" aria-label="IP allowlist for admin actions" bind:value={ipAllowlist} />
				<span class="text-subtle-foreground mt-[7px] block text-[11.5px] leading-[1.5]">
					CIDR ranges, one per line. Admin actions from other IPs are blocked and logged; normal responding is
					unaffected.
				</span>
			</div>

			<div class="flex items-center gap-3">
				<div class="flex-1">
					<div class="text-[13px] font-medium">Enforce two-factor authentication</div>
					<div class="text-subtle-foreground mt-px text-[11.5px]">
						Members without 2FA are prompted at next login and blocked after 7 days.
					</div>
				</div>
				<Switch checked={enforce2fa} onCheckedChange={toggle2fa} aria-label="Enforce two-factor authentication" />
			</div>

			<div class="flex items-center gap-3">
				<div class="flex-1">
					<div class="text-[13px] font-medium">SSO-enforced mode</div>
					<div class="text-subtle-foreground mt-px text-[11.5px]">
						Password login is disabled for everyone; all sign-ins go through your identity provider.
					</div>
				</div>
				<Switch checked={ssoEnforced} onCheckedChange={toggleSso} aria-label="SSO-enforced mode" />
			</div>

			{#if ssoEnforced}
				<Alert.Root tone="info">
					<KeyRoundIcon />
					<Alert.Content>
						<Alert.Title>Break-glass recovery</Alert.Title>
						<Alert.Description>
							If your IdP goes down, instance admins can sign in at /auth/break-glass with a pre-generated recovery code
							(shown once when enforcement turns on) plus 2FA. Every break-glass login pages all other admins and is
							written to the audit log.
						</Alert.Description>
					</Alert.Content>
				</Alert.Root>
			{/if}
		</div>

		<form
			method="POST"
			action="?/save"
			class="self-start"
			use:enhance={() => async ({ result, update }) => {
				await update({ reset: false });
				if (result.type === 'success') toast.success('Security policies saved.');
			}}
		>
			<input type="hidden" name="sessionLifetime" value={sessionLifetime} />
			<input type="hidden" name="ipAllowlist" value={ipAllowlist} />
			<input type="hidden" name="enforce2fa" value={enforce2fa} />
			<input type="hidden" name="ssoEnforced" value={ssoEnforced} />
			<Button type="submit">
				<CheckIcon data-icon="inline-start" />
				Save policies
			</Button>
		</form>
	</div>
{/if}
