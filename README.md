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

- **Agents and workspace:** Clor, [Claude Code](https://github.com/anthropics/claude-code),
  [Codex](https://github.com/openai/codex),
  [code-server](https://github.com/coder/code-server), and VS Code extensions
  for GitHub pull requests and Actions, Python, Go, and YAML.
- **Languages and package managers:** [Node.js](https://github.com/nodejs/node)
  and [npm](https://github.com/npm/cli), [Bun](https://github.com/oven-sh/bun),
  [Deno](https://github.com/denoland/deno), [Go](https://github.com/golang/go),
  [Python](https://github.com/python/cpython), pip, pipx,
  [uv](https://github.com/astral-sh/uv), [pnpm](https://github.com/pnpm/pnpm),
  and [Yarn](https://github.com/yarnpkg/berry).
- **Editors and terminal workflow:** [Neovim](https://github.com/neovim/neovim),
  Vim, Nano, [tmux](https://github.com/tmux/tmux), screen,
  [Lazygit](https://github.com/jesseduffield/lazygit),
  [Yazi](https://github.com/sxyazi/yazi), [fzf](https://github.com/junegunn/fzf),
  [Glow](https://github.com/charmbracelet/glow),
  [bat](https://github.com/sharkdp/bat), [eza](https://github.com/eza-community/eza),
  [fd](https://github.com/sharkdp/fd),
  [zoxide](https://github.com/ajeetdsouza/zoxide), tree, ncdu, htop, and btop.
- **Source control and code quality:** Git, Git LFS,
  [GitHub CLI (`gh`)](https://github.com/cli/cli),
  [gh-dash](https://github.com/dlvhdr/gh-dash),
  [delta](https://github.com/dandavison/delta),
  [difftastic](https://github.com/Wilfred/difftastic),
  [ESLint](https://github.com/eslint/eslint),
  [Prettier](https://github.com/prettier/prettier),
  [ShellCheck](https://github.com/koalaman/shellcheck),
  [shfmt](https://github.com/mvdan/sh),
  [golangci-lint](https://github.com/golangci/golangci-lint),
  [actionlint](https://github.com/rhysd/actionlint),
  [ast-grep](https://github.com/ast-grep/ast-grep),
  [typos](https://github.com/crate-ci/typos),
  [Vale](https://github.com/vale-cli/vale), and
  [lychee](https://github.com/lycheeverse/lychee).
- **Development and automation:** [TypeScript](https://github.com/microsoft/TypeScript),
  [tsx](https://github.com/privatenumber/tsx),
  [Tailwind CSS](https://github.com/tailwindlabs/tailwindcss),
  [gopls](https://github.com/golang/tools/tree/master/gopls),
  [Delve](https://github.com/go-delve/delve), [mise](https://github.com/jdx/mise),
  [direnv](https://github.com/direnv/direnv),
  [just](https://github.com/casey/just),
  [watchexec](https://github.com/watchexec/watchexec),
  [Redocly CLI](https://github.com/Redocly/redocly-cli), and the
  [Remotion](https://github.com/remotion-dev/remotion) space helper.
- **APIs, networking, and security:** [Posting](https://github.com/darrenburns/posting),
  [xh](https://github.com/ducaale/xh), [Hurl](https://github.com/Orange-OpenSource/hurl),
  [grpcurl](https://github.com/fullstorydev/grpcurl),
  [grpcui](https://github.com/fullstorydev/grpcui),
  [websocat](https://github.com/vi/websocat), [oha](https://github.com/hatoo/oha),
  [yq](https://github.com/mikefarah/yq),
  [gitleaks](https://github.com/gitleaks/gitleaks),
  [OSV-Scanner](https://github.com/google/osv-scanner), curl, wget, OpenSSH,
  rsync, [rclone](https://github.com/rclone/rclone), ping, traceroute, mtr,
  netcat, socat, tcpdump, DNS tools, and whois.
- **Data, documents, and media:** [DuckDB](https://github.com/duckdb/duckdb),
  SQLite, jq, [lnav](https://github.com/tstack/lnav),
  [Pandoc](https://github.com/jgm/pandoc), [Typst](https://github.com/typst/typst),
  [Presenterm](https://github.com/mfontanini/presenterm),
  [ffmpeg](https://github.com/FFmpeg/FFmpeg), SoX, ImageMagick, ExifTool,
  Poppler tools, and common archive utilities including zip, 7-Zip, xz, and
  zstd.
- **Browser automation:** [Playwright](https://github.com/microsoft/playwright)
  with Chromium and Firefox. WebKit is not preinstalled; add it when needed
  with `playwright install webkit`.
