<script lang="ts">
	import CheckIcon from '@lucide/svelte/icons/check';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';
	import UpgradeState from '$lib/components/enterprise/upgrade-state.svelte';
	import { ENT_PITCH } from '$lib/enterprise';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

{#if !data.licensed}
	<UpgradeState pitch={ENT_PITCH.roles} />
{:else if data.roles}
	<div class="flex max-w-[860px] flex-col gap-3.5">
		<div class="flex items-center gap-2">
			<span class="text-subtle-foreground text-[13px]">
				Plain-language permissions. Assignment happens on the
				<a href={ws('/workspace')} class="text-brand-foreground hover:underline">Members page</a>.
			</span>
			<div class="flex-1"></div>
			<Button
				size="sm"
				variant="secondary"
				onclick={() => toast('Custom roles start as a copy of an existing one — pick permissions row by row.')}
			>
				<PlusIcon data-icon="inline-start" />
				New custom role
			</Button>
		</div>

		<div class="bg-card overflow-hidden rounded-xl border">
			<div class="overflow-x-auto">
				<table class="w-full min-w-[720px] border-collapse text-[13px]">
					<thead>
						<tr class="text-subtle-foreground text-left text-[11px] tracking-[0.05em] uppercase">
							<th scope="col" class="min-w-[260px] py-2.5 pl-[18px] font-semibold">Permission</th>
							{#each data.roles.roles as role (role)}
								<th scope="col" class="px-3 py-2.5 text-center font-semibold whitespace-nowrap">{role}</th>
							{/each}
						</tr>
					</thead>
					<tbody>
						{#each data.roles.perms as row (row.perm)}
							<tr class="border-t">
								<th scope="row" class="text-foreground py-2.5 pl-[18px] text-left text-[13px] font-normal">{row.perm}</th>
								{#each row.grants as grant, i (i)}
									<td class="px-3 py-2.5 text-center">
										{#if grant}
											<CheckIcon class="mx-auto size-[14px] text-[var(--mint-500)]" />
											<span class="sr-only">granted</span>
										{:else}
											<span class="text-subtle-foreground text-[12px]" aria-hidden="true">—</span>
											<span class="sr-only">not granted</span>
										{/if}
									</td>
								{/each}
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
			<div class="text-subtle-foreground border-t px-4 py-[9px] text-[11.5px]">
				SSO, SCIM, and security policies stay with workspace admins — no role can grant them.
			</div>
		</div>
	</div>
{/if}
