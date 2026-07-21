import type { Tone } from '$lib/dashboard';

export type Deployment = 'cloud' | 'self-hosted';

export type PlanId = 'free' | 'team' | 'business';

export type Plan = {
	id: PlanId;
	name: string;
	price: string;
	period: string;
	caps: string[];
	note: string;
};

export const PLANS: Plan[] = [
	{
		id: 'free',
		name: 'Free',
		price: '€0',
		period: '',
		caps: ['5 responders', '1 status page', '30-day history'],
		note: 'For small teams getting started.'
	},
	{
		id: 'team',
		name: 'Team',
		price: '€29',
		period: '/month',
		caps: ['Unlimited responders', '3 status pages', '1-year history', 'All integrations'],
		note: 'Flat rate. No per-seat billing.'
	},
	{
		id: 'business',
		name: 'Business',
		price: '€99',
		period: '/month',
		caps: [
			'Unlimited responders',
			'Unlimited status pages',
			'SCIM + SSO enforcement',
			'Advanced audit + streaming',
			'Multi-workspace org'
		],
		note: 'Everything in the enterprise section.'
	}
];

export const PLAN_IDS = PLANS.map((plan) => plan.id);
export function isPlanId(value: string): value is PlanId {
	return (PLAN_IDS as string[]).includes(value);
}

export function getPlan(id: string): Plan | undefined {
	return PLANS.find((plan) => plan.id === id);
}

export function planRank(id: PlanId): number {
	return PLAN_IDS.indexOf(id);
}

export function planCta(plan: Plan, currentId: PlanId): { label: string; disabled: boolean; primary: boolean } {
	if (plan.id === currentId) return { label: 'Current plan', disabled: true, primary: false };
	const direction = planRank(plan.id) > planRank(currentId) ? 'Upgrade' : 'Downgrade';
	return { label: `${direction} to ${plan.name}`, disabled: false, primary: plan.id === 'business' };
}

export function proration(targetId: PlanId): { proratedNow: string } {
	if (targetId === 'business') return { proratedNow: '€44.33' };
	if (targetId === 'free') return { proratedNow: '−€18.37 credit' };
	return { proratedNow: '€0.00' };
}

export const NEXT_INVOICE_DATE = '2026-08-01';

export type BillingStatus = 'trial' | 'active' | 'past_due' | 'cancelled';

export type StatusBanner = { tone: Tone; icon: string; title: string; body: string; cta?: string };

export function statusBanner(status: BillingStatus, trialDaysLeft: number): StatusBanner | null {
	switch (status) {
		case 'trial':
			return {
				tone: 'info',
				icon: 'clock',
				title: 'Team trial',
				body: `${trialDaysLeft} days left. When it ends, the workspace goes read-only with a 14-day grace window; you can export everything any time. Paging never stops during the trial.`,
				cta: 'Add payment method'
			};
		case 'past_due':
			return {
				tone: 'warning',
				icon: 'triangle-alert',
				title: 'Trial ended: workspace read-only',
				body: 'You have a 14-day grace window to add a payment method before anything is scheduled for deletion. Export everything any time. Paging keeps working through the grace window.',
				cta: 'Add payment method'
			};
		case 'cancelled':
			return {
				tone: 'warning',
				icon: 'circle-off',
				title: 'Plan cancelled',
				body: `Billing has stopped. The workspace stays usable until ${NEXT_INVOICE_DATE}. Resubscribe any time before 2026-08-31 to keep your data.`
			};
		default:
			return null;
	}
}

export type Invoice = { id: string; date: string; amount: string; status: string };

export type UsageMeter = { kind: string; icon: string; used: number; included: number | 'unlimited' };

export function usagePercent(meter: UsageMeter): number | null {
	if (meter.included === 'unlimited') return null;
	if (meter.included <= 0) return 0;
	return Math.min(100, Math.round((meter.used / meter.included) * 100));
}

export function usageFill(pct: number): string {
	return pct > 85 ? 'var(--warning)' : 'var(--mint-500)';
}

export function formatCount(value: number): string {
	return value.toLocaleString('en-US');
}

export type LicenseStatus = 'active' | 'expiring' | 'expired' | 'none';

export type License = {
	plan: string;
	capacity: string;
	licensee: string;
	issued: string;
	expires: string;
	status: LicenseStatus;
	daysLeft: number;
};

export function licenseBadge(status: LicenseStatus): { tone: Tone; label: string } {
	switch (status) {
		case 'expired':
			return { tone: 'critical', label: 'expired' };
		case 'expiring':
			return { tone: 'warning', label: 'expiring' };
		case 'none':
			return { tone: 'neutral', label: 'community' };
		default:
			return { tone: 'success', label: 'active' };
	}
}

export function licenseAlert(license: License): { tone: 'critical' | 'warning' | 'info'; title: string; body: string } | null {
	if (license.status === 'expired')
		return {
			tone: 'critical',
			title: 'License expired',
			body: 'SSO enforcement, SCIM, and advanced audit are locked. On-call, alerting, and paging keep running. We never stop the pager. Renew to unlock the rest.'
		};
	if (license.status === 'expiring')
		return {
			tone: 'warning',
			title: `License expires in ${license.daysLeft} days`,
			body: `Renew before ${license.expires} to avoid feature locks. Paging always keeps working regardless of license state.`
		};
	if (license.status === 'none')
		return {
			tone: 'info',
			title: 'Running the AGPL core',
			body: 'The open-source core is the whole platform. On-call, alerting, and paging are unrestricted. Activate a license to unlock SSO enforcement, SCIM, and advanced audit.'
		};
	return null;
}

export function parseLicenseKey(form: FormData): { key: string } | { error: string } {
	const key = String(form.get('key') ?? '').replace(/\s+/g, '').toUpperCase().slice(0, 40);
	if (!/^[A-Z0-9-]{8,}$/.test(key)) return { error: 'That key is not valid. Check for typos or request a new one.' };
	return { key };
}

export type DeliveryChannelId = 'sms' | 'voice' | 'push';
export const DELIVERY_CHANNEL_IDS: DeliveryChannelId[] = ['sms', 'voice', 'push'];

export function isChannelId(value: string): value is DeliveryChannelId {
	return (DELIVERY_CHANNEL_IDS as string[]).includes(value);
}

export type DeliveryChannel = { id: DeliveryChannelId; label: string; transit: string; icon: string; on: boolean };

export type BillingProfile = { company: string; vat: string };

export function parseProfile(form: FormData): BillingProfile {
	return {
		company: String(form.get('company') ?? '').replace(/\s+/g, ' ').trim().slice(0, 120),
		vat: String(form.get('vat') ?? '').replace(/\s+/g, '').toUpperCase().slice(0, 20)
	};
}

export type PaymentMethod = { brand: string; last4: string; expires: string } | null;
