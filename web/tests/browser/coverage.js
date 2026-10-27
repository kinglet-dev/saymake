// Measures which code of the page's browser-only modules ran, with V8's
// precise block coverage, across all browser tests, so they are held to
// 100% like the Go code (the method follows kinglet.dev's Preflight
// tests). main.mjs is the composition root: the
// tests run it end to end, as main is covered for Go.
const fs = require("fs");
const path = require("path");
const crypto = require("crypto");
const base = require("@playwright/test");

const outDir = path.join(__dirname, "..", "..", "test-results", "js-coverage");
const measured = /\/js\/ui\/(page|panels|view2d|view3d)\.mjs$/;
const expected = ["page.mjs", "panels.mjs", "view2d.mjs", "view3d.mjs"];

const test = base.test.extend({
  page: async ({ page }, use) => {
    await page.coverage.startJSCoverage({ resetOnNavigation: false });
    await use(page);
    const entries = (await page.coverage.stopJSCoverage()).filter((e) => measured.test(new URL(e.url).pathname));
    fs.mkdirSync(outDir, { recursive: true });
    fs.writeFileSync(path.join(outDir, `${crypto.randomUUID()}.json`), JSON.stringify(entries));
  },
});

function executed(entry) {
  const ran = new Array(entry.source.length).fill(true);
  const ranges = entry.functions.flatMap((f) => f.ranges)
    .sort((a, b) => a.startOffset - b.startOffset || b.endOffset - a.endOffset);
  for (const r of ranges) ran.fill(r.count > 0, r.startOffset, r.endOffset);
  return ran;
}

function uncovered() {
  const scripts = new Map();
  for (const file of fs.existsSync(outDir) ? fs.readdirSync(outDir) : []) {
    for (const entry of JSON.parse(fs.readFileSync(path.join(outDir, file), "utf8"))) {
      const name = path.basename(new URL(entry.url).pathname);
      const ran = executed(entry);
      const seen = scripts.get(name);
      scripts.set(name, { source: entry.source, ran: seen ? seen.ran.map((r, i) => r || ran[i]) : ran });
    }
  }
  const missing = [];
  for (const [name, { source, ran }] of scripts) {
    const lines = new Set();
    ran.forEach((r, i) => { if (!r && /\S/.test(source[i])) lines.add(source.slice(0, i).split("\n").length); });
    if (lines.size > 0) missing.push(`${name}: lines ${[...lines].join(", ")} never ran`);
  }
  return { scripts: [...scripts.keys()].sort(), missing };
}

// ready opens the page and waits until the tool has started.
async function ready(page, url = "/") {
  await page.goto(url);
  await base.expect(page.getByRole("status")).toHaveText("Ready. Describe an object to start.", { timeout: 30000 });
}

// describe types a description and makes it.
async function describe(page, text) {
  await page.getByLabel("Description", { exact: true }).fill(text);
  await page.getByRole("button", { name: "Make it" }).click();
}

module.exports = { test, expect: base.expect, uncovered, outDir, expected, ready, describe };
