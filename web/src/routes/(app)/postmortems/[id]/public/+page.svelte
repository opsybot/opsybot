<script lang="ts">
	import EyeIcon from '@lucide/svelte/icons/eye';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import Page from '$lib/components/layout/page.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<Page title="Postmortems" subtitle="Blameless, drafted from the timeline">
	<div class="flex max-w-[720px] flex-col gap-3.5">
		<div
			class="bg-info-wash border-info-edge text-muted-foreground flex items-center gap-[9px] rounded-md border px-3.5 py-[9px] text-[12.5px]"
		>
			<EyeIcon class="text-info-ink size-[13px] shrink-0" />
			<span>
				Previewing the public page — this is exactly what a visitor sees. No names, no internal links.
			</span>
			<a
				href="/postmortems/{data.id}"
				class="text-muted-foreground hover:text-brand-foreground ml-auto shrink-0 text-[12.5px]"
			>
				Back to editor
			</a>
		</div>

		{#if !data.live || !data.published}
			<div
				class="bg-warning-wash border-warning-edge text-muted-foreground flex items-center gap-[9px] rounded-md border px-3.5 py-[9px] text-[12.5px]"
			>
				<TriangleAlertIcon class="text-warning-ink size-[13px] shrink-0" />
				<span>
					{#if !data.published}
						Nobody can reach this yet — the postmortem has not been published.
					{:else}
						Nobody can reach this yet — the public link is switched off.
					{/if}
				</span>
			</div>
		{/if}

		<article class="rounded-xl border bg-[var(--ink-0)] px-6 py-8 sm:px-12 sm:py-11">
			<header class="mb-7">
				<div class="tracking-display mb-[22px] text-[18px] font-bold">
					{data.organization}<span class="text-primary">.</span>dev
					<span class="text-subtle-foreground font-normal">/ postmortems</span>
				</div>

				<h1 class="tracking-display m-0 text-[30px] leading-[1.2] font-light">{data.title}</h1>

				<div class="text-subtle-foreground mt-2.5 font-mono text-xs">
					{data.date} · impact window {data.window} · {data.resolved ? 'resolved' : 'ongoing'}
				</div>
			</header>

			<div
				class="mb-1.5 grid grid-cols-[repeat(auto-fit,minmax(130px,1fr))] gap-3 border-y py-3.5"
			>
				{#each data.facts as fact (fact.label)}
					<div class="flex flex-col gap-[3px]">
						<span class="text-subtle-foreground tracking-[0.07em] text-[10.5px] uppercase">
							{fact.label}
						</span>
						<strong class="text-sm font-semibold">{fact.value}</strong>
					</div>
				{/each}
			</div>

			<section>
				<h2 class="tracking-heading mt-[26px] mb-2 text-[15px] font-semibold">What happened</h2>
				<p class="text-muted-foreground m-0 text-sm leading-[1.7]">{data.summary}</p>
			</section>

			<section>
				<h2 class="tracking-heading mt-[26px] mb-2 text-[15px] font-semibold">Impact</h2>
				<p class="text-muted-foreground m-0 text-sm leading-[1.7]">{data.impact}</p>
			</section>

			{#if data.factors.length}
				<section>
					<h2 class="tracking-heading mt-[26px] mb-2 text-[15px] font-semibold">
						Contributing factors
					</h2>
					{#each data.factors as factor (factor)}
						<p class="text-muted-foreground m-0 text-sm leading-[1.7]">{factor}</p>
					{/each}
				</section>
			{/if}

			{#if data.changes.length}
				<section>
					<h2 class="tracking-heading mt-[26px] mb-2 text-[15px] font-semibold">
						What we are changing
					</h2>
					<ul class="m-0 flex list-disc flex-col gap-1.5 pl-5">
						{#each data.changes as change (change)}
							<li class="text-muted-foreground text-sm leading-[1.7]">{change}</li>
						{/each}
					</ul>
				</section>
			{/if}

			<footer class="text-subtle-foreground mt-8 border-t pt-4 text-xs">
				Published {data.date} · {data.organization}.dev engineering
			</footer>
		</article>
	</div>
</Page>
