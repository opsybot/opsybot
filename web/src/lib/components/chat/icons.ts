import type { Component } from 'svelte';
import type { LucideProps } from '@lucide/svelte';
import MessageSquareIcon from '@lucide/svelte/icons/message-square';

export const CHAT_ICONS: Record<string, Component<LucideProps>> = {
	'message-square': MessageSquareIcon
};
