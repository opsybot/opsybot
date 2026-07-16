<script lang="ts">
	import { page } from '$app/state';
	import { Badge } from '$lib/components/ui/badge';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import { isCurrentSection, navigation, workspacePath, ws } from '$lib/navigation';
	import type { Session } from '$lib/session';
	import UserCard from './user-card.svelte';
	import Wordmark from './wordmark.svelte';
	import WorkspaceSwitcher from './workspace-switcher.svelte';

	let {
		session,
		counts
	}: {
		session: Session;
		counts: { openIncidents: number };
	} = $props();

	const current = $derived(workspacePath(page.url.pathname, page.params.workspace ?? ''));
</script>

<Sidebar.Root>
	<Sidebar.Header
		class="border-sidebar-border h-14 shrink-0 flex-row items-center gap-[9px] border-b px-[18px]"
	>
		<Wordmark />
		<WorkspaceSwitcher {session} />
	</Sidebar.Header>

	<Sidebar.Content
		role="navigation"
		aria-label="Sections"
		class="gap-[2px] px-2.5 pt-2.5 pb-3.5"
	>
		{#each navigation as section (section.label ?? 'primary')}
			<Sidebar.Group class="gap-[2px] p-0">
				{#if section.label}
					<Sidebar.GroupLabel
						class="text-subtle-foreground tracking-label h-auto px-[11px] pt-3.5 pb-[5px] text-[10.5px] font-semibold uppercase"
					>
						{section.label}
					</Sidebar.GroupLabel>
				{/if}
				<Sidebar.GroupContent>
					<Sidebar.Menu class="gap-[2px]">
						{#each section.items as item (item.href)}
							{@const active = isCurrentSection(current, item.href)}
							<Sidebar.MenuItem>
								<Sidebar.MenuButton
									isActive={active}
									class="h-auto gap-[11px] rounded-sm px-[11px] py-2 text-[13.5px] leading-[18px] font-medium"
								>
									{#snippet child({ props })}
										<a href={ws(item.href)} aria-current={active ? 'page' : undefined} {...props}>
											<item.icon />
											<span class="flex-1">{item.title}</span>
											{#if item.href === '/incidents' && counts.openIncidents > 0}
												<Badge tone="critical" variant="solid" size="sm">
													{counts.openIncidents}
												</Badge>
											{/if}
										</a>
									{/snippet}
								</Sidebar.MenuButton>
							</Sidebar.MenuItem>
						{/each}
					</Sidebar.Menu>
				</Sidebar.GroupContent>
			</Sidebar.Group>
		{/each}
	</Sidebar.Content>

	<Sidebar.Footer class="border-sidebar-border shrink-0 border-t p-2.5">
		<UserCard user={session.user} />
	</Sidebar.Footer>
</Sidebar.Root>
