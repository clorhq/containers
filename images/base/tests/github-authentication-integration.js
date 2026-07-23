"use strict";

const assert = require("assert");
const childProcess = require("child_process");
const fs = require("fs");
const http = require("http");
const os = require("os");
const path = require("path");
const { once } = require("events");
const {
  chromium,
} = require("/home/user/.npm-global/lib/node_modules/@playwright/test");

const probeSource = "/usr/local/lib/clor/tests/vscode-auth-probe";
const stubDirectory = "/usr/local/lib/clor/tests/bin";

function waitForHttp(port, deadline) {
  return new Promise((resolve, reject) => {
    const attempt = () => {
      const request = http.get(
        { host: "127.0.0.1", port, path: "/" },
        (response) => {
          response.resume();
          resolve();
        },
      );
      request.on("error", (error) => {
        if (Date.now() >= deadline) {
          reject(error);
        } else {
          setTimeout(attempt, 100);
        }
      });
    };
    attempt();
  });
}

function collectDiagnostics(directory, token) {
  const diagnostics = [];
  const visit = (entry) => {
    for (const child of fs.readdirSync(entry, { withFileTypes: true })) {
      const childPath = path.join(entry, child.name);
      if (child.isDirectory()) {
        visit(childPath);
      } else if (child.isFile() && child.name.endsWith(".log")) {
        const contents = fs
          .readFileSync(childPath, "utf8")
          .replaceAll(token, "<redacted>");
        diagnostics.push(`FILE ${childPath}\n${contents}`);
      }
    }
  };
  visit(directory);
  return diagnostics.join("\n");
}

function filesContaining(directory, value) {
  const matches = [];
  const needle = Buffer.from(value);
  const visit = (entry) => {
    for (const child of fs.readdirSync(entry, { withFileTypes: true })) {
      const childPath = path.join(entry, child.name);
      if (child.isDirectory()) {
        visit(childPath);
      } else if (child.isFile()) {
        try {
          if (fs.readFileSync(childPath).includes(needle)) {
            matches.push(childPath);
          }
        } catch (error) {
          if (error.code !== "ENOENT") {
            throw error;
          }
        }
      }
    }
  };
  visit(directory);
  return matches;
}

