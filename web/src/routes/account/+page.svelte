<script lang="ts">
	import { untrack } from 'svelte';
	import Building2Icon from '@lucide/svelte/icons/building-2';
	import CheckIcon from '@lucide/svelte/icons/check';
	import InfoIcon from '@lucide/svelte/icons/info';
	import LogInIcon from '@lucide/svelte/icons/log-in';
	import MailIcon from '@lucide/svelte/icons/mail';
	import { toast } from 'svelte-sonner';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import { goto } from '$app/navigation';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import TimezoneSelect from '$lib/components/timezone-select.svelte';
	import { profileSchema } from '$lib/schemas/auth';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const form = superForm(untrack(() => data.form), {
		validators: zod4Client(profileSchema),
		onUpdated: ({ form }) => {
			if (form.message === 'saved')
				toast.success('Profile saved', { description: 'Your name and timezone are updated everywhere.' });
			else if (typeof form.message === 'string') toast.error(form.message);
		}
	});
	const { form: formData, errors, enhance } = form;

	const initials = $derived(
		$formData.name
			.split(' ')
			.map((p) => p[0])
			.filter(Boolean)
			.slice(0, 2)
			.join('')
			.toUpperCase() || 'U'
	);
</script>

<div style="display: flex; flex-direction: column; gap: 18px">
	<div class="acct-note">
		<InfoIcon size={15} style="color: var(--text-tertiary); flex-shrink: 0; margin-top: 1px" />
		<span>
			This is your <strong>personal account</strong>: the same you across every workspace and
			organization. Workspace-specific settings (notification rules, on-call, roles) live inside each
			workspace.
		</span>
	</div>

	<div class="acct-card">
		<header class="acct-card-head"><span class="acct-card-title">Personal details</span></header>
		<form method="POST" action="?/save" use:enhance class="acct-card-body">
			<div class="acct-grid2">
				<div>
					<label class="acct-field-label" for="acct-name">Full name</label>
					<Input id="acct-name" name="name" bind:value={$formData.name} autocomplete="name" />
					{#if $errors.name}<div class="acct-err">{$errors.name}</div>{/if}
				</div>
				<TimezoneSelect {form} name="timezone" label="Timezone" />
			</div>

			<div style="display: flex">
				<Button type="submit">
					<CheckIcon class="size-4" />
					Save changes
				</Button>
			</div>
		</form>
	</div>

	<div class="acct-card">
		<header class="acct-card-head"><span class="acct-card-title">Email address</span></header>
		<div class="acct-row">
			<span class="acct-ic"><MailIcon size={17} /></span>
			<div style="min-width: 0; flex: 1">
				<div style="font-size: 13.5px; font-weight: 500; display: flex; align-items: center; gap: 8px">
					{data.email}
					<Badge tone="success" size="sm">verified</Badge>
				</div>
				<div class="acct-mono" style="font-size: 11px; margin-top: 2px">
					sign-in address · used for password resets and pages
				</div>
			</div>
		</div>
	</div>

	<div class="acct-card">
		<header class="acct-card-head">
			<span class="acct-card-title">Workspaces you belong to</span>
			<span style="font-size: 11.5px; color: var(--text-tertiary); margin-left: 4px">
				{data.workspaces.length} · one account
			</span>
		</header>
		{#each data.workspaces as w (w.id)}
			<div class="acct-row">
				<span class="acct-ic"><Building2Icon size={16} /></span>
				<div style="min-width: 0; flex: 1">
					<div style="font-size: 13.5px; font-weight: 500">{w.name}</div>
					<div class="acct-mono" style="font-size: 11px; margin-top: 2px">
						{w.environment || 'production'} · workspace member
					</div>
				</div>
				<Button size="sm" variant="ghost" onclick={() => goto(`/${w.id}`)}>
					<LogInIcon class="size-3.5" />
					Open
				</Button>
			</div>
		{/each}
		<div class="acct-row" style="background: var(--ink-2)">
			<span style="font-size: 12px; color: var(--text-tertiary); flex: 1">
				Membership and roles are managed inside each workspace.
			</span>
		</div>
	</div>
</div>
