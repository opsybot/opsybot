<script lang="ts">
	import AppSidebar from '$lib/components/layout/app-sidebar.svelte';
	import CommandPalette from '$lib/components/layout/command-palette.svelte';
	import { setAppShell } from '$lib/components/layout/context.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();

	setAppShell(() => data.session);
</script>

<Sidebar.Provider open={data.sidebarOpen} class="h-svh min-h-0 overflow-hidden">
	<a
		href="#page-content"
		class="bg-popover text-popover-foreground focus:ring-ring sr-only rounded-md px-3 py-2 text-sm focus:not-sr-only focus:absolute focus:top-3 focus:left-3 focus:z-50 focus:ring-2"
	>
		Skip to content
	</a>

	<AppSidebar session={data.session} counts={data.counts} />

	<Sidebar.Inset id="page-content" tabindex={-1} class="min-h-0 overflow-hidden">
		{@render children()}
	</Sidebar.Inset>

	<CommandPalette />
</Sidebar.Provider>
