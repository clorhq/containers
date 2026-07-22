ARG BASE_IMAGE="software-development-base:latest"

FROM ${BASE_IMAGE}

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

COPY --chmod=0755 images/desktop/services/vnc/run /etc/service/desktop-vnc/run
COPY --chmod=0755 images/desktop/services/web/run /etc/service/desktop-web/run
COPY --chmod=0755 images/desktop/services/vnc/log/run /etc/service/desktop-vnc/log/run
COPY --chmod=0755 images/desktop/services/web/log/run /etc/service/desktop-web/log/run
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

RUN mkdir --parents /var/log/desktop-vnc /var/log/desktop-web \
  && printf 's100000000\nn3\n' >/var/log/desktop-vnc/config \
  && printf 's100000000\nn3\n' >/var/log/desktop-web/config

ENV VNC_GEOMETRY="1920x1080"
EXPOSE 6080

WORKDIR /root
