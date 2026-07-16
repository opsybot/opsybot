import { env } from '$env/dynamic/private';

export type Scenario = 'active' | 'quiet' | 'not-on-call' | 'empty' | 'degraded';

const SCENARIOS: Scenario[] = ['active', 'quiet', 'not-on-call', 'empty', 'degraded'];

export function scenario(): Scenario {
	const value = env.OPSYBOT_FIXTURE as Scenario | undefined;
	return value && SCENARIOS.includes(value) ? value : 'active';
}

export type Deployment = 'cloud' | 'self-hosted';

export function deployment(): Deployment {
	return env.OPSYBOT_DEPLOYMENT === 'self-hosted' ? 'self-hosted' : 'cloud';
}
