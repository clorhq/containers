const vscode = require("vscode");

async function activate() {
  await vscode.commands.executeCommand("workbench.view.scm");
  try {
    await vscode.authentication.getSession(
      "github",
      ["read:user", "user:email", "repo", "workflow"],
      { silent: true },
    );
  } catch (error) {
    console.warn("Clor GitHub authentication unavailable", error);
  }
}

function deactivate() {}

module.exports = { activate, deactivate };
