import type { AiFeatureId, Model, ModelDraft } from '$lib/ai';
import { USE_DEFAULT, isAiFeature, uid } from '$lib/ai';
import { scenario } from './fixtures';

type Store = {
	enabled: boolean;
	defaultModelId: string | null;
	models: Model[];
	assignments: Record<AiFeatureId, string>;
};

function seed(): Store {
	const models: Model[] = [
		{ id: 'ollama-local', name: 'ollama-prod (self-hosted)', endpoint: 'http://10.0.4.12:11434 · llama-3.3-70b', health: 'ok', latency: '1.9 s' },
		{ id: 'claude', name: 'claude-api', endpoint: 'api.anthropic.com · claude-sonnet-4-5', health: 'ok', latency: '0.8 s' },
		{ id: 'legacy', name: 'gpu-box-2', endpoint: 'http://10.0.4.7:8080 · mixtral-8x7b', health: 'fail', latency: 'timeout 30 s' }
	];
	const assignments: Record<AiFeatureId, string> = { summaries: USE_DEFAULT, postmortems: USE_DEFAULT, correlation: 'claude' };
	return { enabled: true, defaultModelId: 'ollama-local', models, assignments };
}

const store = seed();

const state = scenario();
if (state === 'empty') {
	store.models = [];
	store.defaultModelId = null;
	store.enabled = false;
	store.assignments = { summaries: USE_DEFAULT, postmortems: USE_DEFAULT, correlation: USE_DEFAULT };
}
if (state === 'degraded') {
	const def = store.models.find((model) => model.id === store.defaultModelId);
	if (def) {
		def.health = 'fail';
		def.latency = 'timeout 30 s';
	}
}

function get(id: string): Model | undefined {
	return store.models.find((model) => model.id === id);
}

export function getAiSettings() {
	return {
		enabled: store.enabled,
		defaultModelId: store.defaultModelId,
		models: store.models,
		assignments: store.assignments
	};
}

export function setEnabled(on: boolean): boolean {
	if (on && store.models.length === 0) return false;
	store.enabled = on;
	return true;
}

export function setDefault(id: string): boolean {
	if (!get(id)) return false;
	store.defaultModelId = id;
	return true;
}

export function assignFeature(feature: string, modelId: string): boolean {
	if (!isAiFeature(feature)) return false;
	if (modelId !== USE_DEFAULT && !get(modelId)) return false;
	store.assignments[feature] = modelId;
	return true;
}

export function addModel(draft: ModelDraft): Model {
	const model: Model = { id: uid(), name: draft.name, endpoint: draft.endpoint, health: 'ok', latency: '1.9 s' };
	store.models.push(model);
	if (store.defaultModelId === null) store.defaultModelId = model.id;
	return model;
}

export function removeModel(id: string): boolean {
	const index = store.models.findIndex((model) => model.id === id);
	if (index < 0) return false;
	store.models.splice(index, 1);
	if (store.defaultModelId === id) store.defaultModelId = store.models[0]?.id ?? null;
	for (const feature of Object.keys(store.assignments) as AiFeatureId[]) {
		if (store.assignments[feature] === id) store.assignments[feature] = USE_DEFAULT;
	}
	if (store.models.length === 0) store.enabled = false;
	return true;
}
