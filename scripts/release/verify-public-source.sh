#!/usr/bin/env bash
set -euo pipefail

required_files=(
  LICENSE
  NOTICE
  THIRD-PARTY-LICENSES.md
  README.md
  PUBLIC-STATION-RELEASE.md
)

for file in "${required_files[@]}"; do
  if [[ ! -s "$file" ]]; then
    echo "required public-source file is missing or empty: $file" >&2
    exit 1
  fi
done

grep -Fq 'GNU AFFERO GENERAL PUBLIC LICENSE' LICENSE
grep -Fq 'QuantumNous' NOTICE
grep -Fq 'new-api' PUBLIC-STATION-RELEASE.md
