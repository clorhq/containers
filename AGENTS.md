# AGENTS.md

Container images for software development, published to `ghcr.io/clorhq`.

## Scope: this repository only

Work only in `clorhq/containers`. Do not clone, edit, branch, or open pull
requests against another repository unless explicitly told to in that request.

Plenty of what a space does is *not* built here — the Clor CLI and its web
terminal, the space daemon, the webtui's embedded fonts and pane commands, the
generated `~/.config` files for tools like gh-dash. A problem showing up inside
a space is not evidence that its fix belongs here. Before changing anything,
confirm the thing you are about to edit is actually produced by this tree
(`git grep` it, check which `images/*/Dockerfile` installs it).

When the fix lands outside this repo, **stop and write it up** rather than
going looking for the other repository: name the file or setting, the current
value, the replacement, and the evidence. The user routes it from there.

Two worked examples:

- **gh-dash launched unauthenticated.** The root cause *was* here — the
  Dockerfile left a `go install` copy in `~/go/bin` shadowing the wrapper — so
  it was fixed here.
- **Nerd Font glyphs render as tofu.** The `clor` binary embeds its own
  `JetBrainsMono-*.woff2` and serves it to the browser, so it overrides any
  font installed in the image. Nothing in this repo can fix it, and installing
  fonts here would only have affected the desktop image's X terminal.

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
| `images/<name>/Dockerfile` | One image per directory |
| `images/<name>/tests/` | Test assets copied only into the `test` stage, never into `production` |
| `installers/<tool>` | In-container install script; takes the version as `$1` |
| `installers/verified-github-asset` | Resolves and checksum-verifies a per-arch GitHub release asset |
| `installers/cargo-binstall` | Bootstraps cargo-binstall, used for the pinned Rust-native tools |
| `versions/<tool>` | Host-side script printing the latest upstream version of `<tool>` |
| `versions/lib` | Shared `github_latest` / `github_latest_tag` / `npm_latest` helpers |
| `scripts/build` | Native-arch build of every image into the local Docker store |
| `scripts/push` | Multi-arch build and push to `ghcr.io/clorhq` from a laptop |
| `scripts/update` | Rewrites every `ARG *_VERSION` default to the latest upstream version |
| `.github/workflows/build.yml` | CI: per-arch native builds pushed by digest, then merged into one tag |

`images/base` is the root image. Every other image is `FROM ${BASE_IMAGE}` and
inherits the agents, OS tooling, and base toolchains.

**Only four images are actually built and published:** `base`, `desktop`,
`devops`, and `data`. See "Known drift" for the rest.

## Version pinning

Every downloaded tool is pinned to an explicit `ARG <TOOL>_VERSION="x.y.z"`
default in the Dockerfile. Nothing installs "latest" at build time, so a
rebuild of an unchanged tree produces the same tool set. Each `ARG` carries a
`# Version source: <url>` comment naming where the version comes from; keep it
next to the pin.

Where all three parts exist, they are wired together by name:

```
ARG GOLANGCI_LINT_VERSION="2.12.2"   # in the Dockerfile
versions/golangci-lint               # prints the latest upstream version
installers/golangci-lint 2.12.2      # installs that version in the image
```

`scripts/update` derives the script name from the ARG name (strip `_VERSION`,
lowercase, `_` → `-`) and **silently skips any ARG with no matching
`versions/` script**, printing a `skip …` line to stderr. Most pins are in
that state today (see "Known drift"), so read `scripts/update`'s stderr — a
clean-looking run is not the same as a complete one.

Run `./scripts/update` to bump what it can, then review the diff — a bump is a
real change to the image and should be committed on its own.

Fetching a release asset directly (no installer of its own) goes through
`installers/verified-github-asset OWNER/REPO TAG AMD64_ASSET ARM64_ASSET`,
which picks the asset for the build architecture, verifies the digest GitHub
publishes for it, and echoes the cached path. Prefer it over a bare `curl`.

## Building and testing

```bash
./scripts/build        # all images, native arch, into the local Docker store
```

Local builds pin the `default` (docker-driver) builder on purpose: a
docker-container builder cannot see the host image store, so a derived image's
`FROM software-development-base:latest` would fall through to Docker Hub.

