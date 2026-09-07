#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."
for tool in go bun; do command -v "$tool" >/dev/null; done
scripts/release/verify-public-source.sh
(cd web && bun install --frozen-lockfile && bun run typecheck && bun run test && bun run build)
GOWORK=off make test
printf '%s\n' 'Local frontend/backend checks passed. Release still requires secret/image scanning and database matrix evidence.'
