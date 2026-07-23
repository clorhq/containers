"use strict";

const Module = require("module");
const vscode = require("vscode");
const { wrapAuthenticationProvider } = require("./clor-github-provider");

function loadOriginalExtension() {
  const originalLoad = Module._load;
  const originalRegister =
    vscode.authentication.registerAuthenticationProvider.bind(
      vscode.authentication,
    );
  const authenticationProxy = new Proxy(vscode.authentication, {
    get(target, property, receiver) {
      if (property !== "registerAuthenticationProvider") {
        return Reflect.get(target, property, receiver);
      }

      return (providerId, label, provider, options) => {
        const adaptedProvider = wrapAuthenticationProvider(
          vscode,
          providerId,
          provider,
        );
        const registration = originalRegister(
          providerId,
          label,
          adaptedProvider,
          options,
        );

        if (adaptedProvider === provider) {
          return registration;
        }
        return {
          dispose() {
            adaptedProvider.dispose();
            registration.dispose();
          },
        };
      };
    },
  });
  const vscodeProxy = new Proxy(vscode, {
    get(target, property, receiver) {
      if (property === "authentication") {
        return authenticationProxy;
      }
      return Reflect.get(target, property, receiver);
    },
  });

  Module._load = function clorLoad(request, parent, isMain) {
    if (
      request === "vscode" &&
      parent?.filename?.endsWith("/dist/extension.js")
    ) {
      return vscodeProxy;
    }
    return originalLoad.call(this, request, parent, isMain);
  };

  try {
    return require("./dist/extension.js");
  } finally {
    Module._load = originalLoad;
  }
}

module.exports = loadOriginalExtension();
