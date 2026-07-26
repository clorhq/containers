ARG BASE_IMAGE="software-development-python:latest"

FROM ${BASE_IMAGE} AS data

USER user
ENV HOME="/home/user"
ENV PATH="/home/user/.local/bin:/home/user/go/bin:/home/user/.npm-global/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/home/user/.cache/ms-playwright"
ENV CLOR_DATA_VENV="/home/user/.local/share/clor/data-venv"
WORKDIR /home/user

ARG JUPYTERLAB_VERSION="4.6.2"
ARG MARIMO_VERSION="0.23.15"
ARG HARLEQUIN_VERSION="2.5.2"
ARG HARLEQUIN_POSTGRES_VERSION="1.3.1"
ARG HARLEQUIN_MYSQL_VERSION="1.3.0"
ARG PANDAS_VERSION="3.0.5"
ARG POLARS_VERSION="1.43.0"
ARG PYARROW_VERSION="25.0.0"
ARG DUCKDB_VERSION="1.5.5"
ARG IPYKERNEL_VERSION="7.3.0"
RUN uv venv "${CLOR_DATA_VENV}" \
  && uv pip install --python "${CLOR_DATA_VENV}/bin/python" \
    "jupyterlab==${JUPYTERLAB_VERSION}" \
    "marimo[recommended]==${MARIMO_VERSION}" \
    "harlequin[postgres,mysql,s3]==${HARLEQUIN_VERSION}" \
    "harlequin-postgres==${HARLEQUIN_POSTGRES_VERSION}" \
    "harlequin-mysql==${HARLEQUIN_MYSQL_VERSION}" \
    "pandas==${PANDAS_VERSION}" \
    "polars==${POLARS_VERSION}" \
    "pyarrow==${PYARROW_VERSION}" \
    "duckdb==${DUCKDB_VERSION}" \
    "ipykernel==${IPYKERNEL_VERSION}" \
  && for command in harlequin ipython jupyter jupyter-lab marimo; do \
    ln --symbolic "${CLOR_DATA_VENV}/bin/${command}" "/home/user/.local/bin/${command}"; \
  done \
  && "${CLOR_DATA_VENV}/bin/python" -m ipykernel install \
    --user --name clor-data --display-name "Clor Data"

USER root
ENV HOME="/root"
ENV PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/root/.cache/ms-playwright"
COPY --chmod=0644 images/data/services/jupyterlab.toml /etc/clor/services/jupyterlab.toml
COPY --chmod=0644 images/data/services/marimo.toml /etc/clor/services/marimo.toml
EXPOSE 8888 2718
WORKDIR /root

FROM data AS test
RUN gosu user env \
    HOME=/home/user \
    PATH="/home/user/.local/bin:/home/user/go/bin:/home/user/.npm-global/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
    /home/user/.local/share/clor/data-venv/bin/python -c \
    'import duckdb, ipykernel, marimo, pandas, polars, pyarrow' \
  && gosu user test -w /home/user/.local/share/clor/data-venv \
  && gosu user jupyter-lab --version \
  && gosu user marimo --version \
  && gosu user harlequin --version \
  && grep --fixed-strings --quiet '0.0.0.0' /etc/clor/services/jupyterlab.toml \
  && grep --fixed-strings --quiet '8888' /etc/clor/services/jupyterlab.toml \
  && grep --fixed-strings --quiet '2718' /etc/clor/services/marimo.toml \
  && ! find /home/user -maxdepth 2 -type f \
    \( -name credentials -o -name '*.token' \) -print -quit | grep .

FROM data AS production