Most tests live in the Dockerfile's `test` stage, which asserts against the
finished image — installed binaries on `PATH` for both `root` and `user`,
environment and profile behavior, `bash -n`, `node --check`, `shellcheck`, and
`shfmt -d -i 4 -ci` over every shipped script, then runs the scripts under
`images/base/tests/`. Build the stage directly to run them:

```bash
docker buildx build --target test -f images/base/Dockerfile .
```

Only `base`, `data`, `desktop`, and `devops` have a `test` stage. The desktop
image additionally has `images/desktop/tests/runtime`, which needs a real
container and so runs after the push in CI rather than inside a build stage.

When you add or change a shipped script, add it to the `bash -n`, `shellcheck`,
and `shfmt` lists in the `test` stage.

### Assert what a command resolves to, not just that it exists

The space user's `PATH` is
`~/.local/bin:~/go/bin:~/.npm-global/bin:…:/usr/local/bin`, so **anything left
in `~/go/bin` or `~/.local/bin` shadows a wrapper of the same name in
`/usr/local/bin`**. This has bitten `gh-dash`: its `go install` copy stayed
behind and won over the wrapper that injects `GH_TOKEN`, so it launched
unauthenticated while `command -v gh-dash` still succeeded.

For any tool that ships behind a wrapper, assert the full path:

```dockerfile
&& test "$(command -v gh-dash)" = /usr/local/bin/gh-dash \
```

and delete the build-time copy once it has been installed to its final home.

## CI

`build.yml` runs on pushes to `main` and on `workflow_dispatch`. Each
architecture builds on its own native runner (`ubuntu-latest` for amd64,
`ubuntu-24.04-arm` for arm64) and pushes **by digest** — no QEMU anywhere. The
`merge-*` jobs then stitch the per-arch digests into a single multi-arch
`:latest` tag with `docker buildx imagetools create`.

Jobs: `base` → `merge-base` → `variants` (`desktop`, `devops`) →
`merge-variants`, plus `data` → `merge-data`. Everything downstream builds
`FROM ghcr.io/clorhq/software-development-base:latest`, so it depends on
`merge-base` having published the new base first. **`base` and `merge-base`
finish well before the rest** — if you only need the base digest, watch those
two jobs and ignore the others.

`workflow_dispatch` also accepts `data_only` (rebuild just Data against a given
`base_ref`) and `merge_data_only` (assemble Data from supplied per-arch
digests), which is why the concurrency group varies by mode.

`[skip ci]` in a commit message skips the build entirely — the tree can be
ahead of what `:latest` contains. Use it for docs-only commits so they don't
kick off a full rebuild.

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
  the unpinned latest Clor build. It is the one intentional exception to the
  everything-is-pinned rule, and the build asserts it is non-empty.
- Remote assets fetched by URL are pinned by checksum, either through
  `installers/verified-github-asset` or an explicit `sha256sum --check`.
- Patches to code-server's own files (the GitHub auth provider, `product.json`)
  sit next to a version assertion or checksum so an upstream bump fails loudly
  instead of silently no-opping.

### code-server GitHub authentication

`images/base/vscode/github-authentication/` replaces the built-in provider's
entrypoint so sessions come from `clor github auth` instead of a device-code
flow. Because spaces authenticate with no human to click **Allow**, extensions
must also appear in `product.json`'s `trustedExtensionAuthAccess` — Code
returns nothing to a silent `getSession` for an untrusted extension, and the
extension simply renders as signed out. If a newly added GitHub extension shows
"Sign in", that list is the first place to look.

## Known drift

- **`images/rust`, `images/python`, `images/zig`, `images/ruby` are no longer
  built.** They have Dockerfiles but are absent from the CI matrix and have no
  `test` stage. Their `:latest` tags on GHCR are stale leftovers from before
  the catalog was narrowed. Either restore them to the matrix or delete them —
  right now they look supported and are not.
- **`README.md` lists `software-development-go` and
  `software-development-typescript`**, which have never existed as directories,
  and still advertises the four unbuilt language images above. Those toolchains
  ship in `base`.
- **Most `ARG *_VERSION` pins have no `versions/` script** — roughly eighty,
  including everything installed via cargo-binstall and
  `verified-github-asset` (yazi, delta, eza, mise, ast-grep, hurl, the cloud
  CLIs, the Data image's Python packages, the code-server extensions).
  `scripts/update` cannot bump any of them, so they only move when edited by
  hand.
