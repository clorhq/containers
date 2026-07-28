# AGENTS.md

Container images for software development, published to `ghcr.io/clorhq`.

## This repository is public

`clorhq/containers` is a public repository, and the images it builds are
public on GHCR. Everything committed here — including this file, commit
messages, and anything baked into an image layer — is world-readable forever.

Never add credentials, tokens, private URLs, customer names, internal
hostnames, or anything else that is not already public. Secrets belong in
GitHub Actions secrets, referenced as `${{ secrets.* }}` in the workflow, and
must never be passed as a `--build-arg` or an `ARG`: build arguments are
recorded in the image metadata and readable by anyone who pulls the image.

## Reporting image refs

When asked for image IDs, digests, or refs, give the **multi-arch index
digest** — that is what `:latest` points at and what should be pinned, since
it resolves to the right architecture on any host. Report per-architecture
manifest digests only when a single platform was explicitly asked for, and
label them as such. Always give full untruncated `sha256:` values.

```bash
docker buildx imagetools inspect ghcr.io/clorhq/software-development-base:latest
```

Without a Docker daemon, the registry API returns the same index digest in the
`Docker-Content-Digest` response header:

```bash
REPOSITORY="clorhq/software-development-base"
TOKEN="$(curl --fail --silent --show-error \
    "https://ghcr.io/token?scope=repository:${REPOSITORY}:pull&service=ghcr.io" |
    jq --raw-output .token)"
curl --fail --silent --show-error --head \
    --header "Authorization: Bearer ${TOKEN}" \
    --header "Accept: application/vnd.oci.image.index.v1+json" \
    "https://ghcr.io/v2/${REPOSITORY}/manifests/latest"
```

## Layout

| Path | Purpose |
| --- | --- |
| `images/<name>/Dockerfile` | One image per directory: `base`, plus `rust`, `python`, `zig`, `ruby`, `desktop` |
| `images/<name>/tests/` | Test assets copied only into the `test` stage, never into `production` |
| `installers/<tool>` | In-container install script; takes the version as `$1` |
| `versions/<tool>` | Host-side script printing the latest upstream version of `<tool>` |
| `versions/lib` | Shared `github_latest` / `github_latest_tag` / `npm_latest` helpers |
| `scripts/build` | Native-arch build of every image into the local Docker store |
| `scripts/push` | Multi-arch build and push to `ghcr.io/clorhq` from a laptop |
| `scripts/update` | Rewrites every `ARG *_VERSION` default to the latest upstream version |
| `.github/workflows/build.yml` | CI: per-arch native builds pushed by digest, then merged into one tag |

`images/base` is the root image. Every variant is `FROM ${BASE_IMAGE}` and
inherits the agents, OS tooling, and base toolchains.

## Version pinning

Every downloaded tool is pinned to an explicit `ARG <TOOL>_VERSION="x.y.z"`
default in the Dockerfile. Nothing installs "latest" at build time, so a
rebuild of an unchanged tree produces the same tool set.

The three pieces are wired together by name:

```
ARG GOLANGCI_LINT_VERSION="2.12.2"   # in the Dockerfile
versions/golangci-lint               # prints the latest upstream version
installers/golangci-lint 2.12.2      # installs that version in the image
```

`scripts/update` derives the script name from the ARG name (strip `_VERSION`,
lowercase, `_` → `-`). **Adding a new pinned tool means adding all three**, or
`scripts/update` will skip it with a warning and the pin will silently rot.

Each `ARG` carries a `# Version source: <url>` comment naming where the version
comes from; keep it next to the pin.

Run `./scripts/update` to bump everything, then review the diff — a bump is a
real change to the image and should be committed on its own.

## Building and testing

```bash
./scripts/build        # all images, native arch, into the local Docker store
```

Local builds pin the `default` (docker-driver) builder on purpose: a
docker-container builder cannot see the host image store, so a variant's
`FROM software-development-base:latest` would fall through to Docker Hub.

Most tests live in the Dockerfile's `test` stage, which asserts against the
finished image — installed binaries on `PATH` for both `root` and `user`,
environment and profile behavior, `bash -n`, `node --check`, `shellcheck`, and
`shfmt -d -i 4 -ci` over every shipped script, then runs the scripts under
`images/base/tests/`. Build the stage directly to run them:

```bash
docker buildx build --target test -f images/base/Dockerfile .
```

The desktop image adds `images/desktop/tests/runtime`, which needs a real
container and so runs after the push in CI rather than inside a build stage.

When you add or change a shipped script, add it to the `bash -n`, `shellcheck`,
and `shfmt` lists in the `test` stage.

## CI

`build.yml` runs on pushes to `main` and on `workflow_dispatch`. Each
architecture builds on its own native runner (`ubuntu-latest` for amd64,
`ubuntu-24.04-arm` for arm64) and pushes **by digest** — no QEMU anywhere. The
`merge-base` / `merge-variants` jobs then stitch the per-arch digests into a
single multi-arch `:latest` tag with `docker buildx imagetools create`.

Variants build `FROM ghcr.io/clorhq/software-development-base:latest`, so they
depend on `merge-base` having published the new base first.

`[skip ci]` in a commit message skips the build entirely — the tree can be
ahead of what `:latest` contains.

## Shell style

All scripts are bash and follow one house style:

- `#!/bin/bash` with `set -o errexit`, `set -o nounset`, `set -o pipefail`
- `UPPERCASE` variables, always brace-quoted: `"${VERSION}"`
- Long-form flags everywhere: `--parents`, `--recursive --force`,
  `curl --fail --silent --show-error --location`
- 4-space indent, `shfmt -i 4 -ci` clean, `shellcheck` clean
- Comments explain *why* a non-obvious thing is done, not what the line does

Scripts under `scripts/` and `versions/` run **on the host, including macOS
with stock bash 3.2 and BSD coreutils** — no associative arrays, no `mapfile`,
no GNU-only flags, and in-place edits go through a temp file plus `mv`.
Scripts under `installers/` and inside the images run on Debian with GNU tools,
where the long GNU flags are fine.

Installers cache their download at `/downloads/<archive>` and skip the fetch
when it already exists. `/downloads` is a BuildKit cache mount
(`--mount=type=cache,target=/downloads,sharing=locked`), shared across builds;
user-stage installs use the separate `id=base-user-downloads` cache owned by
uid 1000. The installers themselves are bind-mounted read-only from
`installers/` rather than copied into a layer.

## Conventions worth keeping

- Multi-arch: every installer switches on `dpkg --print-architecture` and fails
  loudly on anything but `amd64` / `arm64`.
- The image drops to `USER user` for user-scoped installs and returns to
  `USER root` afterwards; keep that pairing balanced.
- `test` and `production` are both `FROM base`, so test assets never reach the
  published image.
- `CLOR_BUILD_NONCE` deliberately busts the cache for the layer that installs
  the unpinned latest Clor build.
- Remote assets fetched by URL are pinned by checksum (see `CLOR_LOGO_SHA256`).

## Known drift

`README.md` lists `software-development-go` and `software-development-typescript`
images that do not exist — there is no `images/go` or `images/typescript`, and
neither is in the CI matrix. The Go and TypeScript toolchains ship in `base`.
