# software-development images

Open container images for software development.

Every downloaded tool, including Clor, is pinned to an explicit version and overridable at build time. Images are built for `linux/amd64` and `linux/arm64` and published to `ghcr.io/clorhq`.

## Images

| Image | Toolchain | Includes |
| --- | --- | --- |
| `software-development-base` | Node, Bun, Deno, Go, Python/uv | Clor, Claude, Codex, code-server, Playwright, gh/gh-dash, lazygit, Yazi, Neovim, API/network clients, quality/security scanners, logs/data/content tools |
| `software-development-rust` | rustup Rust | clippy, rustfmt, rust-src, rust-analyzer, cargo-nextest, cargo-watch, cargo-edit, cargo-audit |
| `software-development-go` | Go (from base) | gopls, delve, golangci-lint |
| `software-development-typescript` | Node, Bun, Deno | eslint, prettier, tailwindcss |
| `software-development-python` | uv-managed CPython | ruff, mypy, pyright, poetry, ipython |
| `software-development-zig` | Zig | zls |
| `software-development-ruby` | Ruby | bundler, rubocop, solargraph |
| `software-development-data` | Base image with pinned CPython | JupyterLab, Marimo, Harlequin, pandas, Polars, PyArrow, DuckDB |
| `software-development-devops` | Base image | Web/cloud deployment CLIs, Kubernetes, IaC, signing, and registry tools |

Each language image is `FROM software-development-base`, so the agents, OS
tooling, and the base toolchains are present in every image. The language
images override the managed entrypoint with an unprivileged Bash default,
making them suitable as custom space images.

The Data and DevOps images are intentionally pull-on-demand variants. Their
services and terminal helpers only open local interactive interfaces: image
startup never authenticates to a provider, selects a deployment target, or
changes external infrastructure. Provider credentials remain runtime state,
supplied interactively or through Clor secrets.

## Base image tools

`software-development-base` includes the practical tools below. Exact versions
are pinned in [`images/base/Dockerfile`](images/base/Dockerfile).

- **Agents and workspace:** Clor, Claude Code, Codex, code-server, and VS Code
  extensions for GitHub pull requests and Actions, Python, Go, and YAML.
- **Languages and package managers:** Node.js and npm, Bun, Deno, Go, Python,
  pip, pipx, uv, pnpm, and Yarn.
- **Editors and terminal workflow:** Neovim, Vim, Nano, tmux, screen, Lazygit,
  Yazi, fzf, Glow, bat, eza, fd, zoxide, tree, ncdu, htop, and btop.
- **Source control and code quality:** Git, Git LFS, GitHub CLI (`gh`), gh-dash,
  delta, difftastic, ESLint, Prettier, ShellCheck, shfmt, golangci-lint,
  actionlint, ast-grep, typos, Vale, and lychee.
- **Development and automation:** TypeScript, tsx, Tailwind CSS, gopls, Delve,
  mise, direnv, just, watchexec, Redocly CLI, and the Remotion space helper.
- **APIs, networking, and security:** Posting, xh, Hurl, grpcurl, grpcui,
  websocat, oha, yq, gitleaks, OSV-Scanner, curl, wget, OpenSSH, rsync, rclone,
  ping, traceroute, mtr, netcat, socat, tcpdump, DNS tools, and whois.
- **Data, documents, and media:** DuckDB, SQLite, jq, lnav, Pandoc, Typst,
  Presenterm, ffmpeg, SoX, ImageMagick, ExifTool, Poppler tools, and common
  archive utilities including zip, 7-Zip, xz, and zstd.
- **Browser automation:** Playwright with Chromium and Firefox. WebKit is not
  preinstalled; add it when needed with `playwright install webkit`.
