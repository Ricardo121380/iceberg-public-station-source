# Public-station release process

This repository is a modified source distribution of
[new-api](https://github.com/QuantumNous/new-api). It retains the upstream
project identity, copyright notices, `LICENSE`, `NOTICE`, and third-party
license material. The distribution is available under the GNU Affero General
Public License v3.0 or later; see [LICENSE](LICENSE).

## Local modifications

The public-station build adds invitation-gated LinuxDO registration, Turnstile
protection for OAuth state creation, Root-only invitation administration, and
the release controls documented below. It does not embed deployment secrets,
database snapshots, VPS addresses, or private operational evidence.

## Release contract

Production releases use annotated tags named `public-station-vX.Y.Z.N`. A
release is valid only when all of these refer to the same Git commit:

1. The private development repository's release tag.
2. The public source repository's identical tag.
3. The `source_commit` field in the generated release manifest.

The GitHub workflow builds only `linux/arm64`, using the native build platform
for the static frontend and Go cross-compilation while keeping the final runtime
image ARM64. It scans the candidate image before publication and then publishes
the private GHCR image by immutable digest. A deployment must use
`ghcr.io/<owner>/iceberg-public-station@sha256:<digest>`; it must never use
`latest` or a mutable tag. Task 13 supplies the Compose template that consumes
this digest.

The release workflow and image builder pin Go 1.26.6. The runtime is the
digest-pinned `gcr.io/distroless/static-debian13:nonroot` image, which supplies
the CA certificates and time-zone data required by the static Go binary without
a shell, package manager, or general-purpose OS tooling. It runs as numeric
UID/GID `10001:10001` with writable `/data`, and uses the static
`scripts/release/healthcheck.go` probe instead of installing a shell HTTP
client.

## Build and verification

Run the same release tag validation locally before creating a release:

```sh
scripts/release/validate-release-tag.sh public-station-v1.0.0.1
docker buildx build --platform linux/arm64 --load \
  --tag local/public-station:check .
scripts/release/verify-arm64-image.sh local/public-station:check
```

The GitHub workflow additionally runs backend and frontend tests, Gitleaks,
SBOM generation, and a blocking Trivy HIGH/CRITICAL vulnerability scan. The
public-source synchronisation job uses a write-enabled deploy key attached only
to the public source repository. Its private half is stored as
`PUBLIC_SOURCE_DEPLOY_KEY` only in the private repository's `release` GitHub
Environment. The runner writes it only to its temporary directory, verifies
GitHub's SSH host keys from the GitHub metadata endpoint, and uses an SSH remote
without embedding credentials in Git configuration. The Oracle host later
receives a separate read-only GHCR credential in its protected configuration
directory, never in this repository or image layer.

## Local execution alternative (2026-09-07)

The owner has paused GitHub Actions workflows to avoid reliance on hosted CI
quota. Existing workflow files and history remain available for rollback.
`scripts/release/verify-local.sh` runs the complete frontend and Go checks locally.
It does not replace required real-database compatibility tests, secret scans,
image runtime checks, vulnerability scans or production acceptance.

Releases may now use an offline, content-addressed ARM64 image built from an
exact committed source archive on a Linux ARM64 host. This is an authorised
alternative to the GHCR-only deployment path above. Preserve the same source
commit across the private and public annotated release tags. Record the image
content ID, source commit, scanner version/results, SBOM and recoverable image
archive; do not deploy a mutable image tag. Run production Compose with
`--pull never --no-deps`, back up the previous image configuration first and
restore it if startup/health checks fail. Verify the authenticated user-facing
path and that database/cache containers were not replaced.

The legacy GHCR forced-command deployer accepts only GHCR references and must
not be used for an offline image configuration. Follow the workspace operations
scripts for local-image deployment and rollback. No registry credential changes
are needed for offline deployment. GitHub retains source, tags and review history;
a future automated CI host requires its own explicit setup and verification.

## v1.0.0.19 offline release

The ARM64 binary is cross-compiled with Go 1.26.6 and embedded production web assets. The offline image preserves the exact v1.0.0.18 distroless runtime and healthcheck layers and replaces only /data/new-api. Its immutable local image ID is used in images.env; the private/public annotated source tags identify this source snapshot. No database schema changes are introduced. Telegram operations use a dedicated deployment secret, explicit private-chat actor binding and manual confirmation; no real credentials are shipped in source.
