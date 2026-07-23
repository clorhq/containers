# shellcheck shell=bash
# Login shells reset PATH in /etc/profile, so restore the space user's tools.
# Root must retain the system-only path established by /etc/profile.
if [[ "$(id --user)" -ne 0 ]]; then
    export PATH="${HOME}/.local/bin:${HOME}/go/bin:${HOME}/.npm-global/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
    export PLAYWRIGHT_BROWSERS_PATH="${HOME}/.cache/ms-playwright"
fi
