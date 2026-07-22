# software-development images

Open container images for software development.

Every downloaded tool, including Clor, is pinned to an explicit version and overridable at build time. Images are built for `linux/amd64` and `linux/arm64` and published to `ghcr.io/clorhq`.

## Images

| Image | Toolchain | Includes |
| --- | --- | --- |
| `software-development-base` | Node, Go, uv | Clor, Claude, Codex, code-server, Redocly, pnpm, yarn, typescript, tsx, gh, git, ripgrep, fzf, jq, ffmpeg, imagemagick |
| `software-development-rust` | rustup Rust | clippy, rustfmt, rust-src, rust-analyzer, cargo-nextest, cargo-watch, cargo-edit, cargo-audit |
| `software-development-go` | Go (from base) | gopls, delve, golangci-lint |
| `software-development-typescript` | Node, Bun, Deno | eslint, prettier, tailwindcss |
| `software-development-python` | uv-managed CPython | ruff, mypy, pyright, poetry, ipython |
| `software-development-zig` | Zig | zls |
| `software-development-ruby` | Ruby | bundler, rubocop, solargraph |

Each language image is `FROM software-development-base`, so the agents, OS
tooling, and the base toolchains are present in every image. The language
images override the managed entrypoint with an unprivileged Bash default,
making them suitable as custom space images.
