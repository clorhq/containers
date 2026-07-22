ARG BASE_IMAGE="software-development-base:latest"

FROM ${BASE_IMAGE}

USER user
ENV HOME="/home/user"
ENV PATH="/home/user/.local/bin:/home/user/go/bin:/home/user/.npm-global/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/home/user/.cache/ms-playwright"
WORKDIR /home/user

# uv ships in the base image. Install and pin a uv-managed CPython, then
# expose it as python / python3 on PATH via /home/user/.local/bin. The pin is
# global (the user-level uv config), never a .python-version in the working
# directory, so it applies no matter where under the user's home the managed
# bootstrap checks out the space's repositories.
# Version source: https://www.python.org/api/v2/downloads/release/?is_published=true
ARG PYTHON_VERSION="3.14.6"
ENV UV_PYTHON_INSTALL_DIR="/home/user/.local/uv/python"
RUN uv python install "${PYTHON_VERSION}" \
  && uv python pin --global "${PYTHON_VERSION}" \
  && ln --symbolic --force "$(uv python find ${PYTHON_VERSION})" /home/user/.local/bin/python \
  && ln --symbolic --force "$(uv python find ${PYTHON_VERSION})" /home/user/.local/bin/python3

# Python developer tooling installed as isolated uv tools: the Ruff
# linter/formatter, the mypy and pyright type checkers, the Poetry package
# manager, and the IPython shell. Tool shims land in /home/user/.local/bin.
RUN uv tool install ruff \
  && uv tool install mypy \
  && uv tool install pyright \
  && uv tool install poetry \
  && uv tool install ipython

USER root
ENV HOME="/root"
ENV PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/root/.cache/ms-playwright"
WORKDIR /root
