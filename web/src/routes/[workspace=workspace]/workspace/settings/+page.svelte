<script lang="ts">
	import { tick, untrack } from 'svelte';
	import CheckIcon from '@lucide/svelte/icons/check';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import CopyField from '$lib/components/alertsources/copy-field.svelte';
	import SegmentedToggle from '$lib/components/admin/segmented-toggle.svelte';
	import SettingsSection from '$lib/components/admin/settings-section.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { Switch } from '$lib/components/ui/switch';
	import Tag from '$lib/components/tag.svelte';
	import {
		CADENCE_OPTIONS,
		FIELD_TYPES,
		RETENTION_OPTIONS,
		SEVERITY_TONE,
		THRESHOLD_OPTIONS,
		TIMEZONE_OPTIONS
	} from '$lib/admin';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let name = $state(untrack(() => data.name));
	let timezone = $state(untrack(() => data.timezone));
	let severities = $state(untrack(() => data.severities.map((sev) => ({ ...sev }))));
	let cadence = $state(untrack(() => ({ ...data.cadence })));
	let ssoMode = $state<'oidc' | 'saml'>(untrack(() => data.sso.mode));
	let issuer = $state(untrack(() => data.sso.issuer));
	let clientId = $state(untrack(() => data.sso.clientId));
	let clientSecret = $state('');
	let clearClientSecret = $state(false);
	let samlMetadataUrl = $state(untrack(() => data.sso.samlMetadataUrl));
	let scopes = $state(untrack(() => data.sso.scopes));
	let allowedEmailDomains = $state(untrack(() => data.sso.allowedEmailDomains));
	let ssoEnabled = $state(untrack(() => data.sso.enabled));
	let ssoRequired = $state(untrack(() => data.sso.required));
	let jitProvisioning = $state(untrack(() => data.sso.jitProvisioning));
	const hasClientSecret = $derived(data.sso.hasClientSecret);
	const samlMetadataOut = $derived(
		`${page.url.origin}/v1/auth/sso/${page.params.workspace}/saml/metadata`
	);
	let retention = $state(untrack(() => ({ ...data.retention })));

	let threshold = $state(untrack(() => data.postmortemThreshold));
	$effect(() => {
		threshold = data.postmortemThreshold;
	});
	let thresholdForm: HTMLFormElement;
	async function changeThreshold(next: string) {
		threshold = next;
		await tick();
		thresholdForm.requestSubmit();
	}

	let newFieldName = $state('');
	let newFieldType = $state('text');
	let newFieldOptions = $state('');
	let addForm: HTMLFormElement;

	const settingsJson = $derived(
		JSON.stringify({
			name,
			timezone,
			severities: severities.map((sev) => ({ def: sev.def })),
			cadence,
			sso: {
				mode: ssoMode,
				issuer,
				clientId,
				clientSecret,
				clearClientSecret,
				samlMetadataUrl,
				scopes,
				allowedEmailDomains,
				enabled: ssoEnabled,
				required: ssoRequired,
				jitProvisioning
			},
			retention
		})
	);
</script>

