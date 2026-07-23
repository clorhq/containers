ARG BASE_IMAGE="software-development-base:latest"

FROM ${BASE_IMAGE} AS desktop

USER root

# Add a lightweight Linux desktop, a loopback-only VNC server, and noVNC's
# browser client and WebSocket proxy.
RUN apt-get update && apt-get install --yes \
  arc-theme \
  dbus-x11 \
  fonts-jetbrains-mono \
  novnc \
  tigervnc-standalone-server \
  tigervnc-tools \
  websockify \
  xfce4 \
  xfce4-terminal \
  && rm --recursive --force /var/lib/apt/lists/*

# Debian's desktop entry bypasses PATH, so point it at the container-safe
# Chromium wrapper explicitly.
RUN test "$(grep --fixed-strings --count \
    'Exec=/usr/bin/chromium %U' /usr/share/applications/chromium.desktop)" = 1 \
  && sed --in-place \
    's#Exec=/usr/bin/chromium %U#Exec=/usr/local/bin/chromium %U#' \
    /usr/share/applications/chromium.desktop \
  && grep --fixed-strings --line-regexp \
    'Exec=/usr/local/bin/chromium %U' /usr/share/applications/chromium.desktop

COPY --chmod=0755 images/desktop/bin/desktop-vnc /usr/local/bin/desktop-vnc
COPY --chmod=0755 images/desktop/bin/desktop-web /usr/local/bin/desktop-web
COPY --chmod=0644 images/desktop/services/desktop-vnc.toml /etc/clor/services/desktop-vnc.toml
COPY --chmod=0644 images/desktop/services/desktop-web.toml /etc/clor/services/desktop-web.toml
COPY --chmod=0755 images/desktop/xstartup /usr/local/lib/clor/desktop-xstartup
COPY images/desktop/index.html /usr/share/novnc/index.html
COPY images/desktop/wallpaper.svg /usr/share/backgrounds/clor/clor-dark.svg
COPY --chown=user:user images/desktop/config/ /home/user/.config/

ARG CLOR_LOGO_URL="https://clor.com/static/clor-logo-inverse.png"
ARG CLOR_LOGO_SHA256="2b4caae93d7f6eb9fddaef1376b54e6e299cf3f797d2803d8d7840314825f373"
# ImageMagick uses single-dash options; render once so XFCE only loads PNGs.
RUN curl --fail --silent --show-error --location \
    --output /usr/share/backgrounds/clor/clor-logo-inverse.png \
    "${CLOR_LOGO_URL}" \
  && logo_checksum="$(sha256sum /usr/share/backgrounds/clor/clor-logo-inverse.png)" \
  && test "${logo_checksum%% *}" = "${CLOR_LOGO_SHA256}" \
  && convert -background none \
    /usr/share/backgrounds/clor/clor-dark.svg \
    /tmp/clor-dark-base.png \
  && convert /usr/share/backgrounds/clor/clor-logo-inverse.png \
    -fuzz 6% -transparent black -resize 400x400 /tmp/clor-logo.png \
  && convert /tmp/clor-dark-base.png /tmp/clor-logo.png \
    -gravity center -compose over -composite \
    /usr/share/backgrounds/clor/clor-dark.png \
  && test "$(identify -format '%wx%h' /usr/share/backgrounds/clor/clor-dark.png)" = \
    "1920x1080" \
  && rm --force /tmp/clor-dark-base.png /tmp/clor-logo.png

ENV VNC_GEOMETRY="1920x1080"
EXPOSE 6080

WORKDIR /root

FROM desktop AS test

COPY --chmod=0755 images/desktop/tests /usr/local/lib/clor/desktop-tests
RUN grep --fixed-strings --line-regexp \
    'Exec=/usr/local/bin/chromium %U' /usr/share/applications/chromium.desktop \
  && bash -n \
    /usr/local/bin/desktop-vnc \
    /usr/local/bin/desktop-web \
    /usr/local/lib/clor/desktop-xstartup \
    /usr/local/lib/clor/desktop-tests/runtime \
  && shellcheck \
    /usr/local/bin/desktop-vnc \
    /usr/local/bin/desktop-web \
    /usr/local/lib/clor/desktop-xstartup \
    /usr/local/lib/clor/desktop-tests/runtime \
  && shfmt -d -i 4 -ci \
    /usr/local/bin/desktop-vnc \
    /usr/local/bin/desktop-web \
    /usr/local/lib/clor/desktop-tests/runtime

FROM desktop AS production
