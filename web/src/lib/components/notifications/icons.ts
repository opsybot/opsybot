import type { Component } from 'svelte';
import type { LucideProps } from '@lucide/svelte';
import BellIcon from '@lucide/svelte/icons/bell';
import MailIcon from '@lucide/svelte/icons/mail';
import MessageSquareIcon from '@lucide/svelte/icons/message-square';
import SendIcon from '@lucide/svelte/icons/send';
import WebhookIcon from '@lucide/svelte/icons/webhook';

export const CHANNEL_ICONS: Record<string, Component<LucideProps>> = {
	'message-square': MessageSquareIcon,
	send: SendIcon,
	bell: BellIcon,
	mail: MailIcon,
	webhook: WebhookIcon
};