{#snippet labeledSelect(label: string, current: string, options: string[], onpick: (value: string) => void, width: string)}
	<div class="flex flex-col gap-1.5" style="width:{width}">
		<span class="text-muted-foreground text-[13px] font-medium">{label}</span>
		<Select.Root type="single" value={current} onValueChange={onpick}>
			<Select.Trigger size="sm" aria-label={label}>{current}</Select.Trigger>
			<Select.Content>
				<Select.Group>
					{#each options as option (option)}
						<Select.Item value={option} label={option}>{option}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
	</div>
{/snippet}

<div class="flex max-w-[760px] flex-col gap-3.5">
	<SettingsSection title="General">
		<div class="flex flex-wrap gap-2.5">
			<Field.Field class="min-w-[220px] flex-1 gap-1.5 space-y-0">
				<Field.FieldLabel for="ws-name" class="text-muted-foreground text-[13px] font-medium">Workspace name</Field.FieldLabel>
				<Input id="ws-name" bind:value={name} />
			</Field.Field>
			{@render labeledSelect('Workspace timezone', timezone, TIMEZONE_OPTIONS, (v) => (timezone = v), '240px')}
		</div>
		<p class="text-subtle-foreground m-0 text-[11.5px]">
			The timezone sets day boundaries for metrics and reports. Timestamps everywhere stay UTC.
		</p>
	</SettingsSection>

	<SettingsSection title="Severities" note="definitions appear wherever a severity is picked">
		{#each severities as sev, index (sev.id)}
			<div class="flex items-start gap-2.5">
				<Badge tone={SEVERITY_TONE[sev.id]} size="sm" class="mt-1.5 shrink-0">{sev.id}</Badge>
				<Input class="flex-1" aria-label="{sev.id} definition" bind:value={severities[index].def} />
			</div>
		{/each}
		<div class="flex items-center gap-2.5">
			<form
				method="POST"
				action="?/setThreshold"
				bind:this={thresholdForm}
				use:enhance={() => async ({ result, update }) => {
					await update({ reset: false });
					if (result.type !== 'success') {
						threshold = data.postmortemThreshold;
						toast.error(String((result.type === 'failure' && result.data?.error) || 'Could not save.'));
						return;
					}
					toast(`Postmortems now required for ${threshold} and above.`);
				}}
			>
				<input type="hidden" name="threshold" value={threshold} />
				<div class="flex flex-col gap-1.5" style="width:200px">
					<span class="text-muted-foreground text-[13px] font-medium">Postmortem required for</span>
					<Select.Root type="single" value={threshold} onValueChange={changeThreshold}>
						<Select.Trigger size="sm" aria-label="Postmortem required for">{threshold}</Select.Trigger>
						<Select.Content>
							<Select.Group>
								{#each THRESHOLD_OPTIONS as option (option)}
									<Select.Item value={option} label={option}>{option}</Select.Item>
								{/each}
							</Select.Group>
						</Select.Content>
					</Select.Root>
				</div>
			</form>
			<span class="text-subtle-foreground pt-[22px] text-[12px]">and above, within 3 working days of resolve.</span>
		</div>
	</SettingsSection>

	<SettingsSection title="Custom incident fields">
		{#each data.fields as field (field.id)}
			<div class="bg-inset flex items-center gap-2.5 rounded-md border px-3 py-2.5">
				<div class="min-w-0 flex-1">
					<div class="text-[13px] font-medium">{field.name}</div>
					{#if field.options}<div class="text-subtle-foreground mt-0.5 font-mono text-[10.5px]">{field.options}</div>{/if}
				</div>
				<Tag>{field.type}</Tag>
				<form method="POST" action="?/removeField" use:enhance={() => async ({ update }) => update({ reset: false })}>
					<input type="hidden" name="id" value={field.id} />
					<Button type="submit" variant="ghost" size="icon-sm" aria-label="Remove {field.name}"><Trash2Icon /></Button>
				</form>
			</div>
		{/each}
		<form
			method="POST"
			action="?/addField"
			bind:this={addForm}
			use:enhance={() => async ({ result, update }) => {
				await update({ reset: false });
				if (result.type === 'success') {
					newFieldName = '';
					newFieldOptions = '';
					toast.success('Field added to every new incident.');
				}
			}}
			class="flex flex-wrap items-end gap-2"
		>
			<Field.Field class="min-w-[200px] flex-1 gap-1.5 space-y-0">
				<Field.FieldLabel for="new-field" class="text-muted-foreground text-[13px] font-medium">New field</Field.FieldLabel>
				<Input id="new-field" name="name" bind:value={newFieldName} placeholder="Deploy under suspicion" />
			</Field.Field>
			<div class="flex flex-col gap-1.5" style="width:140px">
				<span class="text-muted-foreground text-[13px] font-medium">Type</span>
				<input type="hidden" name="type" value={newFieldType} />
				<Select.Root type="single" value={newFieldType} onValueChange={(v) => (newFieldType = v)}>
					<Select.Trigger size="sm" aria-label="Field type">{newFieldType}</Select.Trigger>
					<Select.Content>
						<Select.Group>
							{#each FIELD_TYPES as option (option)}
								<Select.Item value={option} label={option}>{option}</Select.Item>
							{/each}
						</Select.Group>
					</Select.Content>
				</Select.Root>
			</div>
			{#if newFieldType === 'select' || newFieldType === 'multi-select'}
				<Field.Field class="min-w-[200px] flex-1 gap-1.5 space-y-0">
					<Field.FieldLabel for="new-field-options" class="text-muted-foreground text-[13px] font-medium">
						Options
					</Field.FieldLabel>
					<Input
						id="new-field-options"
						name="options"
						bind:value={newFieldOptions}
						placeholder="EU, US, APAC"
					/>
				</Field.Field>
			{/if}
			<Button type="submit" size="sm" variant="secondary" disabled={!newFieldName.trim()}>
				<PlusIcon data-icon="inline-start" />
				Add
			</Button>
		</form>
	</SettingsSection>

	<SettingsSection title="Update cadence defaults" note="how often a status update is due, per severity">
		<div class="flex flex-wrap gap-3">
			{#each Object.keys(cadence) as sev (sev)}
				{@render labeledSelect(sev, cadence[sev], CADENCE_OPTIONS, (v) => (cadence[sev] = v), '120px')}
			{/each}
		</div>
	</SettingsSection>

	<SettingsSection title="Single sign-on">
		<SegmentedToggle
			bind:value={ssoMode}
			label="SSO protocol"
			options={[
				{ value: 'oidc', label: 'OIDC' },
				{ value: 'saml', label: 'SAML' }
			]}
		/>

		<div class="flex flex-col gap-2.5">
			<label class="flex items-center justify-between gap-3" for="sso-enabled">
				<span class="flex flex-col">
					<span class="text-[13px] font-medium">Enabled</span>
					<span class="text-subtle-foreground text-[11.5px]">Turn on this workspace's SSO login.</span>
				</span>
				<Switch id="sso-enabled" bind:checked={ssoEnabled} aria-label="SSO enabled" />
			</label>
			<label class="flex items-center justify-between gap-3" for="sso-required">
				<span class="flex flex-col">
					<span class="text-[13px] font-medium">Require SSO</span>
					<span class="text-subtle-foreground text-[11.5px]">Members must sign in through the identity provider.</span>
				</span>
				<Switch id="sso-required" bind:checked={ssoRequired} aria-label="Require SSO" />
			</label>
			<label class="flex items-center justify-between gap-3" for="sso-jit">
				<span class="flex flex-col">
					<span class="text-[13px] font-medium">Create accounts on first sign-in</span>
					<span class="text-subtle-foreground text-[11.5px]">Just-in-time provision members the first time they log in.</span>
				</span>
				<Switch id="sso-jit" bind:checked={jitProvisioning} aria-label="Just-in-time provisioning" />
			</label>
		</div>

		{#if ssoMode === 'oidc'}
			<div class="flex flex-col gap-2.5">
				<Field.Field class="gap-1.5 space-y-0">
					<Field.FieldLabel for="sso-issuer" class="text-muted-foreground text-[13px] font-medium">Issuer URL</Field.FieldLabel>
					<Input id="sso-issuer" class="font-mono text-[12px]" bind:value={issuer} placeholder="https://idp.example.com" />
				</Field.Field>
				<div class="flex flex-wrap gap-2.5">
					<Field.Field class="min-w-[180px] flex-1 gap-1.5 space-y-0">
						<Field.FieldLabel for="sso-client" class="text-muted-foreground text-[13px] font-medium">Client ID</Field.FieldLabel>
						<Input id="sso-client" class="font-mono text-[12px]" bind:value={clientId} />
					</Field.Field>
					<Field.Field class="min-w-[180px] flex-1 gap-1.5 space-y-0">
						<Field.FieldLabel for="sso-secret" class="text-muted-foreground text-[13px] font-medium">Client secret</Field.FieldLabel>
						<Input
							id="sso-secret"
							type="password"
							bind:value={clientSecret}
							disabled={clearClientSecret}
							autocomplete="off"
							placeholder={hasClientSecret ? 'Stored. Leave blank to keep it' : 'Client secret'}
						/>
					</Field.Field>
				</div>
				{#if hasClientSecret}
					<label class="text-muted-foreground flex items-center gap-2 text-[12px]">
						<input type="checkbox" bind:checked={clearClientSecret} class="accent-primary size-3.5" />
						Remove the stored client secret
					</label>
				{/if}
				<Field.Field class="gap-1.5 space-y-0">
					<Field.FieldLabel for="sso-scopes" class="text-muted-foreground text-[13px] font-medium">Scopes</Field.FieldLabel>
					<Input id="sso-scopes" class="font-mono text-[12px]" bind:value={scopes} placeholder="openid email profile" />
					<Field.FieldDescription class="text-subtle-foreground text-[11.5px]">
						Space- or comma-separated. Leave blank for the defaults.
					</Field.FieldDescription>
				</Field.Field>
			</div>
		{:else}
			<div class="flex flex-col gap-2.5">
				<Field.Field class="gap-1.5 space-y-0">
					<Field.FieldLabel for="sso-meta" class="text-muted-foreground text-[13px] font-medium">IdP metadata URL</Field.FieldLabel>
					<Input id="sso-meta" class="font-mono text-[12px]" bind:value={samlMetadataUrl} placeholder="https://sso.example.com/saml/metadata" />
				</Field.Field>
				<CopyField label="Opsybot metadata. Give this to your IdP" value={samlMetadataOut} />
			</div>
		{/if}

		<Field.Field class="gap-1.5 space-y-0">
			<Field.FieldLabel for="sso-domains" class="text-muted-foreground text-[13px] font-medium">Allowed email domains</Field.FieldLabel>
			<textarea
				id="sso-domains"
				bind:value={allowedEmailDomains}
				rows="2"
				placeholder="acme.com"
				class="border-border-strong bg-inset text-foreground focus-visible:border-primary focus-visible:shadow-[var(--focus-ring)] w-full rounded-sm border px-3 py-2 font-mono text-[12px] outline-none"
			></textarea>
			<Field.FieldDescription class="text-subtle-foreground text-[11.5px]">
				One per line. Leave blank to allow any domain.
			</Field.FieldDescription>
		</Field.Field>
	</SettingsSection>

	<SettingsSection title="Data retention" note="self-hosted. You own the disk">
		<div class="flex flex-wrap gap-3">
			{@render labeledSelect('Alert history', retention.alerts, RETENTION_OPTIONS.alerts, (v) => (retention.alerts = v), '160px')}
			{@render labeledSelect('Incident history', retention.incidents, RETENTION_OPTIONS.incidents, (v) => (retention.incidents = v), '160px')}
			{@render labeledSelect('Audit log', retention.audit, RETENTION_OPTIONS.audit, (v) => (retention.audit = v), '160px')}
		</div>
	</SettingsSection>

	<form
		method="POST"
		action="?/save"
		class="self-start"
		use:enhance={() => async ({ result, update }) => {
			await update({ reset: false });
			if (result.type !== 'success') return;
			name = data.name;
			timezone = data.timezone;
			severities = data.severities.map((sev) => ({ ...sev }));
			cadence = { ...data.cadence };
			ssoMode = data.sso.mode;
			issuer = data.sso.issuer;
			clientId = data.sso.clientId;
			clientSecret = '';
			clearClientSecret = false;
			samlMetadataUrl = data.sso.samlMetadataUrl;
			scopes = data.sso.scopes;
			allowedEmailDomains = data.sso.allowedEmailDomains;
			ssoEnabled = data.sso.enabled;
			ssoRequired = data.sso.required;
			jitProvisioning = data.sso.jitProvisioning;
			retention = { ...data.retention };
			toast.success('Workspace settings saved.');
		}}
	>
		<input type="hidden" name="settings" value={settingsJson} />
		<Button type="submit">
			<CheckIcon data-icon="inline-start" />
			Save settings
		</Button>
	</form>
</div>
