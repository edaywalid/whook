# whook landing page

Marketing site for whook, built with TanStack Start (React 19), Tailwind v4, and
Motion. It reuses the explainer diagrams from `../docs/diagrams` as real section
visuals and carries the project's teal brand.

## Develop

```sh
npm install
npm run dev      # http://localhost:3000
```

## Build

```sh
npm run build    # outputs dist/ (client + SSR server)
npm run test
```

## Layout

- `src/routes/__root.tsx` sets the document shell, fonts, and metadata, and wraps
  the tree in `LazyMotion` so animation features load once.
- `src/routes/index.tsx` composes the page from sections in
  `src/components/landing/`.
- `src/styles.css` holds the Tailwind v4 theme tokens (the teal / slate / bone
  brand).
