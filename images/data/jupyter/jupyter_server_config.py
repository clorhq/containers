# Jupyter Server configuration for a Clor space.
#
# A space reaches JupyterLab through the Clor tunnel: the browser loads the
# public "*.clor.host" name inside an iframe on clor.com, and the space daemon
# reverse-proxies the request to the server over loopback. Every default below
# assumes a browser talking to Jupyter directly, and each one breaks under that
# indirection. Access is gated by the unguessable tunnel hostname and by the
# Clor edge's own "frame-ancestors" header, which this file does not touch.
#
# Placement matters. images/data/services/jupyterlab.toml sets
# JUPYTER_PLATFORM_DIRS=1, which makes jupyter_core.paths drop SYSTEM_CONFIG_PATH
# (/usr/local/etc/jupyter, /etc/jupyter) in favour of platformdirs'
# /etc/xdg/jupyter -- so a copy in /etc/jupyter would be silently ignored by the
# supervised service. ENV_CONFIG_PATH (sys.prefix/etc/jupyter) is searched with
# and without that variable, and jupyter-lab runs from the data venv, so
# installing here is the only placement that covers both the service and a
# "jupyter lab" a user starts by hand from the Shell tab.

c = get_config()  # noqa: F821

# JupyterHandler.content_security_policy returns this header verbatim when it is
# present, replacing the built-in "frame-ancestors 'self'" that stops the tab
# from rendering. The edge still sends its own frame-ancestors header, and a
# browser enforces every policy it receives, so the effective policy is still
# the edge's allowlist -- this only stops Jupyter from vetoing it.
c.ServerApp.tornado_settings = {
    "headers": {"Content-Security-Policy": "frame-ancestors *"},
}

# The tunnel rewrites Host to 127.0.0.1:<port>, so check_origin()'s default
# origin-equals-host comparison rejects every websocket the public name opens,
# which is every kernel and terminal.
c.ServerApp.allow_origin = "*"

# The _xsrf cookie is SameSite=Lax, so the browser withholds it from a
# cross-site iframe and every POST (starting a kernel, saving a notebook) would
# 403 with no way for the page to recover.
c.ServerApp.disable_check_xsrf = True

# Stated rather than inferred: the proxied Host header is loopback today, but
# remote access must not depend on that staying true.
c.ServerApp.allow_remote_access = True

# The tunnel terminates TLS, so the real scheme and client address only survive
# in X-Forwarded-Proto and X-Forwarded-For.
c.ServerApp.trust_xheaders = True

# set_login_cookie() infers "secure" from request.protocol, which is plain http
# on the proxied hop; without Secure the browser rejects SameSite=None outright.
c.IdentityProvider.secure_cookie = True

# Tornado 6 set_cookie kwargs. SameSite=None is what lets the identity cookie be
# sent at all from inside the cross-site iframe.
c.IdentityProvider.cookie_options = {"samesite": "None", "secure": True}

# Deliberately not set: ServerApp.allow_credentials. Paired with the
# "Access-Control-Allow-Origin: *" that allow_origin produces it is a
# combination browsers reject, and the tab's own requests are same-origin, so
# CORS credentials never enter into it.
