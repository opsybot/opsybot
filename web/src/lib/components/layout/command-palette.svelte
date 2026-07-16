<script lang="ts">
	import * as Command from '$lib/components/ui/command';
	import { navigation } from '$lib/navigation';
	import { useAppShell } from './context.svelte';

	const shell = useAppShell();

	function onkeydown(event: KeyboardEvent) {
		if (event.key.toLowerCase() === 'k' && (event.metaKey || event.ctrlKey)) {
			event.preventDefault();
			shell.toggleCommand();
		}
	}
</script>

<svelte:window {onkeydown} />

<Command.Dialog
	bind:open={shell.commandOpen}
	title="Command palette"
	description="Jump to any page in the workspace."
>
	<Command.Input placeholder="Where to?" />
	<Command.List>
		<Command.Empty>Nothing matches that.</Command.Empty>
		{#each navigation as section, index (section.label ?? 'primary')}
			{#if index > 0}
				<Command.Separator />
			{/if}
			<Command.Group heading={section.label ?? 'Go to'}>
				{#each section.items as item (item.href)}
					<Command.LinkItem href={item.href} onSelect={() => (shell.commandOpen = false)}>
						<item.icon />
						<span>{item.title}</span>
					</Command.LinkItem>
				{/each}
			</Command.Group>
		{/each}
	</Command.List>
</Command.Dialog>
