ARG BASE_IMAGE="software-development-base:latest"

FROM ${BASE_IMAGE}

USER root

# The base image already ships Node with pnpm, yarn, typescript, and tsx.
# This image adds the Bun and Deno runtimes system-wide into /usr/local/bin
# so all three JavaScript runtimes are ready.
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

USER user
ENV HOME="/home/user"
ENV PATH="/home/user/.local/bin:/home/user/go/bin:/home/user/.npm-global/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/home/user/.cache/ms-playwright"
WORKDIR /home/user

# Global JS/TS developer tooling on top of the base: the ESLint linter and
# the Prettier formatter.
RUN npm install --global eslint prettier

USER root
ENV HOME="/root"
ENV PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/root/.cache/ms-playwright"
WORKDIR /root
