ARG BASE_IMAGE="software-development-base:latest"

FROM ${BASE_IMAGE}

# Install Rust via rustup as the user account so the toolchain and cargo state
# live under /home/user/.rustup and /home/user/.cargo, matching the base image's
# per-user tool layout. clippy, rustfmt, rust-src and rust-analyzer ship as
# rustup components so the linter, formatter and language server always match
# the pinned toolchain. sh -s feeds the piped rustup-init script the flags
# after --; rustup-init's -y (no long form) skips the confirmation prompt.
USER user
ENV HOME="/home/user"
ENV PATH="/home/user/.local/bin:/home/user/go/bin:/home/user/.npm-global/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/home/user/.cache/ms-playwright"
# Version source: https://static.rust-lang.org/dist/channel-rust-stable.toml
ARG RUST_VERSION="1.97.0"
RUN curl --proto '=https' --tlsv1.2 --fail --silent --show-error --location \
    https://sh.rustup.rs \
  | sh -s -- -y --profile minimal --no-modify-path --default-toolchain "${RUST_VERSION}" \
    --component clippy --component rustfmt --component rust-src --component rust-analyzer
ENV PATH="/home/user/.cargo/bin:${PATH}"

# Cargo development tools: the nextest test runner, a file watcher for
# rebuild-on-save, cargo-edit for dependency management, and cargo-audit for
# vulnerability scanning.
RUN cargo install --locked cargo-nextest cargo-watch cargo-edit cargo-audit

WORKDIR /home/user
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["bash"]
