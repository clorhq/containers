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
  gnupg \
  python3 \
  python3-pip \
  python3-venv \
  python3-dev \
  python3-setuptools \
  python3-wheel \
  pipx \
  make \
  ripgrep \
  g++ \
  dumb-init \
  runit \
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
  && rm --recursive --force /var/lib/apt/lists/*

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

# Install Go system-wide.
# Version source: https://go.dev/VERSION?m=text
ARG GO_VERSION="1.26.5"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/go "${GO_VERSION}"

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

# Install Codex for the space user.
# Version source: https://github.com/openai/codex/releases/latest
ARG CODEX_VERSION="0.144.6"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,id=base-user-downloads,target=/downloads,sharing=locked,uid=1000,gid=1000 \
  /installers/codex "${CODEX_VERSION}"

# Install Claude for the space user.
# Version source: https://downloads.claude.ai/claude-code-releases/latest
ARG CLAUDE_VERSION="2.1.214"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,id=base-user-downloads,target=/downloads,sharing=locked,uid=1000,gid=1000 \
  /installers/claude "${CLAUDE_VERSION}"

# Configure per-user npm globals.
RUN mkdir --parents /home/user/.npm-global \
  && npm config set prefix /home/user/.npm-global

# Install Redocly CLI.
ARG REDOCLY_VERSION="2.30.5"
RUN npm install --global "@redocly/cli@${REDOCLY_VERSION}"

# Install pnpm and Yarn.
ARG PNPM_VERSION="11.14.0"
ARG YARN_VERSION="4.17.1"
RUN corepack enable --install-directory /home/user/.npm-global/bin \
  && corepack prepare "pnpm@${PNPM_VERSION}" --activate \
  && corepack prepare "yarn@${YARN_VERSION}" --activate

# Install core TypeScript tools.
# Version source: https://registry.npmjs.org/typescript/latest
ARG TYPESCRIPT_VERSION="7.0.2"
ARG TSX_VERSION="4.23.1"
RUN npm install --global "typescript@${TYPESCRIPT_VERSION}" "tsx@${TSX_VERSION}"

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
ENV CHROME_PATH="/usr/bin/chromium"

# Install the fallback Clor release.
ARG CLOR_VERSION="1.8.5"
RUN curl --fail --silent --show-error --location https://clor.com/install.sh \
  | CLOR_VERSION="v${CLOR_VERSION}" bash

# Verify home ownership.
RUN test -z "$(find /home/user \
    \( ! -user user -o ! -group user \) \
    -print -quit)" \
  && test -w /home/user

USER root
# The production entrypoint starts privileged, with no user-writable paths.
ENV HOME="/root"
ENV PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/root/.cache/ms-playwright"
COPY --chmod=0755 images/base/entrypoint /usr/local/lib/clor/entrypoint
RUN mkdir --parents \
  /etc/service \
  /var/lib/clor

ENV CLOR_SPACE_USER="user"
ENV CLOR_SPACE_HOME="/home/user"
WORKDIR /root
ENTRYPOINT ["/usr/bin/dumb-init", "--", "/usr/local/lib/clor/entrypoint"]

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
  && command -v shellcheck shfmt xvfb-run >/dev/null \
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
      && command -v playwright >/dev/null' \
  && bash -n \
    /usr/local/lib/clor/entrypoint \
    /usr/local/lib/clor/tests/entrypoint \
    /etc/profile.d/clor-space-tools.sh \
  && shellcheck \
    /usr/local/lib/clor/entrypoint \
    /usr/local/lib/clor/tests/entrypoint \
    /usr/local/lib/clor/tests/bin/clor \
    /usr/local/lib/clor/tests/bin/curl \
    /etc/profile.d/clor-space-tools.sh \
  && shfmt -d -i 4 -ci \
    /usr/local/lib/clor/entrypoint \
    /usr/local/lib/clor/tests/entrypoint \
    /usr/local/lib/clor/tests/bin/clor \
    /usr/local/lib/clor/tests/bin/curl \
    /etc/profile.d/clor-space-tools.sh \
  && /usr/local/lib/clor/tests/entrypoint
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
