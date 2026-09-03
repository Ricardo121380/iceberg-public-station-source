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
