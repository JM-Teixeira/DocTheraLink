# Klyniva frontend

React + TypeScript + Vite + Tailwind, data with TanStack Query, routes with
raviger, texts with i18next (`public/locale/en.json`, `pt.json`). Architecture
and conventions: root [CONTEXT.md](../CONTEXT.md).

```bash
npm install
npm run dev            # dev server; the API address comes from REACT_CARE_API_URL
npm test               # Vitest (unit + component tests)
npm run test:coverage  # same, with the 80 % gate on src/components/Patient
npm run lint:strict    # lint of the Klyniva code
npx tsc --noEmit       # type check
```

## Layout

```
src/components/<Name>/  one folder per feature: pages, dialogs, api.ts,
                        pure helpers (*Utils.ts) with their *.test.ts
src/components/Common/  page frame, help pop-up, view-as, shared widgets
src/Utils/              requests (retry, view-as headers), dates, settings, roles
src/hooks/              useAuthUser, useDateFormat, useApplySettings, ...
src/style/              Tailwind entry and the dark theme (theme.css)
public/locale/          translations; every Portuguese plural also needs _zero
```

## Conventions

- Logic that can be tested without a browser lives in `*Utils.ts` with tests;
  components stay thin. Browser journeys are in `../e2e/browser`.
- Every page passes `help="<topic>"` to `Page` and has `help__<topic>` in both
  locales (a test checks it).
- No literal UI text in components (lint:strict); no long dashes in texts.
- At runtime `/config.js` sets `REACT_CARE_API_URL`; `same-origin` means the
  address the page was opened from (what the cluster and compose use).

The app started as a fork of [CARE](https://github.com/ohcnetwork/care_fe)
(MIT, see `LICENSE`); unused CARE code is being removed as Klyniva grows.
