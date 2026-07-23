"use strict";

const childProcess = require("child_process");
const crypto = require("crypto");

const REFRESH_SKEW_MS = 60_000;
const REFRESH_RETRY_MS = 15_000;
const MAX_TIMER_DELAY_MS = 2_147_483_647;

class ClorGitHubAuthenticationProvider {
  constructor(vscode, options = {}) {
    this._vscode = vscode;
    this._connectionId =
      options.connectionId ?? process.env.CLOR_GITHUB_CONNECTION_ID;
    this._execFile = options.execFile ?? childProcess.execFile;
    this._now = options.now ?? Date.now;
    this._setTimeout = options.setTimeout ?? setTimeout;
    this._clearTimeout = options.clearTimeout ?? clearTimeout;
    this._credential = undefined;
    this._refreshPromise = undefined;
    this._refreshTimer = undefined;
    this._sessions = new Map();
    this._sessionChangeEmitter = new vscode.EventEmitter();
    this.onDidChangeSessions = this._sessionChangeEmitter.event;
  }

  async getSessions(scopes = [], options = {}) {
    const credential = await this._ensureCredential();
    const session = this._sessionForScopes(scopes, credential);

    if (
      options.account &&
      (options.account.label !== session.account.label ||
        options.account.id !== session.account.id)
    ) {
      return [];
    }
    return [session];
  }

  async createSession(scopes = []) {
    const credential = await this._ensureCredential();
    return this._sessionForScopes(scopes, credential);
  }

  async removeSession(sessionId) {
    for (const [scopeKey, session] of this._sessions) {
      if (session.id === sessionId) {
        this._sessions.delete(scopeKey);
        this._sessionChangeEmitter.fire({
          added: [],
          removed: [session],
          changed: [],
        });
        return;
      }
    }
  }

  dispose() {
    if (this._refreshTimer !== undefined) {
      this._clearTimeout(this._refreshTimer);
      this._refreshTimer = undefined;
    }
    this._credential = undefined;
    this._sessions.clear();
    this._sessionChangeEmitter.dispose();
  }

  async _ensureCredential() {
    const now = this._now();
    if (
      this._credential &&
      now < this._credential.refreshAt
    ) {
      return this._credential;
    }

    if (this._credential && now >= this._credential.expiresAt) {
      this._expireCredential();
    }

    try {
      return await this._refreshCredential();
    } catch {
      if (this._credential && this._now() < this._credential.expiresAt) {
        this._scheduleRefresh(REFRESH_RETRY_MS);
        return this._credential;
      }
      this._expireCredential();
      throw new Error("Clor GitHub authentication failed");
    }
  }

  _refreshCredential() {
    if (!this._refreshPromise) {
      this._refreshPromise = this._requestCredential()
        .then((credential) => {
          this._installCredential(credential);
          return credential;
        })
        .finally(() => {
          this._refreshPromise = undefined;
        });
    }
    return this._refreshPromise;
  }

  _requestCredential() {
    return new Promise((resolve, reject) => {
      this._execFile(
        "clor",
        ["github", "auth", "--stdout-format", "json"],
        {
          encoding: "utf8",
          maxBuffer: 1024 * 1024,
          shell: false,
          windowsHide: true,
        },
        (error, stdout) => {
          if (error) {
            reject(new Error("Clor GitHub authentication command failed"));
            return;
          }

          let response;
          try {
            response = JSON.parse(stdout);
          } catch {
            reject(new Error("Clor GitHub authentication returned invalid JSON"));
            return;
          }

          if (
            !response ||
            typeof response !== "object" ||
            typeof response.access_token !== "string" ||
            response.access_token.length === 0 ||
            typeof response.account_name !== "string" ||
            response.account_name.length === 0 ||
            typeof response.connection_id !== "string" ||
            response.connection_id !== this._connectionId ||
            response.provider !== "github" ||
            typeof response.subject !== "string" ||
            response.subject.length === 0
          ) {
            reject(
              new Error("Clor GitHub authentication returned invalid credentials"),
            );
            return;
          }

          const expiry = Date.parse(response.expiry);
          if (!Number.isFinite(expiry) || expiry <= this._now()) {
            reject(
              new Error("Clor GitHub authentication returned invalid credentials"),
            );
            return;
          }

          resolve({
            accessToken: response.access_token,
            accountName: response.account_name,
            connectionId: response.connection_id,
            expiresAt: expiry,
            subject: response.subject,
          });
        },
      );
    });
  }

