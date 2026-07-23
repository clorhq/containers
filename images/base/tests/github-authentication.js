"use strict";

const assert = require("assert");
const fs = require("fs");
const os = require("os");
const path = require("path");

const {
  ClorGitHubAuthenticationProvider,
  wrapAuthenticationProvider,
} = require(
  "/usr/local/lib/code-server/4.129.0/lib/vscode/extensions/" +
    "github-authentication/clor-github-provider.js",
);

class EventEmitter {
  constructor() {
    this.listeners = new Set();
    this.event = (listener) => {
      this.listeners.add(listener);
      return { dispose: () => this.listeners.delete(listener) };
    };
  }

  fire(value) {
    for (const listener of this.listeners) {
      listener(value);
    }
  }

  dispose() {
    this.listeners.clear();
  }
}

const vscode = { EventEmitter };

function response(token, expiry, overrides = {}) {
  return JSON.stringify({
    access_token: token,
    account_name: "octocat",
    connection_id: "connection-1",
    expiry: new Date(expiry).toISOString(),
    provider: "github",
    subject: "583231",
    ...overrides,
  });
}

function queuedExecFile(results, calls) {
  return (file, args, options, callback) => {
    calls.push({ file, args, options });
    const result = results.shift();
    setImmediate(() => {
      if (result instanceof Error) {
        callback(result, "", "sensitive diagnostic");
      } else {
        callback(undefined, result, "");
      }
    });
  };
}

async function assertRejectsSanitized(promise) {
  await assert.rejects(promise, (error) => {
    assert.strictEqual(error.message, "Clor GitHub authentication failed");
    assert(!error.message.includes("token"));
    assert(!error.message.includes("sensitive"));
    return true;
  });
}

async function testDelegation() {
  const original = {};
  assert.strictEqual(
    wrapAuthenticationProvider(vscode, "github", original, ""),
    original,
  );
  assert.strictEqual(
    wrapAuthenticationProvider(
      vscode,
      "github-enterprise",
      original,
      "connection-1",
    ),
    original,
  );
}

async function testSessionsRefreshAndExpiry() {
  const calls = [];
  const events = [];
  const timers = [];
  let now = Date.parse("2026-07-23T00:00:00Z");
  const firstExpiry = now + 10 * 60_000;
  const secondExpiry = now + 20 * 60_000;
  const results = [
    response("first-secret-token", firstExpiry),
    response("second-secret-token", secondExpiry),
    new Error("refresh failed with first-secret-token"),
    new Error("expired refresh failed with second-secret-token"),
  ];
  const provider = new ClorGitHubAuthenticationProvider(vscode, {
    connectionId: "connection-1",
    execFile: queuedExecFile(results, calls),
    now: () => now,
    setTimeout: (callback, delay) => {
      const timer = { callback, delay, unref() {} };
      timers.push(timer);
      return timer;
    },
    clearTimeout: () => {},
  });
  provider.onDidChangeSessions((event) => events.push(event));

  const [first] = await provider.getSessions(["repo", "workflow"]);
  assert.strictEqual(first.accessToken, "first-secret-token");
  assert.deepStrictEqual(first.account, { label: "octocat", id: "583231" });
  assert.deepStrictEqual(first.scopes, ["repo", "workflow"]);
  assert(first.id.startsWith("clor-github:connection-1:"));
  assert.strictEqual(calls.length, 1);
  assert.strictEqual(calls[0].file, "clor");
  assert.deepStrictEqual(calls[0].args, [
    "github",
    "auth",
    "--stdout-format",
    "json",
  ]);
  assert.strictEqual(calls[0].options.shell, false);
  assert.strictEqual(process.env.GH_TOKEN, undefined);
  assert.strictEqual(process.env.GITHUB_TOKEN, undefined);
  assert.strictEqual(events.length, 1);
  assert.deepStrictEqual(events[0].added, [first]);
  assert(timers[0].delay > 0 && timers[0].delay < firstExpiry - now);
  assert.deepStrictEqual(
    await provider.getSessions(["repo", "workflow"], {
      account: { label: "octocat", id: "another-subject" },
    }),
    [],
  );

  await provider._refreshInBackground();
  const [second] = await provider.getSessions(["repo", "workflow"]);
  assert.strictEqual(second.accessToken, "second-secret-token");
  assert.strictEqual(second.id, first.id);
  assert.strictEqual(events.length, 2);
  assert.deepStrictEqual(events[1].changed, [second]);

  now = secondExpiry - 30_000;
  const [duringFailure] = await provider.getSessions(["repo", "workflow"]);
  assert.strictEqual(duringFailure.accessToken, "second-secret-token");
  assert.strictEqual(calls.length, 3);
  assert.strictEqual(events.length, 2);

  now = secondExpiry + 1;
  await assertRejectsSanitized(
    provider.getSessions(["repo", "workflow"]),
  );
  assert.strictEqual(calls.length, 4);
  assert.strictEqual(events.length, 3);
  assert.deepStrictEqual(events[2].removed, [duringFailure]);

  provider.dispose();
}

async function testInvalidResponsesFailClosed() {
  const invalidResponses = [
    "not json",
    "null",
    response("secret", Date.now() + 60_000, {
      connection_id: "wrong-connection",
    }),
    response("secret", Date.now() - 60_000),
  ];

  for (const invalidResponse of invalidResponses) {
    const provider = new ClorGitHubAuthenticationProvider(vscode, {
      connectionId: "connection-1",
      execFile: queuedExecFile([invalidResponse], []),
    });
    await assertRejectsSanitized(provider.getSessions(["repo"]));
    provider.dispose();
  }
}

async function testClorStubAndNoPersistence() {
  const testRoot = fs.mkdtempSync(
    path.join(os.tmpdir(), "clor-github-provider-test-"),
  );
  const clorLog = path.join(testRoot, "clor.log");
  const expiry = Date.now() + 10 * 60_000;
  const token = "stub-secret-token";
  fs.writeFileSync(clorLog, "");

  const savedPath = process.env.PATH;
  process.env.PATH = `/usr/local/lib/clor/tests/bin:${savedPath}`;
  process.env.GITHUB_AUTH_PROVIDER_TEST_CLOR_LOG = clorLog;
  process.env.GITHUB_AUTH_PROVIDER_TEST_RESPONSE = response(token, expiry);
  delete process.env.GH_TOKEN;
  delete process.env.GITHUB_TOKEN;

  try {
    const provider = new ClorGitHubAuthenticationProvider(vscode, {
      connectionId: "connection-1",
    });
    const [session] = await provider.getSessions(["repo"]);
    assert.strictEqual(session.accessToken, token);
    assert.strictEqual(
      fs.readFileSync(clorLog, "utf8"),
      "github auth --stdout-format json\n",
    );
    assert(!fs.readFileSync(clorLog, "utf8").includes(token));
    assert.strictEqual(process.env.GH_TOKEN, undefined);
    assert.strictEqual(process.env.GITHUB_TOKEN, undefined);
    provider.dispose();
  } finally {
    process.env.PATH = savedPath;
    delete process.env.GITHUB_AUTH_PROVIDER_TEST_CLOR_LOG;
    delete process.env.GITHUB_AUTH_PROVIDER_TEST_RESPONSE;
    fs.rmSync(testRoot, { recursive: true, force: true });
  }
}

async function main() {
  await testDelegation();
  await testSessionsRefreshAndExpiry();
  await testInvalidResponsesFailClosed();
  await testClorStubAndNoPersistence();
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
