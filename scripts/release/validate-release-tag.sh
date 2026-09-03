#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 public-station-vX.Y.Z.N" >&2
  exit 64
fi

release_tag="$1"
if [[ ! "$release_tag" =~ ^public-station-v[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "invalid release tag: $release_tag" >&2
  echo "expected: public-station-vX.Y.Z.N" >&2
  exit 64
fi
