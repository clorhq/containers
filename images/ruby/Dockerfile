ARG BASE_IMAGE="software-development-base:latest"

FROM ${BASE_IMAGE}

USER root

# Build dependencies for compiling Ruby from source with ruby-install.
RUN apt-get update && apt-get install --yes \
  autoconf \
  bison \
  libssl-dev \
  libyaml-dev \
  libreadline-dev \
  zlib1g-dev \
  libncurses-dev \
  libffi-dev \
  libgdbm-dev \
  && rm --recursive --force /var/lib/apt/lists/*

# Install ruby-install, then compile the pinned Ruby into /usr/local.
ARG RUBY_INSTALL_VERSION="0.9.3"
RUN curl --fail --silent --show-error --location \
    "https://github.com/postmodern/ruby-install/releases/download/v${RUBY_INSTALL_VERSION}/ruby-install-${RUBY_INSTALL_VERSION}.tar.gz" \
    --output /tmp/ruby-install.tar.gz \
  && tar --extract --gzip --directory /tmp --file /tmp/ruby-install.tar.gz \
  && make --directory "/tmp/ruby-install-${RUBY_INSTALL_VERSION}" install \
  && rm --recursive --force /tmp/ruby-install.tar.gz "/tmp/ruby-install-${RUBY_INSTALL_VERSION}"

# Version source: https://www.ruby-lang.org/en/downloads/releases/
ARG RUBY_VERSION="4.0.6"
RUN ruby-install --no-install-deps --cleanup --system ruby "${RUBY_VERSION}"

USER user
ENV HOME="/home/user"
ENV PATH="/home/user/.local/bin:/home/user/go/bin:/home/user/.npm-global/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/home/user/.cache/ms-playwright"
WORKDIR /home/user

# Per-user gem home so gem installs land under /home/user, matching the base
# image's per-user tool layout.
ENV GEM_HOME="/home/user/.gem"
ENV PATH="/home/user/.gem/bin:${PATH}"
ENV CLOR_SPACE_PATH_PREFIX="/home/user/.gem/bin"

# Ruby developer tooling: the Bundler dependency manager, the RuboCop
# linter/formatter, and the Solargraph language server.
RUN gem install --no-document bundler rubocop solargraph

USER root
ENV HOME="/root"
ENV PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV PLAYWRIGHT_BROWSERS_PATH="/root/.cache/ms-playwright"
WORKDIR /root
