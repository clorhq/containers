ARG OS_IMAGE="debian:bookworm-slim"

FROM ${OS_IMAGE} AS base

# Terminal and UTF-8 defaults.
ENV TERM="xterm-256color"
ENV COLORTERM="truecolor"
ENV LANG="C.UTF-8"
ENV LC_ALL="C.UTF-8"

# Install system tools.
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
  firefox-esr \
  chromium \
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
ARG GITHUB_VERSION="2.96.0"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/github "${GITHUB_VERSION}"

# Install Node.js
# Version source: https://nodejs.org/dist/index.json (latest entry with an LTS name)
ARG NODE_VERSION="24.18.0"
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
ARG DENO_VERSION="2.9.2"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/deno "${DENO_VERSION}"

# Tailwind CSS standalone CLI.
ARG TAILWINDCSS_VERSION="4.1.13"
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
ARG UV_VERSION="0.11.28"
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
  "yazi-fm@${YAZI_VERSION}" \
  "yazi-cli@${YAZI_VERSION}" \
  "git-delta@${DELTA_VERSION}" \
  "difftastic@${DIFFTASTIC_VERSION}" \
  "eza@${EZA_VERSION}" \
  "zoxide@${ZOXIDE_VERSION}" \
  "just@${JUST_VERSION}" \
  "watchexec-cli@${WATCHEXEC_VERSION}" \
  "mise@${MISE_VERSION}" \
  "ast-grep@${AST_GREP_VERSION}" \
  "xh@${XH_VERSION}" \
  "oha@${OHA_VERSION}" \
  "websocat@${WEBSOCAT_VERSION}" \
  "typos-cli@${TYPOS_VERSION}" \
  "lychee@${LYCHEE_VERSION}" \
  "typst-cli@${TYPST_VERSION}" \
  "presenterm@${PRESENTERM_VERSION}" \
  && if ! command -v ast-grep >/dev/null 2>&1; then \
    ln --symbolic "$(command -v sg)" /usr/local/bin/ast-grep; \
  fi

# Hurl's cargo-binstall metadata does not match the layout of its current
# release archive. Install the exact upstream archive directly and verify the
# digest GitHub publishes for the selected architecture.
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  hurl_archive="$("/installers/verified-github-asset" Orange-OpenSource/hurl "${HURL_VERSION}" \
    "hurl-${HURL_VERSION}-x86_64-unknown-linux-gnu.tar.gz" \
    "hurl-${HURL_VERSION}-aarch64-unknown-linux-gnu.tar.gz")" \
  && mkdir --parents /tmp/hurl \
  && tar --extract --gzip --strip-components=1 \
    --directory=/tmp/hurl --file="${hurl_archive}" \
  && install --mode=0755 /tmp/hurl/bin/hurl /tmp/hurl/bin/hurlfmt /usr/local/bin/ \
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

# Install Go language tooling for the space user.
ARG GOPLS_VERSION="0.21.1"
RUN go install "golang.org/x/tools/gopls@v${GOPLS_VERSION}"

ARG DELVE_VERSION="1.7.3"
RUN go install "github.com/go-delve/delve/cmd/dlv@v${DELVE_VERSION}"

# Install pinned Go-native developer tools into the system path. gh-dash is
# kept behind its authentication wrapper, just as the real gh binary is.
ARG GH_DASH_VERSION="4.25.2"
ARG YQ_VERSION="4.53.3"
ARG GRPCURL_VERSION="1.9.3"
ARG GRPCUI_VERSION="1.5.2"
ARG ACTIONLINT_VERSION="1.7.12"
ARG GITLEAKS_VERSION="8.30.1"
ARG OSV_SCANNER_VERSION="2.4.0"
ARG VALE_VERSION="3.15.2"
ARG DIRENV_VERSION="2.37.1"
RUN go install "github.com/dlvhdr/gh-dash/v4@v${GH_DASH_VERSION}" \
  && go install "github.com/mikefarah/yq/v4@v${YQ_VERSION}" \
  && go install "github.com/fullstorydev/grpcurl/cmd/grpcurl@v${GRPCURL_VERSION}" \
  && go install "github.com/fullstorydev/grpcui/cmd/grpcui@v${GRPCUI_VERSION}" \
  && go install "github.com/rhysd/actionlint/cmd/actionlint@v${ACTIONLINT_VERSION}" \
  && go install "github.com/zricethezav/gitleaks/v8@v${GITLEAKS_VERSION}" \
  && go install "github.com/google/osv-scanner/v2/cmd/osv-scanner@v${OSV_SCANNER_VERSION}" \
  && go install "github.com/errata-ai/vale/v3/cmd/vale@v${VALE_VERSION}" \
  && go install "github.com/direnv/direnv/v2@v${DIRENV_VERSION}"

