ARG BASE_IMAGE="software-development-base:latest"

FROM ${BASE_IMAGE} AS devops

USER root
ENV HOME="/root"
ENV PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

# Official Azure CLI packages are signed and published for both Debian
# architectures. Pin the exact Bookworm package rather than following latest.
ARG AZURE_CLI_VERSION="2.88.0"
RUN install --directory --mode=0755 /etc/apt/keyrings \
  && curl --fail --silent --show-error --location \
    https://packages.microsoft.com/keys/microsoft.asc \
    | gpg --dearmor --output /etc/apt/keyrings/microsoft.gpg \
  && chmod go+r /etc/apt/keyrings/microsoft.gpg \
  && printf '%s\n' \
    'Types: deb' \
    'URIs: https://packages.microsoft.com/repos/azure-cli/' \
    'Suites: bookworm' \
    'Components: main' \
    "Architectures: $(dpkg --print-architecture)" \
    'Signed-by: /etc/apt/keyrings/microsoft.gpg' \
    >/etc/apt/sources.list.d/azure-cli.sources \
  && apt-get update \
  && apt-get install --yes \
    "azure-cli=${AZURE_CLI_VERSION}-1~bookworm" \
    age \
  && rm --recursive --force /var/lib/apt/lists/*

# Provider-owned Node CLIs live in an isolated prefix. The PATH-facing
# wrappers prefer node_modules/.bin from the repository and fall back here.
ARG VERCEL_VERSION="57.0.0"
ARG WRANGLER_VERSION="4.114.0"
ARG NETLIFY_VERSION="27.0.0"
ARG HEROKU_VERSION="11.8.1"
ARG FIREBASE_VERSION="15.24.0"
RUN npm install --global --prefix /opt/clor/deploy-node \
    "vercel@${VERCEL_VERSION}" \
    "wrangler@${WRANGLER_VERSION}" \
    "netlify-cli@${NETLIFY_VERSION}" \
    "heroku@${HEROKU_VERSION}" \
    "firebase-tools@${FIREBASE_VERSION}" \
  && npm cache clean --force
COPY --chmod=0755 images/devops/clor-deploy-node-tool /usr/local/lib/clor/clor-deploy-node-tool
RUN for tool in vercel wrangler netlify heroku firebase; do \
    ln --symbolic /usr/local/lib/clor/clor-deploy-node-tool "/usr/local/bin/${tool}"; \
  done

# Official architecture-specific AWS and Google Cloud distributions.
ARG AWS_CLI_VERSION="2.36.8"
ARG GCLOUD_VERSION="577.0.0"
RUN case "$(dpkg --print-architecture)" in \
    amd64) AWS_ARCH=x86_64; AWS_SHA256=829a2bbad5b4821a8d006a133c34fa5ce2f362ae003ac39aa0865a6d1e1711f2; GCLOUD_ARCH=x86_64; GCLOUD_SHA256=0b32d330446ce7b0f57f253e7efab4636c18fb1f87a3ac31c6c3f2a2a697525e ;; \
    arm64) AWS_ARCH=aarch64; AWS_SHA256=fedb42aaa0f4e9623c626008139a0e4741cfa3f7ea3067c5a479390126e09f58; GCLOUD_ARCH=arm; GCLOUD_SHA256=dbac26bdf80d72b5d13538e3a215dcbfe2781edfd2d69723effbeef3839cffb8 ;; \
    *) exit 1 ;; \
    esac \
  && curl --fail --silent --show-error --location \
    --output /tmp/awscliv2.zip \
    "https://awscli.amazonaws.com/awscli-exe-linux-${AWS_ARCH}-${AWS_CLI_VERSION}.zip" \
  && printf '%s  %s\n' "${AWS_SHA256}" /tmp/awscliv2.zip | sha256sum --check --strict \
  && unzip -q /tmp/awscliv2.zip -d /tmp/aws \
  && /tmp/aws/aws/install \
  && curl --fail --silent --show-error --location \
    --output /tmp/google-cloud-cli.tar.gz \
    "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-${GCLOUD_VERSION}-linux-${GCLOUD_ARCH}.tar.gz" \
  && printf '%s  %s\n' "${GCLOUD_SHA256}" /tmp/google-cloud-cli.tar.gz | sha256sum --check --strict \
  && tar --extract --gzip --directory=/opt --file=/tmp/google-cloud-cli.tar.gz \
  && /opt/google-cloud-sdk/install.sh --quiet \
    --usage-reporting=false --path-update=false --command-completion=false \
  && ln --symbolic /opt/google-cloud-sdk/bin/gcloud /usr/local/bin/gcloud \
  && rm --recursive --force /tmp/aws /tmp/awscliv2.zip /tmp/google-cloud-cli.tar.gz

# Provider-owned release artifacts for web-platform CLIs. These avoid source
# builds that can reinterpret the providers' module-level replace directives,
# and GitHub's published digest is verified for every selected asset.
ARG FLY_VERSION="0.4.74"
ARG RAILWAY_VERSION="5.28.1"
ARG RENDER_VERSION="2.22.0"
ARG DOCTL_VERSION="1.164.0"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=secret,id=github_token,required=false \
  --mount=type=cache,target=/downloads,sharing=locked \
  fly_archive="$("/installers/verified-github-asset" superfly/flyctl "v${FLY_VERSION}" \
    "flyctl_${FLY_VERSION}_Linux_x86_64.tar.gz" \
    "flyctl_${FLY_VERSION}_Linux_arm64.tar.gz")" \
  && tar --extract --gzip --directory=/tmp --file="${fly_archive}" \
  && install --mode=0755 /tmp/flyctl /usr/local/bin/flyctl \
  && ln --symbolic /usr/local/bin/flyctl /usr/local/bin/fly \
  && render_archive="$("/installers/verified-github-asset" render-oss/cli "v${RENDER_VERSION}" \
    "cli_${RENDER_VERSION}_linux_amd64.zip" \
    "cli_${RENDER_VERSION}_linux_arm64.zip")" \
  && mkdir --parents /tmp/render \
  && unzip -q "${render_archive}" -d /tmp/render \
  && install --mode=0755 "/tmp/render/cli_v${RENDER_VERSION}" /usr/local/bin/render \
  && doctl_archive="$("/installers/verified-github-asset" digitalocean/doctl "v${DOCTL_VERSION}" \
    "doctl-${DOCTL_VERSION}-linux-amd64.tar.gz" \
    "doctl-${DOCTL_VERSION}-linux-arm64.tar.gz")" \
  && tar --extract --gzip --directory=/tmp --file="${doctl_archive}" \
  && install --mode=0755 /tmp/doctl /usr/local/bin/doctl \
  && cargo-binstall --no-confirm --force "railwayapp@${RAILWAY_VERSION}" \
  && rm --recursive --force /tmp/doctl /tmp/flyctl /tmp/render

# Kubernetes, infrastructure, secret, signing, and registry tooling.
ARG KUBECTL_VERSION="1.36.3"
RUN case "$(dpkg --print-architecture)" in \
    amd64) ARCH=amd64 ;; arm64) ARCH=arm64 ;; *) exit 1 ;; esac \
  && curl --fail --silent --show-error --location \
    --output /tmp/kubectl "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl" \
  && curl --fail --silent --show-error --location \
    --output /tmp/kubectl.sha256 "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl.sha256" \
  && printf '%s  %s\n' "$(cat /tmp/kubectl.sha256)" /tmp/kubectl | sha256sum --check --strict \
  && install --mode=0755 /tmp/kubectl /usr/local/bin/kubectl \
  && rm /tmp/kubectl /tmp/kubectl.sha256

ARG HELM_VERSION="4.2.3"
ARG KUSTOMIZE_VERSION="5.8.1"
ARG K9S_VERSION="0.51.0"
ARG STERN_VERSION="1.34.0"
ARG OPENTOFU_VERSION="1.12.5"
ARG SOPS_VERSION="3.13.3"
ARG AGE_VERSION="1.3.1"
ARG COSIGN_VERSION="3.1.2"
ARG CRANE_VERSION="0.21.7"
ARG ORAS_VERSION="1.3.3"
# The module and build caches are BuildKit cache mounts, not image content:
# these ten builds write about three gigabytes into them for roughly a hundred
# megabytes of binaries, and none of it is needed to run them.
RUN --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
  --mount=type=cache,target=/root/go/pkg,sharing=locked \
  export GOBIN=/usr/local/bin \
  && go install "helm.sh/helm/v4/cmd/helm@v${HELM_VERSION}" \
  && go install "sigs.k8s.io/kustomize/kustomize/v5@v${KUSTOMIZE_VERSION}" \
  && go install "github.com/derailed/k9s@v${K9S_VERSION}" \
  && go install "github.com/stern/stern@v${STERN_VERSION}" \
  && go install "github.com/getsops/sops/v3/cmd/sops@v${SOPS_VERSION}" \
  && go install "filippo.io/age/cmd/age@v${AGE_VERSION}" \
  && go install "filippo.io/age/cmd/age-keygen@v${AGE_VERSION}" \
  && go install "github.com/sigstore/cosign/v3/cmd/cosign@v${COSIGN_VERSION}" \
  && go install "github.com/google/go-containerregistry/cmd/crane@v${CRANE_VERSION}" \
  && go install "oras.land/oras/cmd/oras@v${ORAS_VERSION}"

# OpenTofu's module uses replace directives and is intentionally not
# installable through `go install module@version`. Use its provider release
# artifact and independently verify GitHub's digest for the exact asset.
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=secret,id=github_token,required=false \
  --mount=type=cache,target=/downloads,sharing=locked \
  tofu_archive="$("/installers/verified-github-asset" opentofu/opentofu "v${OPENTOFU_VERSION}" \
    "tofu_${OPENTOFU_VERSION}_linux_amd64.zip" \
    "tofu_${OPENTOFU_VERSION}_linux_arm64.zip")" \
  && mkdir --parents /tmp/tofu \
  && unzip -q "${tofu_archive}" -d /tmp/tofu \
  && install --mode=0755 /tmp/tofu/tofu /usr/local/bin/tofu \
  && rm --recursive --force /tmp/tofu

ARG PULUMI_VERSION="3.254.0"
RUN --mount=type=bind,source=installers,target=/installers,readonly \
  --mount=type=secret,id=github_token,required=false \
  --mount=type=cache,target=/downloads,sharing=locked \
  pulumi_archive="$("/installers/verified-github-asset" pulumi/pulumi "v${PULUMI_VERSION}" \
    "pulumi-v${PULUMI_VERSION}-linux-x64.tar.gz" \
    "pulumi-v${PULUMI_VERSION}-linux-arm64.tar.gz")" \
  && tar --extract --gzip --directory=/tmp --file="${pulumi_archive}" \
  && install --mode=0755 /tmp/pulumi/pulumi /usr/local/bin/pulumi \
  && install --mode=0755 /tmp/pulumi/pulumi-language-* /usr/local/bin/ \
  && rm --recursive --force /tmp/pulumi

ARG ANSIBLE_CORE_VERSION="2.21.2"
RUN uv tool install "ansible-core==${ANSIBLE_CORE_VERSION}" \
  && for tool in /root/.local/bin/ansible*; do \
    ln --symbolic "${tool}" "/usr/local/bin/$(basename "${tool}")"; \
  done

COPY --chmod=0755 images/devops/clor-deploy /usr/local/bin/clor-deploy
ENV AZURE_CORE_COLLECT_TELEMETRY="false"
ENV CLOUDSDK_CORE_DISABLE_USAGE_REPORTING="true"
ENV FIREBASE_CLI_DISABLE_UPDATE_CHECK="true"
ENV NETLIFY_TELEMETRY_DISABLED="1"
ENV PULUMI_SKIP_UPDATE_CHECK="true"
ENV VERCEL_TELEMETRY_DISABLED="1"
ENV WRANGLER_SEND_METRICS="false"
WORKDIR /root

FROM devops AS test
COPY --chmod=0755 images/devops/tests /usr/local/lib/clor/devops-tests
RUN command -v \
    age age-keygen ansible aws az cosign crane doctl firebase fly gcloud helm \
    heroku k9s kubectl kustomize netlify oras pulumi railway render sops stern \
    tofu vercel wrangler >/dev/null \
  && for tool in age age-keygen cosign crane helm k9s kustomize oras sops \
    stern; do \
    test "$(command -v "${tool}")" = "/usr/local/bin/${tool}" || exit 1; \
  done \
  && test -z "$(find /root/go /root/.cache/go-build /home/user/go \
    /home/user/.cache/go-build -type f -print -quit 2>/dev/null)" \
  && bash -n /usr/local/bin/clor-deploy \
    /usr/local/lib/clor/clor-deploy-node-tool \
    /usr/local/lib/clor/devops-tests/clor-deploy-node-tool \
  && shellcheck /usr/local/bin/clor-deploy \
    /usr/local/lib/clor/clor-deploy-node-tool \
    /usr/local/lib/clor/devops-tests/clor-deploy-node-tool \
  && shfmt -d -i 4 -ci /usr/local/bin/clor-deploy \
    /usr/local/lib/clor/clor-deploy-node-tool \
    /usr/local/lib/clor/devops-tests/clor-deploy-node-tool \
  && /usr/local/lib/clor/devops-tests/clor-deploy-node-tool \
  && ! find /root /home/user -maxdepth 3 -type f \
    \( -name credentials -o -name credentials.json -o -name config.json \
    -o -name '*.token' \) -print -quit | grep .

FROM devops AS production
