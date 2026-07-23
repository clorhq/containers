"use strict";

const fs = require("fs");
const vscode = require("vscode");

async function activate() {
  const scopes = ["read:user", "user:email", "repo", "workflow"];
  const expectedToken = process.env.AUTH_PROBE_EXPECTED_TOKEN;
  fs.writeFileSync(
    `${process.env.AUTH_PROBE_RESULT}.started`,
    JSON.stringify({
      clorExecutables: process.env.PATH.split(":").filter((directory) =>
        fs.existsSync(`${directory}/clor`),
      ),
      hasClorResponse: Boolean(
        process.env.GITHUB_AUTH_PROVIDER_TEST_RESPONSE,
      ),
    }),
  );
  const session = await vscode.authentication.getSession(
    "github",
    scopes,
    { silent: true },
  );

  if (!expectedToken) {
    if (session) {
      throw new Error("normal GitHub authentication unexpectedly had a session");
    }
    fs.writeFileSync(process.env.AUTH_PROBE_RESULT, "delegated\n");
    return;
  }

  if (
    !session ||
    session.accessToken !== expectedToken ||
    session.account.label !== "octocat" ||
    session.account.id !== "583231" ||
    JSON.stringify(session.scopes) !== JSON.stringify(scopes)
  ) {
    throw new Error("Clor returned an unexpected GitHub session");
  }
  fs.writeFileSync(
    process.env.AUTH_PROBE_RESULT,
    JSON.stringify({
      account: session.account,
      scopes: session.scopes,
      status: "authenticated",
    }),
  );
}

function deactivate() {}

module.exports = { activate, deactivate };
