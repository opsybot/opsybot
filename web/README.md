# Opsybot web

SvelteKit frontend for Opsybot. Svelte 5, Tailwind 4, shadcn-svelte, pnpm.

The UI is currently view-only: screens render from fixture stores in
`src/lib/server/` until the Go API replaces them.

```sh
pnpm install
pnpm dev
```

`pnpm check` runs svelte-check. `pnpm build` produces the production bundle.
