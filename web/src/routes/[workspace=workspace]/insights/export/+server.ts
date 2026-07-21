import { COHORT_FLOOR, parseFilters, scopeLabel, toCsv, type Filters } from '$lib/insights';
import {
	getAlertAnalytics,
	getDefinitions,
	getFollowupCompletion,
	getOnCallLoad,
	getOverview,
	insightsAvailable
} from '$lib/server/insights';
import type { RequestHandler } from './$types';

type Tab = 'overview' | 'alerts' | 'load' | 'followups' | 'definitions';
const TABS: Tab[] = ['overview', 'alerts', 'load', 'followups', 'definitions'];

function rowsFor(tab: Tab, filters: Filters): (string | number)[][] {
	switch (tab) {
		case 'alerts': {
			const alerts = getAlertAnalytics(filters);
			return [['Metric', 'Value', 'Detail'], ...alerts.stats.map((stat) => [stat.key, stat.value, stat.note])];
		}
		case 'load': {
			const load = getOnCallLoad(filters);
			if (load.withheld) {
				return [['Notice'], [`Fewer than ${COHORT_FLOOR} people match this filter, on-call load is withheld.`]];
			}
			return [
				['Person', 'Team', 'On-call hours', 'Pages', 'Night pages', 'Weekend pages'],
				...load.rows.map((row) => [row.name, row.team, row.hours, row.pages, row.night, row.weekend])
			];
		}
		case 'followups': {
			const followups = getFollowupCompletion(filters);
			return [
				['Metric', 'Value'],
				...followups.stats.map((stat) => [stat.key, stat.value]),
				[],
				['Team', 'Completion %'],
				...followups.byTeam.map((team) => [team.team, `${team.pct}%`])
			];
		}
		case 'definitions':
			return [['Term', 'Definition'], ...getDefinitions().map((entry) => [entry.term, entry.definition])];
		default: {
			const overview = getOverview(filters);
			return [
				['Metric', 'Value', 'Change', 'Comparison'],
				...overview.metrics.map((metric) => [metric.label, metric.value, metric.delta, overview.comparison]),
				[],
				['Stage', 'Median'],
				...overview.stages.map((stage) => [stage.label, stage.value])
			];
		}
	}
}

export const GET: RequestHandler = ({ url }) => {
	const filters = parseFilters(url);
	const requested = url.searchParams.get('tab');
	const tab: Tab = TABS.includes(requested as Tab) ? (requested as Tab) : 'overview';

	const rows =
		tab === 'definitions' || insightsAvailable()
			? [['Scope', scopeLabel(filters)], [], ...rowsFor(tab, filters)]
			: [['Scope', scopeLabel(filters)], [], ['Notice'], ['No incidents resolved yet: nothing to measure.']];

	return new Response(toCsv(rows), {
		headers: {
			'content-type': 'text/csv; charset=utf-8',
			'content-disposition': `attachment; filename="insights-${tab}-${filters.range}.csv"`
		}
	});
};