# Configure per-user npm globals.
RUN mkdir --parents /home/user/.npm-global \
  && npm config set prefix /home/user/.npm-global

# Install Redocly CLI.
ARG REDOCLY_VERSION="2.30.5"
RUN npm install --global "@redocly/cli@${REDOCLY_VERSION}"

# Posting is isolated from project Python environments while its shim stays on
# the ordinary space PATH.
ARG POSTING_VERSION="2.10.0"
RUN uv tool install "posting==${POSTING_VERSION}"

# Install pnpm and Yarn.
ARG PNPM_VERSION="11.14.0"
ARG YARN_VERSION="4.17.1"
RUN corepack enable --install-directory /home/user/.npm-global/bin \
  && corepack prepare "pnpm@${PNPM_VERSION}" --activate \
  && corepack prepare "yarn@${YARN_VERSION}" --activate

# Install TypeScript developer tooling.
# Version source: https://registry.npmjs.org/typescript/latest
ARG TYPESCRIPT_VERSION="7.0.2"
ARG TSX_VERSION="4.23.1"
RUN npm install --global \
  "typescript@${TYPESCRIPT_VERSION}" \
  "tsx@${TSX_VERSION}" \
  eslint \
  prettier

# Install Playwright and the complete browser stack.
# Version source: https://registry.npmjs.org/@playwright/test/latest
ARG PLAYWRIGHT_VERSION="1.61.1"
RUN npm install --global "@playwright/test@${PLAYWRIGHT_VERSION}"

USER root
# Docker does not recalculate HOME or PATH when USER changes.
ENV HOME="/root"
ENV PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/root/.cache/ms-playwright"
RUN /home/user/.npm-global/bin/playwright install-deps chromium firefox webkit \
  && rm --recursive --force /var/lib/apt/lists/*

USER user
# Restore the per-user environment before installing browser payloads.
ENV HOME="/home/user"
ENV PATH="/home/user/.local/bin:/home/user/go/bin:/home/user/.npm-global/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/home/user/.cache/ms-playwright"
RUN playwright install chromium firefox webkit

ENV DISABLE_AUTOUPDATER="1"
COPY --chmod=0755 images/base/chromium /usr/local/bin/chromium
ENV CHROME_PATH="/usr/local/bin/chromium"

USER root
# The production entrypoint starts privileged, with no user-writable paths.
ENV HOME="/root"
ENV PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/root/.cache/ms-playwright"
# Adapt only this exact code-server release's built-in GitHub provider. The
# original provider remains responsible for normal and GitHub Enterprise auth.
RUN install --mode=0755 /home/user/go/bin/gh-dash /usr/local/lib/clor/gh-dash \
  && for tool in yq grpcurl grpcui actionlint gitleaks osv-scanner vale direnv; do \
    install --mode=0755 "/home/user/go/bin/${tool}" "/usr/local/bin/${tool}"; \
  done
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

# Keep frequently updated agent CLIs in the final filesystem layer so rebuilds
# reuse every expensive language-tooling and browser layer above.
# Version source: https://github.com/openai/codex/releases/latest
ARG CODEX_VERSION="0.145.0"
# Version source: https://downloads.claude.ai/claude-code-releases/latest
ARG CLAUDE_VERSION="2.1.220"
# The build entrypoints supply a unique value so the unpinned Clor installer
# runs on every image build without invalidating any earlier layer.
ARG CLOR_BUILD_NONCE
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,id=base-user-downloads,target=/downloads,sharing=locked,uid=1000,gid=1000 \
  test -n "${CLOR_BUILD_NONCE}" \
  && gosu user env HOME=/home/user /installers/codex "${CODEX_VERSION}" \
  && gosu user env HOME=/home/user /installers/claude "${CLAUDE_VERSION}" \
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
ENV CLOR_SPACE_DIRECTORY="/space"
WORKDIR /root
ENTRYPOINT ["/usr/local/lib/clor/start"]

# Exercise the completed image without copying test assets into production.
FROM base AS test
COPY --chmod=0755 images/base/tests /usr/local/lib/clor/tests
# Bash has no long spelling for -c; login mode verifies /etc/profile behavior.
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
      && command -v claude clor codex dlv eslint gopls lazygit playwright prettier tsx tsc >/dev/null \
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
    /usr/local/lib/clor/tests/clor-api-workbench \
    /usr/local/lib/clor/tests/clor-node-app \
    /usr/local/lib/clor/tests/start \
    /usr/local/lib/clor/tests/remotion-space \
    /etc/profile.d/clor-space-tools.sh \
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