  _installCredential(credential) {
    const changed = [];
    const previousCredential = this._credential;
    const now = this._now();
    const remaining = Math.max(0, credential.expiresAt - now);
    credential.refreshAt =
      now + Math.max(1_000, remaining - this._refreshSkew(remaining));
    this._credential = credential;

    for (const [scopeKey, session] of this._sessions) {
      const replacement = this._createSession(session.scopes, credential);
      this._sessions.set(scopeKey, replacement);
      changed.push(replacement);
    }

    this._scheduleRefresh();
    if (previousCredential && changed.length > 0) {
      this._sessionChangeEmitter.fire({
        added: [],
        removed: [],
        changed,
      });
    }
  }

  _sessionForScopes(scopes, credential) {
    const requestedScopes = [...scopes];
    const scopeKey = JSON.stringify([...new Set(requestedScopes)].sort());
    let session = this._sessions.get(scopeKey);

    if (
      !session ||
      session.accessToken !== credential.accessToken ||
      session.account.id !== credential.subject
    ) {
      session = this._createSession(requestedScopes, credential);
      this._sessions.set(scopeKey, session);
      this._sessionChangeEmitter.fire({
        added: [session],
        removed: [],
        changed: [],
      });
    }
    return session;
  }

  _createSession(scopes, credential) {
    const scopeKey = JSON.stringify([...new Set(scopes)].sort());
    const scopeDigest = crypto
      .createHash("sha256")
      .update(scopeKey)
      .digest("hex")
      .slice(0, 16);

    return {
      id: `clor-github:${credential.connectionId}:${scopeDigest}`,
      accessToken: credential.accessToken,
      account: {
        label: credential.accountName,
        id: credential.subject,
      },
      scopes: [...scopes],
    };
  }

  _refreshSkew(remaining) {
    return Math.min(REFRESH_SKEW_MS, Math.max(1_000, remaining / 10));
  }

  _scheduleRefresh(delay) {
    if (this._refreshTimer !== undefined) {
      this._clearTimeout(this._refreshTimer);
    }

    const refreshDelay =
      delay === undefined
        ? Math.max(1_000, this._credential.refreshAt - this._now())
        : Math.max(
            1_000,
            Math.min(delay, this._credential.expiresAt - this._now()),
          );
    this._refreshTimer = this._setTimeout(() => {
      this._refreshTimer = undefined;
      void this._refreshInBackground();
    }, Math.min(refreshDelay, MAX_TIMER_DELAY_MS));
    this._refreshTimer.unref?.();
  }

  async _refreshInBackground() {
    try {
      await this._refreshCredential();
    } catch {
      if (this._credential && this._now() < this._credential.expiresAt) {
        this._scheduleRefresh(REFRESH_RETRY_MS);
      } else {
        this._expireCredential();
      }
    }
  }

  _expireCredential() {
    if (this._refreshTimer !== undefined) {
      this._clearTimeout(this._refreshTimer);
      this._refreshTimer = undefined;
    }
    this._credential = undefined;

    if (this._sessions.size > 0) {
      const removed = [...this._sessions.values()];
      this._sessions.clear();
      this._sessionChangeEmitter.fire({
        added: [],
        removed,
        changed: [],
      });
    }
  }
}

function wrapAuthenticationProvider(
  vscode,
  providerId,
  provider,
  connectionId = process.env.CLOR_GITHUB_CONNECTION_ID,
  options = {},
) {
  if (providerId !== "github" || !connectionId) {
    return provider;
  }
  return new ClorGitHubAuthenticationProvider(vscode, {
    ...options,
    connectionId,
  });
}

module.exports = {
  ClorGitHubAuthenticationProvider,
  wrapAuthenticationProvider,
};
