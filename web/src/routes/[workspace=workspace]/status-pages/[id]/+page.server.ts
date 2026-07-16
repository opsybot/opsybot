import { error, fail, redirect } from '@sveltejs/kit';
import { COMPONENT_STATE_LABEL, COMPONENT_STATE_TONE, type Visibility } from '$lib/statuspages';
import {
	deletePage,
	getPage,
	moveComponent,
	pageNameTaken,
	rotateToken,
	setPublished,
	updatePage
} from '$lib/server/statuspages';
import type { Actions, PageServerLoad } from './$types';

const VISIBILITIES: Visibility[] = ['public', 'token', 'auth'];

export const load: PageServerLoad = ({ params }) => {
	const page = getPage(params.id);
	if (!page) error(404, `No status page called ${params.id}.`);

	return {
		page: {
			...page,
			components: page.components.map((component) => ({
				...component,
				stateLabel: COMPONENT_STATE_LABEL[component.state],
				stateTone: COMPONENT_STATE_TONE[component.state]
			}))
		}
	};
};

export const actions: Actions = {
	save: async ({ request, params }) => {
		const form = await request.formData();

		const name = String(form.get('name') ?? '').trim();
		if (!name) return fail(400, { error: 'Give the page a name.' });

		const domain = String(form.get('domain') ?? '').trim();
		// Hostname: dot-separated labels, each starting and ending alphanumeric
		if (!/^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$/.test(domain)) {
			return fail(400, { error: 'Enter a hostname, like status.acme.dev.' });
		}
		if (pageNameTaken(domain, params.id)) {
			return fail(400, { error: 'Another page already uses that domain.' });
		}

		const visibility = VISIBILITIES.find((option) => option === form.get('visibility')) ?? 'public';

		updatePage(params.id, {
			name,
			description: String(form.get('description') ?? '').trim(),
			pageTitle: String(form.get('pageTitle') ?? '').trim(),
			domain,
			visibility,
			accent: String(form.get('accent') ?? 'mint'),
			utcDefault: form.get('utcDefault') === 'on',
			showUptime: form.get('showUptime') === 'on',
			allowIndexing: form.get('allowIndexing') === 'on'
		});

		if (domain !== params.id) redirect(303, `/${params.workspace}/status-pages/${domain}`);
		return { saved: true };
	},

	moveComponent: async ({ request, params }) => {
		const form = await request.formData();
		moveComponent(params.id, String(form.get('id')), form.get('by') === 'up' ? -1 : 1);
	},

	rotateToken: async ({ params }) => {
		const token = rotateToken(params.id);
		return { token };
	},

	unpublish: async ({ params }) => {
		setPublished(params.id, false);
	},

	publish: async ({ params }) => {
		setPublished(params.id, true);
	},

	remove: async ({ params }) => {
		deletePage(params.id);
		redirect(303, `/${params.workspace}/status-pages`);
	}
};
