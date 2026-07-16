import BookOpenIcon from '@lucide/svelte/icons/book-open';
import ChartLineIcon from '@lucide/svelte/icons/chart-line';
import GitPullRequestIcon from '@lucide/svelte/icons/git-pull-request';
import type { Component } from 'svelte';
import type { LucideProps } from '@lucide/svelte';

export type LinkKind = 'runbook' | 'dashboard' | 'repository';

export const LINK_KINDS: {
	kind: LinkKind;
	label: string;
	icon: Component<LucideProps>;
	placeholder: string;
}[] = [
	{ kind: 'runbook', label: 'Runbook', icon: BookOpenIcon, placeholder: 'runbooks/…' },
	{ kind: 'dashboard', label: 'Dashboard', icon: ChartLineIcon, placeholder: 'grafana/…' },
	{ kind: 'repository', label: 'Repository', icon: GitPullRequestIcon, placeholder: 'acme/…' }
];

export type Service = {
	id: string;
	team: string;
	description: string;
	links: Record<LinkKind, string>;
	deps: string[];
	statusComponents: string[];
};

export function dependentsOf(serviceId: string, all: Service[]): string[] {
	return all
		.filter((service) => service.deps.includes(serviceId))
		.map((service) => service.id)
		.sort();
}

export const CATALOG_TEAMS = ['payments', 'platform', 'frontend'];
