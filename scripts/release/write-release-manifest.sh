#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 6 ]]; then
  echo "usage: $0 release-tag image digest source-commit public-source-repository output-file" >&2
  exit 64
fi

release_tag="$1"
image="$2"
digest="$3"
source_commit="$4"
public_source_repository="$5"
output_file="$6"

mkdir -p "$(dirname "$output_file")"
jq -n \
  --arg release_tag "$release_tag" \
  --arg image "$image" \
  --arg digest "$digest" \
  --arg source_commit "$source_commit" \
  --arg public_source_repository "$public_source_repository" \
  --arg generated_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  '{
    release_tag: $release_tag,
    source_commit: $source_commit,
    public_source_repository: $public_source_repository,
    image: ($image + "@" + $digest),
    platform: "linux/arm64",
    generated_at: $generated_at
  }' > "$output_file"
