import type { Component } from 'svelte';
import type { LucideProps } from '@lucide/svelte';
import BracesIcon from '@lucide/svelte/icons/braces';
import ChartLineIcon from '@lucide/svelte/icons/chart-line';
import FlameIcon from '@lucide/svelte/icons/flame';
import GlobeIcon from '@lucide/svelte/icons/globe';
import HeartPulseIcon from '@lucide/svelte/icons/heart-pulse';

export const ICON: Record<string, Component<LucideProps>> = {
	flame: FlameIcon,
	'chart-line': ChartLineIcon,
	globe: GlobeIcon,
	braces: BracesIcon,
	'heart-pulse': HeartPulseIcon
};
