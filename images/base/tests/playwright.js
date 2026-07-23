"use strict";

const {
    chromium,
    firefox,
    webkit,
} = require("/home/user/.npm-global/lib/node_modules/@playwright/test");

async function verifyBrowser(name, browserType, headless, launchOptions = {}) {
    const browser = await browserType.launch({ headless, ...launchOptions });
    const page = await browser.newPage();
    await page.setContent("<!doctype html><title>offline</title><main>ready</main>");
    if ((await page.title()) !== "offline") {
        throw new Error(`${name} did not render the offline page`);
    }
    if ((await page.locator("main").textContent()) !== "ready") {
        throw new Error(`${name} returned unexpected page content`);
    }
    await browser.close();
}

async function main() {
    const chromiumExecutablePath =
        process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH;
    if (chromiumExecutablePath) {
        await verifyBrowser("System Chromium", chromium, true, {
            executablePath: chromiumExecutablePath,
        });
        return;
    }

    if (process.env.PLAYWRIGHT_HEADED === "true") {
        await verifyBrowser("Chromium under Xvfb", chromium, false);
        return;
    }

    await verifyBrowser("Chromium", chromium, true);
    await verifyBrowser("Firefox", firefox, true);
    await verifyBrowser("WebKit", webkit, true);
}

main().catch((error) => {
    console.error(error);
    process.exitCode = 1;
});
