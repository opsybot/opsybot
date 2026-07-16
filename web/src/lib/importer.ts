export const IMP_STEPS = ['Connect', 'Dry run', 'Decisions', 'Import', 'Verify', 'Test page', 'Cutover'];

export type DecisionKind = 'user' | 'integration' | 'routing';

export type DryRunCreated = { kind: string; n: number; note: string };
export type DecisionChoice = { value: string; label: string };
export type Decision = { id: string; title: string; detail: string; kind: DecisionKind; choices: DecisionChoice[] };
export type Skipped = { title: string; reason: string };
export type DryRun = { created: DryRunCreated[]; decisions: Decision[]; skipped: Skipped[] };

export type CompareRow = { schedule: string; opsy: string; og: string; match: boolean };
export type CutoverSource = { source: string; from: string; to: string };

export type ImportPlan = { dryrun: DryRun; compare: CompareRow[]; cutover: CutoverSource[] };

export function isValidApiKey(key: string): boolean {
	return key.trim().length >= 8;
}
