ARG BASE_IMAGE="software-development-base:latest"

FROM ${BASE_IMAGE}

USER root

# The Go toolchain ships in the base image. This image adds the linter,
# language server, and debugger.

# golangci-lint release binary.
ARG GOLANGCI_LINT_VERSION="2.12.2"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=cache,target=/downloads,sharing=locked \
  /installers/golangci-lint "${GOLANGCI_LINT_VERSION}"

USER user
ENV HOME="/home/user"
ENV PATH="/home/user/.local/bin:/home/user/go/bin:/home/user/.npm-global/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/home/user/.cache/ms-playwright"

# gopls language server and the delve debugger build from source into
# /home/user/go/bin, which is already on PATH from the base image.
ARG GOPLS_VERSION="0.21.1"
RUN go install "golang.org/x/tools/gopls@v${GOPLS_VERSION}"

ARG DELVE_VERSION="1.7.3"
RUN go install "github.com/go-delve/delve/cmd/dlv@v${DELVE_VERSION}"

WORKDIR /home/user
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["bash"]
