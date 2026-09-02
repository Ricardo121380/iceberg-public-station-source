# Iceberg Public Station Upstream Baseline

This repository starts from the corresponding source of NewAPI
`v1.0.0-rc.30`.

| Field | Value |
|---|---|
| Upstream repository | `https://github.com/QuantumNous/new-api.git` |
| Upstream tag | `v1.0.0-rc.30` |
| Upstream commit | `27ff6a8767e728f879d52770c273d4f73214a430` |
| Imported local commit | `6a9ccafea198b78c9454ec9191ea80bf29c03957` |
| Source archive | `https://codeload.github.com/QuantumNous/new-api/tar.gz/refs/tags/v1.0.0-rc.30` |
| License | AGPL-3.0; preserve `LICENSE`, `NOTICE`, and `THIRD-PARTY-LICENSES.md` |

The initial source was imported from the official GitHub tag archive after the
local Git protocol transfer did not complete. The immutable upstream tag and
commit above are the source-of-truth provenance; the annotated local tag
`upstream-v1.0.0-rc.30` records the same mapping.

## Update procedure

1. Fetch and verify the intended upstream tag from `upstream`.
2. Create `sync/upstream-<tag>` from the current production baseline.
3. Produce a source diff, review database/authentication/relay changes, and
   run the full test matrix before merging.
4. Record the new upstream tag, SHA, local import or merge commit, image
   digest, and public source tag in the release mapping.

## Secret-scan baseline

`.gitleaksignore` contains ten exact findings from the imported upstream commit
`6a9ccafea198b78c9454ec9191ea80bf29c03957`. They are audited false positives:
test-only token masking fixtures, a dashboard-version marker, Jimeng model
identifiers, PEM-envelope construction code, and an API-key masking example.
The entries are fingerprinted by commit, file, rule, and line. They do not
ignore future occurrences of the same patterns.

Do not replace or remove NewAPI/QuantumNous attribution, license, notices, or
the source-offer obligations of a public network deployment.
