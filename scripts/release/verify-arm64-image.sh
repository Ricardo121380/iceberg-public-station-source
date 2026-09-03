#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 image-reference" >&2
  exit 64
fi

image_reference="$1"
image_platform="$(docker image inspect --format '{{.Os}}/{{.Architecture}}' "$image_reference")"
if [[ "$image_platform" != 'linux/arm64' ]]; then
  echo "expected linux/arm64 image, got: $image_platform" >&2
  exit 1
fi

configured_user="$(docker image inspect --format '{{.Config.User}}' "$image_reference")"
if [[ "$configured_user" != '10001:10001' ]]; then
  echo "expected non-root image user 10001:10001, got: $configured_user" >&2
  exit 1
fi

healthcheck="$(docker image inspect --format '{{json .Config.Healthcheck}}' "$image_reference")"
if [[ "$healthcheck" == 'null' || "$healthcheck" == '<nil>' ]]; then
  echo 'image does not define a healthcheck' >&2
  exit 1
fi

runtime_uid="$(docker run --rm --platform linux/arm64 --entrypoint /usr/bin/id "$image_reference" -u)"
if [[ "$runtime_uid" != '10001' ]]; then
  echo "expected container runtime uid 10001, got: $runtime_uid" >&2
  exit 1
fi

container_name="public-station-health-${RANDOM}${RANDOM}"
cleanup() {
  docker rm --force "$container_name" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker run --detach --rm --name "$container_name" -p 127.0.0.1::3000 "$image_reference" >/dev/null
for _ in $(seq 1 45); do
  port="$(docker port "$container_name" 3000/tcp | awk -F: '{print $NF}')"
  health_status="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{end}}' "$container_name")"
  if [[ "$health_status" == 'healthy' ]] \
    && [[ -n "$port" ]] \
    && curl --fail --silent --show-error "http://127.0.0.1:${port}/api/status" | grep -Eq '"success":[[:space:]]*true'; then
    exit 0
  fi
  sleep 1
done

docker logs "$container_name" >&2 || true
echo 'container did not become healthy within 45 seconds' >&2
exit 1
