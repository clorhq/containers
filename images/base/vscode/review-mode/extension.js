const vscode = require("vscode");

async function activate() {
  await vscode.commands.executeCommand("workbench.view.scm");
}

function deactivate() {}

module.exports = { activate, deactivate };
