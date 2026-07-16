import type {
	BillingProfile,
	BillingStatus,
	DeliveryChannel,
	Invoice,
	License,
	PaymentMethod,
	PlanId,
	UsageMeter
} from '$lib/billing';
import { scenario } from './fixtures';

const INVOICES: Invoice[] = [
	{ id: 'INV-2026-07', date: '2026-07-01', amount: '€29.00', status: 'paid' },
	{ id: 'INV-2026-06', date: '2026-06-01', amount: '€29.00', status: 'paid' },
	{ id: 'INV-2026-05', date: '2026-05-01', amount: '€29.00', status: 'paid' },
	{ id: 'INV-2026-04', date: '2026-04-01', amount: '€29.00', status: 'paid' }
];

const USAGE = (): UsageMeter[] => [
	{ kind: 'SMS', icon: 'message-square', used: 412, included: 1000 },
	{ kind: 'Voice', icon: 'phone', used: 38, included: 100 },
	{ kind: 'Push', icon: 'smartphone', used: 5820, included: 'unlimited' }
];

const CHANNELS = (): DeliveryChannel[] => [
	{ id: 'sms', label: 'SMS', transit: 'phone number + message text', icon: 'message-square', on: true },
	{ id: 'voice', label: 'Voice', transit: 'phone number + spoken alert summary', icon: 'phone', on: true },
	{ id: 'push', label: 'Push', transit: 'device token (handled locally, never leaves your instance)', icon: 'smartphone', on: true }
];

const ACTIVE_LICENSE = (): License => ({
	plan: 'Business (self-hosted)',
	capacity: 'unlimited responders · 3 workspaces',
	licensee: 'Acme Corp GmbH',
	issued: '2026-01-14',
	expires: '2027-01-14',
	status: 'active',
	daysLeft: 183
});

type Store = {
	currentPlanId: PlanId;
	billingStatus: BillingStatus;
	trialDaysLeft: number;
	cancelReason: string;
	payment: PaymentMethod;
	profile: BillingProfile;
	invoices: Invoice[];
	usage: UsageMeter[];
	license: License;
	deliveryLinked: boolean;
	channels: DeliveryChannel[];
};

function seed(): Store {
	return {
		currentPlanId: 'team',
		billingStatus: 'active',
		trialDaysLeft: 9,
		cancelReason: '',
		payment: { brand: 'Visa', last4: '4242', expires: '08/2028' },
		profile: { company: 'Acme Corp GmbH', vat: 'DE123456789' },
		invoices: [...INVOICES],
		usage: USAGE(),
		license: ACTIVE_LICENSE(),
		deliveryLinked: true,
		channels: CHANNELS()
	};
}

const store = seed();
const state = scenario();

if (state === 'active' || state === 'quiet') {
	store.billingStatus = 'trial';
	store.trialDaysLeft = 9;
	store.payment = null;
	store.invoices = [];
}

if (state === 'empty') {
	store.billingStatus = 'trial';
	store.trialDaysLeft = 14;
	store.payment = null;
	store.profile = { company: '', vat: '' };
	store.invoices = [];
	store.usage = store.usage.map((meter) => ({ ...meter, used: 0 }));
	store.deliveryLinked = false;
	store.license = {
		plan: 'Community (AGPL core)',
		capacity: 'unlimited responders · unrestricted paging',
		licensee: '—',
		issued: '—',
		expires: '—',
		status: 'none',
		daysLeft: 0
	};
}
if (state === 'quiet') {
	store.license = { ...store.license, status: 'expiring', expires: '2026-07-21', daysLeft: 6 };
}
if (state === 'degraded') {
	store.billingStatus = 'past_due';
	store.trialDaysLeft = 0;
	store.payment = null;
	store.invoices = [];
	store.license = { ...store.license, status: 'expired', expires: '2026-06-14', daysLeft: 0 };
}

export function getBilling() {
	return {
		currentPlanId: store.currentPlanId,
		status: store.billingStatus,
		trialDaysLeft: store.trialDaysLeft,
		payment: store.payment,
		profile: store.profile
	};
}

export function getAccount() {
	return {
		payment: store.payment,
		profile: store.profile,
		usage: store.usage,
		invoices: store.invoices
	};
}

export function getLicense(): License {
	return store.license;
}

export function getDelivery() {
	return { linked: store.deliveryLinked, channels: store.channels, usage: store.usage };
}

export function changePlan(planId: PlanId): boolean {
	store.currentPlanId = planId;
	if (store.billingStatus === 'trial' || store.billingStatus === 'past_due') {
		store.billingStatus = 'active';
		store.payment = store.payment ?? { brand: 'Visa', last4: '4242', expires: '08/2028' };
	}
	return true;
}

export function saveProfile(profile: BillingProfile): void {
	store.profile = profile;
}

export function cancelPlan(reason: string): boolean {
	store.billingStatus = 'cancelled';
	store.cancelReason = reason;
	return true;
}

export function activateLicense(): boolean {
	store.license = ACTIVE_LICENSE();
	return true;
}

export function setDeliveryLinked(linked: boolean): boolean {
	store.deliveryLinked = linked;
	return true;
}

export function setChannel(id: string, on: boolean): boolean {
	const channel = store.channels.find((entry) => entry.id === id);
	if (!channel) return false;
	channel.on = on;
	return true;
}
