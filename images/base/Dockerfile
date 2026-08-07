ARG OS_IMAGE="debian:bookworm-slim"
# Version source: https://github.com/tmux/tmux/releases/latest
ARG TMUX_VERSION="3.7b"

FROM ${OS_IMAGE} AS tmux-builder

RUN apt-get update && apt-get install --yes \
  bison \
  build-essential \
  ca-certificates \
  curl \
  jq \
  libevent-dev \
  libncurses-dev \
  pkg-config \
  && rm --recursive --force /var/lib/apt/lists/*

ARG TMUX_VERSION
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/tmux "${TMUX_VERSION}"

FROM ${OS_IMAGE} AS base

# Terminal and UTF-8 defaults.
ENV TERM="xterm-256color"
ENV COLORTERM="truecolor"
ENV LANG="C.UTF-8"
ENV LC_ALL="C.UTF-8"

# Install system tools. No browser is installed from apt: the image already
# carries Playwright's own Chromium and Firefox builds, and the wrapper on
# PATH at /usr/local/bin/chromium runs the Playwright Chromium. The fonts came
# in as dependencies of those packages, and a browser with no fonts renders
# every screenshot as boxes, so they are now requested by name.
RUN apt-get update && apt-get install --yes \
  bubblewrap \
  ca-certificates \
  curl \
  unzip \
  zip \
  git \
  git-lfs \
  gnupg \
  python3 \
  python3-pip \
  python3-venv \
  python3-dev \
  python3-setuptools \
  python3-wheel \
  pipx \
  make \
  cmake \
  ninja-build \
  pkg-config \
  ripgrep \
  g++ \
  build-essential \
  xz-utils \
  gosu \
  rsync \
  procps \
  vim \
  nano \
  less \
  net-tools \
  jq \
  pwgen \
  wget \
  iputils-ping \
  iproute2 \
  tmux \
  screen \
  ncurses-term \
  netcat-openbsd \
  traceroute \
  mtr \
  fzf \
  minisign \
  file \
  lsof \
  strace \
  gdb \
  psmisc \
  bind9-dnsutils \
  whois \
  tree \
  ncdu \
  bat \
  fd-find \
  htop \
  btop \
  sysstat \
  sqlite3 \
  openssl \
  openssh-client \
  sshpass \
  zstd \
  man-db \
  manpages \
  bsdmainutils \
  tcpdump \
  socat \
  xxd \
  pv \
  p7zip-full \
  parallel \
  alsa-utils \
  libasound2-plugins \
  pulseaudio-utils \
  mpg123 \
  ffmpeg \
  sox \
  imagemagick \
  exiftool \
  poppler-utils \
  rclone \
  shellcheck \
  shfmt \
  xvfb \
  fonts-liberation \
  fonts-dejavu-core \
  fonts-noto-color-emoji \
  && if [ "$(dpkg --print-architecture)" = "amd64" ]; then apt-get install --yes ltrace; fi \
  && ln --symbolic /usr/bin/fdfind /usr/local/bin/fd \
  && ln --symbolic /usr/bin/batcat /usr/local/bin/bat \
  && rm --recursive --force /var/lib/apt/lists/*

# Install complete upstream distributions for tools where Debian Bookworm is
# substantially behind. The shared installer verifies GitHub's published
# digest for the exact architecture-specific release asset.
ARG NEOVIM_VERSION="0.12.4"
ARG PANDOC_VERSION="3.10.1"
ARG LNAV_VERSION="0.14.0"
ARG DUCKDB_VERSION="1.5.5"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  nvim_archive="$("/installers/verified-github-asset" neovim/neovim "v${NEOVIM_VERSION}" \
    nvim-linux-x86_64.tar.gz nvim-linux-arm64.tar.gz)" \
  && mkdir --parents "/opt/neovim/${NEOVIM_VERSION}" \
  && tar --extract --gzip --strip-components=1 \
    --directory="/opt/neovim/${NEOVIM_VERSION}" --file="${nvim_archive}" \
  && ln --symbolic "/opt/neovim/${NEOVIM_VERSION}/bin/nvim" /usr/local/bin/nvim \
  && pandoc_archive="$("/installers/verified-github-asset" jgm/pandoc "${PANDOC_VERSION}" \
    "pandoc-${PANDOC_VERSION}-linux-amd64.tar.gz" \
    "pandoc-${PANDOC_VERSION}-linux-arm64.tar.gz")" \
  && mkdir --parents "/opt/pandoc/${PANDOC_VERSION}" \
  && tar --extract --gzip --strip-components=1 \
    --directory="/opt/pandoc/${PANDOC_VERSION}" --file="${pandoc_archive}" \
  && for tool in "/opt/pandoc/${PANDOC_VERSION}/bin/"*; do \
    ln --symbolic "${tool}" "/usr/local/bin/$(basename "${tool}")"; \
  done \
  && lnav_archive="$("/installers/verified-github-asset" tstack/lnav "v${LNAV_VERSION}" \
    "lnav-${LNAV_VERSION}-linux-musl-x86_64.zip" \
    "lnav-${LNAV_VERSION}-linux-musl-arm64.zip")" \
  && unzip -q "${lnav_archive}" -d /tmp/lnav \
  && install --mode=0755 "/tmp/lnav/lnav-${LNAV_VERSION}/lnav" /usr/local/bin/lnav \
  && duckdb_archive="$("/installers/verified-github-asset" duckdb/duckdb "v${DUCKDB_VERSION}" \
    duckdb_cli-linux-amd64.zip duckdb_cli-linux-arm64.zip)" \
  && unzip -q "${duckdb_archive}" -d /tmp/duckdb \
  && install --mode=0755 /tmp/duckdb/duckdb /usr/local/bin/duckdb \
  && rm --recursive --force /tmp/lnav /tmp/duckdb

# Installer downloads use a shared BuildKit cache.

# Install GitHub CLI.
# Version source: https://github.com/cli/cli/releases/latest
ARG GITHUB_VERSION="2.97.0"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/github "${GITHUB_VERSION}"

# Install Node.js
# Version source: https://nodejs.org/dist/index.json (latest entry with an LTS name)
ARG NODE_VERSION="24.18.1"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/node "${NODE_VERSION}"

# Install additional JavaScript runtimes and tooling system-wide.
# Version source: https://api.github.com/repos/oven-sh/bun/releases/latest
ARG BUN_VERSION="1.3.14"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/bun "${BUN_VERSION}"

# Version source: https://api.github.com/repos/denoland/deno/releases/latest
ARG DENO_VERSION="2.9.4"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/deno "${DENO_VERSION}"

# Tailwind CSS standalone CLI.
ARG TAILWINDCSS_VERSION="4.3.3"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/tailwindcss "${TAILWINDCSS_VERSION}"

# Install Go system-wide.
# Version source: https://go.dev/VERSION?m=text
ARG GO_VERSION="1.26.5"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/go "${GO_VERSION}"

# Install golangci-lint system-wide.
ARG GOLANGCI_LINT_VERSION="2.12.2"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/golangci-lint "${GOLANGCI_LINT_VERSION}"

# Install uv system-wide.
ARG UV_VERSION="0.12.1"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/uv "${UV_VERSION}"

# Install terminal tools.
# Version source: https://github.com/jesseduffield/lazygit/releases/latest
ARG LAZYGIT_VERSION="0.63.1"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/lazygit "${LAZYGIT_VERSION}"

# Version source: https://github.com/charmbracelet/glow/releases/latest
ARG GLOW_VERSION="2.1.2"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/glow "${GLOW_VERSION}"

# Install pinned terminal navigation, editing, automation, networking, quality,
# data, and content tools. cargo-binstall obtains the projects' upstream
# prebuilt releases and verifies their checksums; --no-confirm keeps image
# builds noninteractive and exact package versions prevent silent drift.
ARG CARGO_BINSTALL_VERSION="1.21.1"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/cargo-binstall "${CARGO_BINSTALL_VERSION}"

ARG YAZI_VERSION="26.5.6"
ARG DELTA_VERSION="0.19.2"
ARG DIFFTASTIC_VERSION="0.69.0"
ARG EZA_VERSION="0.23.5"
ARG ZOXIDE_VERSION="0.10.0"
ARG JUST_VERSION="1.57.0"
ARG WATCHEXEC_VERSION="2.5.1"
ARG MISE_VERSION="2026.7.13"
ARG AST_GREP_VERSION="0.45.0"
ARG XH_VERSION="0.26.1"
ARG HURL_VERSION="8.0.1"
ARG OHA_VERSION="1.15.0"
ARG WEBSOCAT_VERSION="1.14.1"
ARG TYPOS_VERSION="1.48.0"
ARG LYCHEE_VERSION="0.24.2"
ARG TYPST_VERSION="0.15.1"
ARG PRESENTERM_VERSION="0.16.1"
RUN cargo-binstall --no-confirm --force \
  "git-delta@${DELTA_VERSION}" \
  "difftastic@${DIFFTASTIC_VERSION}" \
  "eza@${EZA_VERSION}" \
  "zoxide@${ZOXIDE_VERSION}" \
  "just@${JUST_VERSION}" \
  "watchexec-cli@${WATCHEXEC_VERSION}" \
  "mise@${MISE_VERSION}" \
  "ast-grep@${AST_GREP_VERSION}" \
  "xh@${XH_VERSION}" \
  "websocat@${WEBSOCAT_VERSION}" \
  "typos-cli@${TYPOS_VERSION}" \
  "lychee@${LYCHEE_VERSION}" \
  "typst-cli@${TYPST_VERSION}" \
  "presenterm@${PRESENTERM_VERSION}" \
  && if ! command -v ast-grep >/dev/null 2>&1; then \
    ln --symbolic "$(command -v sg)" /usr/local/bin/ast-grep; \
  fi

# Install Yazi's pinned architecture-specific upstream archive directly into
# the system path. This keeps both public commands in the production
# filesystem independently of cargo-binstall's package metadata and bin path.
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  YAZI_ARCHIVE="$("/installers/verified-github-asset" sxyazi/yazi "v${YAZI_VERSION}" \
    yazi-x86_64-unknown-linux-musl.zip \
    yazi-aarch64-unknown-linux-musl.zip)" \
  && mkdir --parents /tmp/yazi \
  && unzip -q "${YAZI_ARCHIVE}" -d /tmp/yazi \
  && YAZI_DIRECTORY="$(find /tmp/yazi \
    -mindepth 1 -maxdepth 1 -type d -name 'yazi-*' -print -quit)" \
  && test -n "${YAZI_DIRECTORY}" \
  && install --mode=0755 \
    "${YAZI_DIRECTORY}/yazi" \
    "${YAZI_DIRECTORY}/ya" \
    /usr/local/bin/ \
  && rm --recursive --force /tmp/yazi

# Hurl's and oha's cargo-binstall metadata do not match the layout of their
# current release assets. Install the exact upstream artifacts directly and
# verify the digests GitHub publishes for the selected architecture.
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  hurl_archive="$("/installers/verified-github-asset" Orange-OpenSource/hurl "${HURL_VERSION}" \
    "hurl-${HURL_VERSION}-x86_64-unknown-linux-gnu.tar.gz" \
    "hurl-${HURL_VERSION}-aarch64-unknown-linux-gnu.tar.gz")" \
  && mkdir --parents /tmp/hurl \
  && tar --extract --gzip --strip-components=1 \
    --directory=/tmp/hurl --file="${hurl_archive}" \
  && install --mode=0755 /tmp/hurl/bin/hurl /tmp/hurl/bin/hurlfmt /usr/local/bin/ \
  && oha_binary="$("/installers/verified-github-asset" hatoo/oha "v${OHA_VERSION}" \
    oha-linux-amd64 oha-linux-arm64)" \
  && install --mode=0755 "${oha_binary}" /usr/local/bin/oha \
  && rm --recursive --force /tmp/hurl

# Install the browser-based VS Code server.
# Version source: https://github.com/coder/code-server/releases/latest
ARG VSCODE_VERSION="4.129.0"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/vscode "${VSCODE_VERSION}"

# Open the Source Control view after the workbench starts.
COPY images/base/vscode/review-mode \
  /usr/local/lib/code-server/${VSCODE_VERSION}/lib/vscode/extensions/clor-review-mode

# Create the space user.
RUN groupadd \
    --gid 1000 \
    user \
  && useradd \
    --uid 1000 \
    --gid user \
    --groups shadow,audio \
    --home-dir /home/user \
    --create-home \
    --shell /bin/bash \
    user

COPY --chmod=0644 images/base/profile.d/space-tools.sh \
  /etc/profile.d/clor-space-tools.sh

USER user
ENV HOME="/home/user"
WORKDIR /home/user
ENV PATH="/home/user/.local/bin:/home/user/go/bin:/home/user/.npm-global/bin:${PATH}"
ENV PLAYWRIGHT_BROWSERS_PATH="/home/user/.cache/ms-playwright"

# Seed the browser editor with a review-oriented global layout.
COPY --chown=user:user images/base/vscode/settings.json \
  /home/user/.local/share/code-server/User/settings.json

# Skip Lazygit's introductory prompts in newly created spaces.
COPY --chown=user:user images/base/lazygit/config.yml \
  /home/user/.config/lazygit/config.yml

# Version source: https://open-vsx.org
# Pull Requests 0.156.0 is the newest release compatible with Code 1.129.0.
ARG GITHUB_PULL_REQUESTS_VERSION="0.156.0"
ARG GITHUB_ACTIONS_VERSION="0.32.3"
ARG GITHUB_THEME_VERSION="6.3.5"
ARG PYTHON_EXTENSION_VERSION="2026.4.0"
ARG DEBUGPY_VERSION="2026.6.0"
ARG PYTHON_ENVIRONMENTS_VERSION="1.36.0"
ARG GO_EXTENSION_VERSION="0.56.0"
ARG YAML_EXTENSION_VERSION="1.24.0"
RUN code-server --config /dev/null --force \
  --install-extension "github.vscode-pull-request-github@${GITHUB_PULL_REQUESTS_VERSION}" \
  --install-extension "github.vscode-github-actions@${GITHUB_ACTIONS_VERSION}" \
  --install-extension "github.github-vscode-theme@${GITHUB_THEME_VERSION}" \
  --install-extension "ms-python.python@${PYTHON_EXTENSION_VERSION}" \
  --install-extension "ms-python.debugpy@${DEBUGPY_VERSION}" \
  --install-extension "ms-python.vscode-python-envs@${PYTHON_ENVIRONMENTS_VERSION}" \
  --install-extension "golang.go@${GO_EXTENSION_VERSION}" \
  --install-extension "redhat.vscode-yaml@${YAML_EXTENSION_VERSION}"

# Configure per-user npm globals.
RUN mkdir --parents /home/user/.npm-global \
  && npm config set prefix /home/user/.npm-global

# Install Redocly CLI. npm keeps a second copy of every package it installs in
# ~/.npm; clear it in the same layer so the tarballs never reach the image.
ARG REDOCLY_VERSION="2.43.2"
RUN npm install --global "@redocly/cli@${REDOCLY_VERSION}" \
  && npm cache clean --force

# Posting is isolated from project Python environments while its shim stays on
# the ordinary space PATH.
ARG POSTING_VERSION="2.10.0"
RUN uv tool install "posting==${POSTING_VERSION}"

# Install pnpm and Yarn.
ARG PNPM_VERSION="11.18.0"
ARG YARN_VERSION="4.18.0"
RUN corepack enable --install-directory /home/user/.npm-global/bin \
  && corepack prepare "pnpm@${PNPM_VERSION}" --activate \
  && corepack prepare "yarn@${YARN_VERSION}" --activate \
  && npm cache clean --force

# Install TypeScript developer tooling.
# Version source: https://registry.npmjs.org/typescript/latest
ARG TYPESCRIPT_VERSION="7.0.2"
ARG TSX_VERSION="4.23.1"
RUN npm install --global \
  "typescript@${TYPESCRIPT_VERSION}" \
  "tsx@${TSX_VERSION}" \
  eslint \
  prettier \
  && npm cache clean --force

# Install Playwright and its browsers.
# Version source: https://registry.npmjs.org/@playwright/test/latest
ARG PLAYWRIGHT_VERSION="1.62.1"
RUN npm install --global "@playwright/test@${PLAYWRIGHT_VERSION}" \
  && npm cache clean --force

USER root
# Docker does not recalculate HOME or PATH when USER changes.
ENV HOME="/root"
ENV PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/root/.cache/ms-playwright"
RUN /home/user/.npm-global/bin/playwright install-deps chromium firefox \
  && rm --recursive --force /var/lib/apt/lists/*

USER user
# Restore the per-user environment before installing browser payloads.
ENV HOME="/home/user"
ENV PATH="/home/user/.local/bin:/home/user/go/bin:/home/user/.npm-global/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/home/user/.cache/ms-playwright"
# WebKit is not installed: it is the only browser nothing in a space reaches
# for by default, and it costs more than the other two combined. A space that
# needs it runs `playwright install webkit`.
RUN playwright install chromium firefox

ENV DISABLE_AUTOUPDATER="1"
COPY --chmod=0755 images/base/chromium /usr/local/bin/chromium
ENV CHROME_PATH="/usr/local/bin/chromium"

USER root
# The production entrypoint starts privileged, with no user-writable paths.
ENV HOME="/root"
ENV PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/root/.cache/ms-playwright"
# Install the Go language server, debugger, and pinned Go-native developer
# tools. GOBIN puts every binary straight into its final location, so nothing
# is left behind in a GOPATH bin directory to be copied or shadow a wrapper:
# the space PATH puts /home/user/go/bin ahead of /usr/local/bin, and a stray
# gh-dash there starts with no GH_TOKEN because it never reaches the wrapper.
# The module and build caches are BuildKit cache mounts rather than image
# content; go install writes roughly two gigabytes into them, and none of it
# is needed to run the binaries. Mount the whole of GOPATH/pkg, not just
# pkg/mod: the checksum database Go maintains lands beside it in pkg/sumdb.
ARG GOPLS_VERSION="0.23.0"
ARG DELVE_VERSION="1.27.0"
ARG GH_DASH_VERSION="4.25.2"
ARG YQ_VERSION="4.53.3"
ARG GRPCURL_VERSION="1.9.3"
ARG GRPCUI_VERSION="1.5.2"
ARG ACTIONLINT_VERSION="1.7.12"
ARG GITLEAKS_VERSION="8.30.1"
ARG OSV_SCANNER_VERSION="2.4.0"
ARG VALE_VERSION="3.15.2"
ARG DIRENV_VERSION="2.37.1"
RUN --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
  --mount=type=cache,target=/root/go/pkg,sharing=locked \
  mkdir --parents /usr/local/lib/clor \
  && GOBIN=/usr/local/lib/clor go install "github.com/dlvhdr/gh-dash/v4@v${GH_DASH_VERSION}" \
  && export GOBIN=/usr/local/bin \
  && go install "golang.org/x/tools/gopls@v${GOPLS_VERSION}" \
  && go install "github.com/go-delve/delve/cmd/dlv@v${DELVE_VERSION}" \
  && go install "github.com/mikefarah/yq/v4@v${YQ_VERSION}" \
  && go install "github.com/fullstorydev/grpcurl/cmd/grpcurl@v${GRPCURL_VERSION}" \
  && go install "github.com/fullstorydev/grpcui/cmd/grpcui@v${GRPCUI_VERSION}" \
  && go install "github.com/rhysd/actionlint/cmd/actionlint@v${ACTIONLINT_VERSION}" \
  && go install "github.com/zricethezav/gitleaks/v8@v${GITLEAKS_VERSION}" \
  && go install "github.com/google/osv-scanner/v2/cmd/osv-scanner@v${OSV_SCANNER_VERSION}" \
  && go install "github.com/errata-ai/vale/v3/cmd/vale@v${VALE_VERSION}" \
  && go install "github.com/direnv/direnv/v2@v${DIRENV_VERSION}" \
  && test ! -e /home/user/go/bin
# Resolve Playwright's Chromium once, at build time, so the wrapper on PATH
# does not have to guess at the browser revision in its directory name. The
# directory below that is architecture specific -- chrome-linux64 on amd64,
# chrome-linux on arm64 -- so match the binary by name and keep the parent a
# wildcard. chromium_headless_shell-* is excluded twice over: its directory
# does not start with "chromium-" and its binary is not named chrome. The find
# fails the build if a Playwright bump ever changes that layout again.
RUN chromium_binary="$(find /home/user/.cache/ms-playwright \
    -mindepth 3 -maxdepth 3 -type f -executable \
    -name chrome -path '*/chromium-*/*' -print -quit)" \
  && test -n "${chromium_binary}" \
  && ln --symbolic "${chromium_binary}" /usr/local/lib/clor/chromium \
  && test -x /usr/local/lib/clor/chromium
# Adapt only this exact code-server release's built-in GitHub provider. The
# original provider remains responsible for normal and GitHub Enterprise auth.
RUN test "${VSCODE_VERSION}" = "4.129.0" \
  && printf '%s  %s\n' \
    26035111b16f05f6bdaefe51f200312b12b104d0f9e60ec94c536990c5c8ac0e \
    "/usr/local/lib/code-server/${VSCODE_VERSION}/lib/vscode/extensions/github-authentication/dist/extension.js" \
    cbe87ebc0e2d1274c396b669794bec0f1f10c01985f9bb9ae6135f3af7ed97b7 \
    "/usr/local/lib/code-server/${VSCODE_VERSION}/lib/vscode/extensions/github-authentication/package.json" \
  | sha256sum --check --strict
COPY --chmod=0644 images/base/vscode/github-authentication/*.js \
  /usr/local/lib/code-server/${VSCODE_VERSION}/lib/vscode/extensions/github-authentication/
RUN github_auth_package="/usr/local/lib/code-server/${VSCODE_VERSION}/lib/vscode/extensions/github-authentication/package.json" \
  && test "$(grep --fixed-strings --count \
    '"main":"./dist/extension.js"' "${github_auth_package}")" = 1 \
  && sed --in-place \
    's#"main":"./dist/extension.js"#"main":"./clor-adapter.js"#' \
    "${github_auth_package}" \
  && test "$(jq --raw-output .main "${github_auth_package}")" = \
    "./clor-adapter.js"
# Code only hands a session to an extension that the user has consented to, or
# that product.json pre-trusts. Spaces authenticate through Clor with no user
# to click Allow, so the GitHub extensions we ship are pre-trusted too:
# without this the Actions view sits at "Sign in to GitHub" even though the
# provider above is holding a valid session. Git, Pull Requests, and Copilot
# are already on the stock list.
RUN product_json="/usr/local/lib/code-server/${VSCODE_VERSION}/lib/vscode/product.json" \
  && test "$(jq --raw-output '.trustedExtensionAuthAccess | type' \
    "${product_json}")" = "array" \
  && jq '.trustedExtensionAuthAccess |= (if index("github.vscode-github-actions") then . else . + ["github.vscode-github-actions"] end)' \
    "${product_json}" >/tmp/product.json \
  && mv /tmp/product.json "${product_json}" \
  && chmod 0644 "${product_json}" \
  && test "$(jq --raw-output '.trustedExtensionAuthAccess | index("github.vscode-github-actions") != null' \
    "${product_json}")" = "true"
COPY --chmod=0755 images/base/gh /usr/local/bin/gh
COPY --chmod=0755 images/base/gh-dash /usr/local/bin/gh-dash
COPY --chmod=0755 images/base/clor-node-app /usr/local/bin/clor-node-app
COPY --chmod=0755 images/base/clor-api-workbench /usr/local/bin/clor-api-workbench
COPY --chmod=0755 images/base/remotion-space /usr/local/bin/remotion-space
COPY --chmod=0755 images/base/start /usr/local/lib/clor/start
COPY --chmod=0644 images/base/services/code-server.toml /etc/clor/services/code-server.toml
RUN mkdir --parents \
  /etc/clor/services \
  /var/lib/clor \
  /var/log/clor

# Bookworm's tmux predates the csi-u format required for Codex to distinguish
# modified keys safely. Keep the apt package for its runtime dependencies and
# replace only the binary with the pinned upstream build. Install the matching
# agent-friendly config system-wide for both root and the space user.
COPY --from=tmux-builder /out/tmux /usr/local/bin/tmux
COPY --chmod=0644 images/base/tmux/tmux.conf /etc/tmux.conf

# Keep frequently updated agent CLIs in the final filesystem layers so rebuilds
# reuse every expensive language-tooling and browser layer above.
# Version source: https://github.com/openai/codex/releases/latest
ARG CODEX_VERSION="0.146.1"
# Version source: https://downloads.claude.ai/claude-code-releases/latest
ARG CLAUDE_VERSION="2.1.223"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,id=base-user-downloads,target=/downloads,sharing=locked,uid=1000,gid=1000 \
  gosu user env HOME=/home/user /installers/codex "${CODEX_VERSION}" \
  && gosu user env HOME=/home/user /installers/claude "${CLAUDE_VERSION}"

# Install the Claude Code plugins every space gets by default, so no space pays
# a first-run download for them. The list is the file being copied; the
# installer resolves each marketplace and asserts the plugin ends up enabled.
# This layer follows the Claude CLI it configures and precedes the Clor layer,
# whose nonce would otherwise re-push these payloads on every build.
COPY --chmod=0644 images/base/claude-plugins /usr/local/lib/clor/claude-plugins
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  gosu user env \
    HOME=/home/user \
    PATH="/home/user/.local/bin:/usr/local/bin:/usr/bin:/bin" \
    /installers/claude-plugins /usr/local/lib/clor/claude-plugins

# The Clor CLI is the one unpinned install in the image, and the build
# entrypoints supply a unique nonce so it runs on every build. Keep it alone in
# the last layer: the nonce invalidates whatever layer it sits in, and every
# node then re-pulls that layer on every build even when nothing else changed.
# Bundling the pinned agents above with it would make that a 200 MB download
# per build instead of the CLI's own handful of megabytes.
ARG CLOR_BUILD_NONCE
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,id=base-user-downloads,target=/downloads,sharing=locked,uid=1000,gid=1000 \
  test -n "${CLOR_BUILD_NONCE}" \
  && curl --fail --silent --show-error --location \
    --output /downloads/clor-install.sh https://clor.com/install.sh \
  && gosu user env \
    HOME=/home/user \
    PATH="/home/user/.local/bin:/home/user/go/bin:/home/user/.npm-global/bin:${PATH}" \
    CLOR_AUTOUPDATE=false \
    bash /downloads/clor-install.sh \
  && test -z "$(find /home/user \
    \( ! -user user -o ! -group user \) \
    -print -quit)" \
  && gosu user test -w /home/user

ENV CLOR_SPACE_USER="user"
ENV CLOR_SPACE_HOME="/home/user"
# The checkout root is the space account's home, so this repeats
# CLOR_SPACE_HOME. The space daemon re-exports the variable from that home
# rather than reading the value baked here, which is what lets placement stop
# depending on the image; the baked value is what a container run by hand
# reads, and what the agent tooling trusts as a project directory.
ENV CLOR_SPACE_DIRECTORY="/home/user"
WORKDIR /root
ENTRYPOINT ["/usr/local/lib/clor/start"]

# Exercise the completed image without copying test assets into production.
FROM base AS test
COPY --chmod=0755 images/base/tests /usr/local/lib/clor/tests
# Bash has no long spelling for -c; login mode verifies /etc/profile behavior.
# The chain is silent, so a failure reports only an exit code. To find out
# which assertion failed, prefix it with `set -x &&` and rebuild the stage.
RUN test "${HOME}" = /root \
  && test "${PWD}" = /root \
  && test "${PATH}" = /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  && test "${PLAYWRIGHT_BROWSERS_PATH}" = /root/.cache/ms-playwright \
  && bash --login -c 'test "${HOME}" = /root \
      && test "${PATH}" = /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
      && test "${PLAYWRIGHT_BROWSERS_PATH}" = /root/.cache/ms-playwright' \
  && command -v \
    bun \
    bunx \
    deno \
    go \
    golangci-lint \
    actionlint \
    ast-grep \
    bat \
    delta \
    difft \
    duckdb \
    eza \
    fd \
    gh-dash \
    gitleaks \
    grpcui \
    grpcurl \
    hurl \
    just \
    lychee \
    mise \
    nvim \
    oha \
    osv-scanner \
    posting \
    pandoc \
    presenterm \
    remotion-space \
    shellcheck \
    shfmt \
    tailwindcss \
    typst \
    typos \
    vale \
    watchexec \
    websocat \
    xh \
    ya \
    yazi \
    yq \
    xvfb-run \
    zoxide \
    >/dev/null \
  && for tool in actionlint direnv dlv gitleaks gopls grpcui grpcurl \
    osv-scanner vale yq; do \
    test "$(command -v "${tool}")" = "/usr/local/bin/${tool}" || exit 1; \
  done \
  && test -x /usr/local/lib/clor/gh-dash \
  && test "$(command -v tmux)" = /usr/local/bin/tmux \
  && test ! -e /home/user/go/bin \
  && test -z "$(find /root/go /root/.cache/go-build /home/user/go \
    /home/user/.cache/go-build -type f -print -quit 2>/dev/null)" \
  && test ! -e /usr/bin/chromium \
  && test ! -e /usr/bin/firefox-esr \
  && case "$(readlink /usr/local/lib/clor/chromium)" in \
    /home/user/.cache/ms-playwright/chromium-*/chrome-linux*/chrome) ;; \
    *) exit 1 ;; \
    esac \
  && test -x /usr/local/lib/clor/chromium \
  && gosu user env \
    HOME=/home/user \
    PATH="/home/user/.local/bin:/home/user/go/bin:/home/user/.npm-global/bin:${PATH}" \
    PLAYWRIGHT_BROWSERS_PATH=/home/user/.cache/ms-playwright \
    bash --login -c 'test "${HOME}" = /home/user \
      && [[ "${PATH}" == *"/home/user/.local/bin"* ]] \
      && [[ "${PATH}" == *"/home/user/go/bin"* ]] \
      && [[ "${PATH}" == *"/home/user/.npm-global/bin"* ]] \
      && [[ "${PATH}" != *"/root/"* ]] \
      && test "${PLAYWRIGHT_BROWSERS_PATH}" = /home/user/.cache/ms-playwright \
      && command -v claude clor codex dlv eslint gopls lazygit playwright prettier tsx tsc ya yazi >/dev/null \
      && test "$(command -v ya)" = /usr/local/bin/ya \
      && test "$(command -v yazi)" = /usr/local/bin/yazi \
      && test "$(command -v gh)" = /usr/local/bin/gh \
      && test "$(command -v gh-dash)" = /usr/local/bin/gh-dash \
      && ya --version | grep --fixed-strings --quiet "${YAZI_VERSION}" \
      && yazi --version | grep --fixed-strings --quiet "${YAZI_VERSION}" \
      && grep --fixed-strings --line-regexp --quiet \
        "disableStartupPopups: true" \
        /home/user/.config/lazygit/config.yml' \
  && bash -n \
    /usr/local/bin/chromium \
    /usr/local/bin/clor-api-workbench \
    /usr/local/bin/clor-node-app \
    /usr/local/bin/gh \
    /usr/local/bin/gh-dash \
    /usr/local/bin/remotion-space \
    /usr/local/lib/clor/start \
    /usr/local/lib/clor/tests/gh \
    /usr/local/lib/clor/tests/gh-dash \
    /usr/local/lib/clor/tests/claude-plugins \
    /usr/local/lib/clor/tests/clor-api-workbench \
    /usr/local/lib/clor/tests/clor-node-app \
    /usr/local/lib/clor/tests/start \
    /usr/local/lib/clor/tests/remotion-space \
    /etc/profile.d/clor-space-tools.sh \
  && TERM=xterm-256color COLORTERM=truecolor script --quiet --return --command \
    "tmux -L tuicheck new-session \"tmux display -p '#{client_termfeatures}' >/tmp/termfeatures; tmux show-options -sv extended-keys >/tmp/extended-keys; tmux show-options -gv extended-keys-format >/tmp/extended-keys-format; tmux show-options -gv allow-passthrough >/tmp/allow-passthrough; tmux show-options -gv focus-events >/tmp/focus-events\"" \
    /dev/null >/dev/null \
  && grep --fixed-strings --quiet RGB /tmp/termfeatures \
  && grep --fixed-strings --quiet extkeys /tmp/termfeatures \
  && grep --fixed-strings --line-regexp --quiet on /tmp/extended-keys \
  && grep --fixed-strings --line-regexp --quiet csi-u /tmp/extended-keys-format \
  && grep --fixed-strings --line-regexp --quiet on /tmp/allow-passthrough \
  && grep --fixed-strings --line-regexp --quiet on /tmp/focus-events \
  && rm --force \
    /tmp/allow-passthrough \
    /tmp/extended-keys \
    /tmp/extended-keys-format \
    /tmp/focus-events \
    /tmp/termfeatures \
  && test "$(jq --raw-output '.trustedExtensionAuthAccess | index("github.vscode-github-actions") != null' \
    /usr/local/lib/code-server/${VSCODE_VERSION}/lib/vscode/product.json)" = "true" \
  && node --check \
    /usr/local/lib/code-server/${VSCODE_VERSION}/lib/vscode/extensions/github-authentication/clor-adapter.js \
  && node --check \
    /usr/local/lib/code-server/${VSCODE_VERSION}/lib/vscode/extensions/github-authentication/clor-github-provider.js \
  && node --check /usr/local/lib/clor/tests/github-authentication.js \
  && node --check /usr/local/lib/clor/tests/github-authentication-integration.js \
  && node --check /usr/local/lib/clor/tests/vscode-auth-probe/extension.js \
  && shellcheck \
    /usr/local/bin/chromium \
    /usr/local/bin/clor-api-workbench \
    /usr/local/bin/clor-node-app \
    /usr/local/bin/gh \
    /usr/local/bin/gh-dash \
    /usr/local/bin/remotion-space \
    /usr/local/lib/clor/start \
    /usr/local/lib/clor/tests/gh \
    /usr/local/lib/clor/tests/gh-dash \
    /usr/local/lib/clor/tests/claude-plugins \
    /usr/local/lib/clor/tests/clor-api-workbench \
    /usr/local/lib/clor/tests/clor-node-app \
    /usr/local/lib/clor/tests/start \
    /usr/local/lib/clor/tests/remotion-space \
    /usr/local/lib/clor/tests/bin/clor \
    /usr/local/lib/clor/tests/bin/curl \
    /usr/local/lib/clor/tests/bin/remotion-git \
    /usr/local/lib/clor/tests/bin/remotion-npx \
    /usr/local/lib/clor/tests/bin/remotion-package-manager \
    /usr/local/lib/clor/tests/bin/remotion-remotion \
    /etc/profile.d/clor-space-tools.sh \
  && shfmt -d -i 4 -ci \
    /usr/local/bin/chromium \
    /usr/local/bin/clor-api-workbench \
    /usr/local/bin/clor-node-app \
    /usr/local/bin/gh \
    /usr/local/bin/gh-dash \
    /usr/local/bin/remotion-space \
    /usr/local/lib/clor/start \
    /usr/local/lib/clor/tests/gh \
    /usr/local/lib/clor/tests/gh-dash \
    /usr/local/lib/clor/tests/claude-plugins \
    /usr/local/lib/clor/tests/clor-api-workbench \
    /usr/local/lib/clor/tests/clor-node-app \
    /usr/local/lib/clor/tests/start \
    /usr/local/lib/clor/tests/remotion-space \
    /usr/local/lib/clor/tests/bin/clor \
    /usr/local/lib/clor/tests/bin/curl \
    /usr/local/lib/clor/tests/bin/remotion-git \
    /usr/local/lib/clor/tests/bin/remotion-npx \
    /usr/local/lib/clor/tests/bin/remotion-package-manager \
    /usr/local/lib/clor/tests/bin/remotion-remotion \
    /etc/profile.d/clor-space-tools.sh \
  && node /usr/local/lib/clor/tests/github-authentication.js \
  && /usr/local/lib/clor/tests/claude-plugins \
  && /usr/local/lib/clor/tests/clor-node-app \
  && /usr/local/lib/clor/tests/clor-api-workbench \
  && /usr/local/lib/clor/tests/gh \
  && /usr/local/lib/clor/tests/gh-dash \
  && /usr/local/lib/clor/tests/remotion-space \
  && /usr/local/lib/clor/tests/start
RUN gosu user env \
    HOME=/home/user \
    PLAYWRIGHT_BROWSERS_PATH=/home/user/.cache/ms-playwright \
    node /usr/local/lib/clor/tests/github-authentication-integration.js
RUN gosu user env \
    HOME=/home/user \
    PLAYWRIGHT_BROWSERS_PATH=/home/user/.cache/ms-playwright \
    PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH="${CHROME_PATH}" \
    node /usr/local/lib/clor/tests/playwright.js
RUN gosu user env \
    HOME=/home/user \
    PLAYWRIGHT_BROWSERS_PATH=/home/user/.cache/ms-playwright \
    node /usr/local/lib/clor/tests/playwright.js \
  && gosu user env \
    HOME=/home/user \
    PLAYWRIGHT_BROWSERS_PATH=/home/user/.cache/ms-playwright \
    PLAYWRIGHT_HEADED=true \
    xvfb-run --auto-servernum node /usr/local/lib/clor/tests/playwright.js

FROM base AS production
