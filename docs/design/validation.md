# Iceberg frontend design validation — 2026-09-05

Scope: default homepage, sign-in and sign-up. Branch `codex/iceberg-frontend`, based on production commit `a5cb953`. Custom homepage rendering and the shared AuthLayout used by other authentication screens are unchanged. Existing OAuth, invitation, Turnstile, consent and redirect components are reused.

## Checks

- `bun run typecheck`: PASS.
- Scoped `oxlint -c .oxlintrc.json`: PASS for all changed/new TSX files.
- Scoped oxfmt readback with protected copyright headers retained: PASS.
- Existing locale values preserved byte-equivalently as JSON values; new English and simplified Chinese content, English fallback for other supported languages. Existing locale formatting and protected attribution keys preserved.
- `bun run test src/features/auth/components/__tests__/linuxdo-invite-login.test.tsx src/features/auth/hooks/__tests__/use-oauth-login.test.tsx src/features/auth/lib/__tests__/oauth-callback-mode.test.ts src/features/home/components/__tests__/station-connection.test.tsx`: 4 files, 20 tests PASS.
- `bun run build`: PASS. Existing application's aggregate output is approximately 58 MB before compression; this work does not optimize unrelated route/vendor bundles. Added photography approximately 410 KiB and CJK display subset approximately 20 KiB. No new dependencies.
- Impeccable detector on changed page components: `[]`.
- CodeGraph sync and real `StationHome` symbol query: PASS.
- Production-build browser inspection: homepage, sign-in, sign-up at 1440px and 390px; registration at 320px; dark English homepage at 1440px. No horizontal overflow.
- Browser interactions: in-page connection link, API address copy, client tabs, FAQ.

## Preview

Run in `web/`:

```sh
VITE_REACT_APP_SERVER_URL=https://iceberg.tiktok.vip bun run preview --host 127.0.0.1 --port 4174
```

This serves the local production build; `/api` reads proxy to the station. Do not use this preview as an OAuth end-to-end acceptance environment. It is not deployed. API credentials are never embedded; the cURL example contains only YOUR_API_KEY.

Turnstile returns a connection failure on localhost. Registration correctly remains disabled and exposes retry. Successful registration and OAuth callback must be accepted on the trusted site domain before a production release; no real invitation was consumed and no OAuth login was submitted during this design validation.

Screenshots: project workspace `.impeccable/review/{desktop,mobile,login-desktop,login-mobile,register-desktop,register-mobile,register-320,dark-english}.png`.

Independent read-only finish review: **ship**. All eight screenshot captures reviewed; no material visual, UX or functional regression found within source/screenshot scope. The reviewer did not independently rerun tests or perform browser interaction. Successful Turnstile/OAuth registration remains unverified locally.

# Revision 2 — logo-led Liquid Glass

The user asked to match the original illustrated icon. Replaced photo/serif design with the original unchanged mascot, sans-serif titles, icy blues, rose actions and restrained glass surfaces. Added shared `StationArt`; all existing authentication and connection logic retained.

- 20 tests across the same four suites: PASS.
- Typecheck, scoped lint, production build: PASS.
- Source detector: `[]` (the tool's default comp gate does not apply to this code-led reference-driven revision).
- CodeGraph sync and `StationArt` query: PASS.
- Production build inspected at 1440px / 390px, sign-up at 320px, plus dark homepage. No document overflow.
- Independent review compared the original icon and all six captures: **ship**, no material design/UX findings within supplied scope.
- Glass uses CSS blur/saturation, specular inset edges and soft offset shadows, not physical optical refraction. Reduced-transparency and unsupported-blur fallbacks preserve readable opaque surfaces. Reduced-motion disables entrance and action movement.
- Header switches to the existing compact navigation at 1000px so longer language labels fit intermediate widths.
- The prior third-party photo and serif font subset are removed. Original station logo is unchanged; new Noto Sans SC title subset has its OFL record.
- Screenshots: workspace `.impeccable/review-v2/`.
- Still local preview only; production Turnstile/OAuth end-to-end acceptance is not claimed.
