ARG BASE_IMAGE="software-development-base:latest"

FROM ${BASE_IMAGE}

USER root

# Zig compiler, downloaded from a community mirror and verified against the
# official minisign signature hosted on ziglang.org.
# Version source: https://ziglang.org/download/index.json (latest tagged entry)
ARG ZIG_VERSION="0.16.0"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/zig "${ZIG_VERSION}"

# ZLS, the Zig language server.
ARG ZLS_VERSION="0.16.0"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/zls "${ZLS_VERSION}"

USER user
ENV HOME="/home/user"
ENV PATH="/home/user/.local/bin:/home/user/go/bin:/home/user/.npm-global/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/home/user/.cache/ms-playwright"
WORKDIR /home/user
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["bash"]