async function waitForFile(file, server, deadline) {
  while (Date.now() < deadline) {
    if (fs.existsSync(file)) {
      return;
    }
    if (server.exitCode !== null) {
      throw new Error(`code-server exited with status ${server.exitCode}`);
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
  throw new Error(`timed out waiting for ${file}`);
}

async function stopServer(server) {
  if (server.exitCode !== null) {
    return;
  }
  server.kill("SIGTERM");
  await Promise.race([
    once(server, "exit"),
    new Promise((resolve) => setTimeout(resolve, 5_000)),
  ]);
  if (server.exitCode === null) {
    server.kill("SIGKILL");
    await once(server, "exit");
  }
}

async function runProbe(configured, port) {
  const testRoot = fs.mkdtempSync(
    path.join(os.tmpdir(), "clor-github-extension-host-"),
  );
  const extensionsDirectory = path.join(testRoot, "extensions");
  const userDataDirectory = path.join(testRoot, "user-data");
  const workspace = path.join(testRoot, "workspace");
  const resultFile = path.join(testRoot, "result");
  const clorLog = path.join(testRoot, "clor.log");
  const serverLog = path.join(testRoot, "code-server.log");
  const token = "extension-host-secret-token";
  const extensionDirectory = path.join(
    extensionsDirectory,
    "github.vscode-pull-request-github-1.0.0",
  );

  fs.mkdirSync(extensionsDirectory, { recursive: true });
  fs.mkdirSync(userDataDirectory, { recursive: true });
  fs.mkdirSync(workspace, { recursive: true });
  fs.cpSync(probeSource, extensionDirectory, { recursive: true });
  fs.writeFileSync(clorLog, "");
  const logDescriptor = fs.openSync(serverLog, "w");
  const environment = {
    ...process.env,
    HOME: path.join(testRoot, "home"),
    PATH: `${stubDirectory}:${process.env.PATH}`,
    AUTH_PROBE_RESULT: resultFile,
    GITHUB_AUTH_PROVIDER_TEST_CLOR_LOG: clorLog,
  };

  if (configured) {
    environment.CLOR_GITHUB_CONNECTION_ID = "connection-1";
    environment.AUTH_PROBE_EXPECTED_TOKEN = token;
    environment.GITHUB_AUTH_PROVIDER_TEST_RESPONSE = JSON.stringify({
      access_token: token,
      account_name: "octocat",
      connection_id: "connection-1",
      expiry: new Date(Date.now() + 10 * 60_000).toISOString(),
      provider: "github",
      subject: "583231",
    });
    const directResponse = childProcess.execFileSync(
      "clor",
      ["github", "auth", "--stdout-format", "json"],
      { encoding: "utf8", env: environment, shell: false },
    );
    assert.strictEqual(JSON.parse(directResponse).access_token, token);
    fs.writeFileSync(clorLog, "");
  } else {
    delete environment.CLOR_GITHUB_CONNECTION_ID;
    delete environment.AUTH_PROBE_EXPECTED_TOKEN;
    delete environment.GITHUB_AUTH_PROVIDER_TEST_RESPONSE;
  }

  fs.mkdirSync(environment.HOME, { recursive: true });
  const homeBin = path.join(environment.HOME, ".local", "bin");
  fs.mkdirSync(homeBin, { recursive: true });
  fs.copyFileSync(path.join(stubDirectory, "clor"), path.join(homeBin, "clor"));
  fs.chmodSync(path.join(homeBin, "clor"), 0o755);
  const server = childProcess.spawn(
    "code-server",
    [
      "--config",
      "/dev/null",
      "--auth",
      "none",
      "--disable-telemetry",
      "--disable-workspace-trust",
      "--disable-getting-started-override",
      "--bind-addr",
      `127.0.0.1:${port}`,
      "--extensions-dir",
      extensionsDirectory,
      "--user-data-dir",
      userDataDirectory,
      workspace,
    ],
    {
      env: environment,
      stdio: ["ignore", logDescriptor, logDescriptor],
    },
  );
  fs.closeSync(logDescriptor);

  let browser;
  try {
    const deadline = Date.now() + 30_000;
    await waitForHttp(port, deadline);
    browser = await chromium.launch({ headless: true });
    const page = await browser.newPage();
    await page.goto(`http://127.0.0.1:${port}`, {
      waitUntil: "domcontentloaded",
      timeout: 30_000,
    });
    await waitForFile(resultFile, server, deadline);

    const result = fs.readFileSync(resultFile, "utf8");
    const commandLog = fs.readFileSync(clorLog, "utf8");
    const codeServerLog = fs.readFileSync(serverLog, "utf8");
    assert(!result.includes(token));
    assert(!commandLog.includes(token));
    assert(!codeServerLog.includes(token));

    if (configured) {
      assert.deepStrictEqual(JSON.parse(result), {
        account: { label: "octocat", id: "583231" },
        scopes: ["read:user", "user:email", "repo", "workflow"],
        status: "authenticated",
      });
      assert.strictEqual(
        commandLog,
        "github auth --stdout-format json\n",
      );
    } else {
      assert.strictEqual(result, "delegated\n");
      assert.strictEqual(commandLog, "");
    }
    assert.deepStrictEqual(filesContaining(testRoot, token), []);
  } catch (error) {
    const startedFile = `${resultFile}.started`;
    if (fs.existsSync(startedFile)) {
      error.message += `\nPROBE ${fs.readFileSync(startedFile, "utf8")}`;
    }
    error.message += `\n${collectDiagnostics(testRoot, token)}`;
    throw error;
  } finally {
    await browser?.close();
    await stopServer(server);
    fs.rmSync(testRoot, { recursive: true, force: true });
  }
}

async function main() {
  await runProbe(true, 18_080);
  await runProbe(false, 18_081);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
