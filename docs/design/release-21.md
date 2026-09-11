# Iceberg console release .21

Approved console-icecabin changes based on .20: default Iceberg preset, console chrome, chart palettes and empty states, and animated overview/wallet numbers. Existing saved theme preferences remain respected. Homepage opening is unchanged.

Fixed an observed interruption defect before release: a mid-animation value update now continues from the current visible value, not the last settled value. Added a regression that fails before and passes after the fix. Final quota formatting and backend calculations are unchanged.

No backend/schema/dependency changes. Release uses Go1.26.6 Linux ARM64 binary with embedded production frontend, atop the exact .20 runtime image. Runtime checks, scans, backup and browser acceptance are recorded in workspace evidence. Legal settings, account permissions and Telegram/monitor configuration are outside this release.
