import type { Component } from 'svelte';
import type { LucideProps } from '@lucide/svelte';
import FileTextIcon from '@lucide/svelte/icons/file-text';
import GlobeIcon from '@lucide/svelte/icons/globe';
import ListChecksIcon from '@lucide/svelte/icons/list-checks';
import MegaphoneIcon from '@lucide/svelte/icons/megaphone';
import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
import UsersIcon from '@lucide/svelte/icons/users';
import WebhookIcon from '@lucide/svelte/icons/webhook';

export const ICON: Record<string, Component<LucideProps>> = {
	megaphone: MegaphoneIcon,
	users: UsersIcon,
	'file-text': FileTextIcon,
	webhook: WebhookIcon,
	'list-checks': ListChecksIcon,
	globe: GlobeIcon,
	'shield-check': ShieldCheckIcon
};
