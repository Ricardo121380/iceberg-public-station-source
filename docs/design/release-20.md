# Frontend polish release .20

Deploy the approved station-polish design based on .19 (705bcbb). User explicitly confirmed preserving full opening playback on every homepage visit, regardless of the system reduced-motion preference. The new design's absence of a skip button is retained; pointer/keyboard dismissal continues without intercepting the original event. The opening's SVG artwork and 3400ms duration are unchanged.

Only frontend code/assets, tests, this release note and VERSION differ from .19. No backend, database schema, Telegram configuration or legal-document activation changes are part of this release. Formatting adjustments are confined to changed files. Existing upstream lint findings match the production baseline.

Validation: updated motion-preference test failed before the correction; all 479 frontend tests and typecheck passed. Full root and relaykit Go regression passed. Release binary is cross-compiled with Go 1.26.6 for Linux ARM64, embedding the production frontend. Offline image retains the exact .19 runtime/healthcheck layers and updates the application binary plus accurate release metadata. Image runtime, secret/vulnerability scan and live browser acceptance are recorded in workspace release evidence.
