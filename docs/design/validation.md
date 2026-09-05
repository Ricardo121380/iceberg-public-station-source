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

# Revision 3 — derive the interface from the icon

User explicitly rejected the enlarged logo composition. Removed `StationArt` and introduced a flat SVG background of ice facets and sweeping wake curves, functional `StationLaunchpad`, outlined onboarding markers and shared full-scene auth styling. The station mark remains only at normal branding sizes. User's “轻舟已撞大冰山” caption is preserved.

- New launchpad tests cover clipboard success/failure, manual selection, signed-out login redirect to /keys, and signed-in direct /keys link.
- 24 tests across five auth/connection/launchpad files: PASS.
- Typecheck, scoped lint, formatter and production build: PASS.
- Detector `[]`; no large brand illustration remains in page content.
- Screenshots at 1440, 390, 320 registration, 820 English, plus dark theme. No document overflow; measured large raster images (>70px wide) = 0.
- Decorative flat SVG uses explicit geometry rather than a generated raster or logo trace. No new dependencies/assets required.
- Existing Turnstile fails on localhost and correctly keeps registration disabled with retry. Live verification still requires the trusted deployment domain; no production change was made.

Revision 3 final review: **ship**. Independent reviewer opened all eight final captures and inspected source; no material presentation issues in scope. Functional browser verification: public address copied successfully; key link reached `/sign-in?redirect=%2Fkeys`. Reduced-preference behavior is verified on the new launchpad, not inferred from v2. No live OAuth success is asserted.

# First-visit opening — local preview

User requested a simple shrimp-boat collision using the existing icon. Five transparent PNG crops retain the original pixels; the reproducible masks are in `extract-opening-assets.py`. A 2.4-second CSS sequence sails the boat in, recoils on impact, separates the ice, throws two chips and a splash, then fades into the unchanged homepage.

- Mounted only on the default homepage; no auth, registration, navigation or backend logic changed.
- One playback per browser storage profile, recorded under `iceberg-opening-v1-seen`. Pointer/keyboard input dismisses without preventing the original event. A visible skip button is provided.
- Reduced-motion preference, unavailable storage, failed artwork, and artwork taking more than 1.5 seconds all skip the opening.
- Eight component behavior tests passed. Typecheck and production build passed; scoped lint has no errors and four `prefer-add-event-listener` warnings on private preload Image handlers.
- Desktop and 390px mobile keyframes visually inspected. Browser-frame GIF and screenshots are in workspace `.impeccable/opening-review/`.
- Live local-browser DOM observation measured a 2404 ms playback; reload did not play again, and reduced-motion produced no opening or seen flag.
- Clicking the real key-creation link during playback dismissed the opening and reached `/sign-in?redirect=%2Fkeys` in the same action.
- This addition has not been deployed; production remains the approved revision 3 release.

## Opening revision 2 — depth and iceberg scale

User requested motion from inside the scene toward the viewer and a larger iceberg. The boat now stays centered, starts behind the joined ice at 0.18 scale, breaks through the opening, and grows to 1.28 scale in the foreground. The ice layers are 64% and 52% of the stage width (formerly 37% and 31%); the base boat is 40% (formerly 51%). The original side-view artwork is preserved; the depth comes from occlusion, scale, and vertical perspective rather than a newly drawn front view.

- Only animation CSS changed in this revision. Duration, first-visit storage, skip handling, and operating flows remain unchanged.
- Production build passed. Desktop and 390px mobile start/impact/exit frames inspected together; mountain peaks remain clearly taller than the boat.
- Ego screenshot calls repeatedly timed out, so the final visual capture used an isolated Playwright browser. The browser and temporary recording hooks were closed after capture.
- Final animation preview: workspace `output/playwright/opening-v2/虾船破冰-由内向外.gif`. Local preview only.

## Opening revision 3 — break through the lens

User requested an abrupt close-up after breaking the iceberg. The impact now adds a brief, small stage shake, sends the ice and chips outward, and holds the emerging boat for one beat before accelerating from 0.72 to 10 times its base size. The final boat extends beyond the viewport and immediately clears into the homepage; no enlarged static logo remains. The caption appears only before the impact.

- CSS-only revision; the existing 2.4-second lifecycle, first-visit and reduced-motion handling, skip behavior, and application operations are unchanged.
- Desktop and 390px mobile collision, approach, and close-up frames visually inspected. Production build and diff whitespace checks passed.
- Current preview: workspace `output/playwright/opening-v3/虾船破冰-突脸版.gif`. Not deployed.
