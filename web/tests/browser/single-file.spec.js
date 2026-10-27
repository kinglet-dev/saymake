// The built page: one file, opened from disk, with its strict CSP.
const path = require("path");
const { test, expect, describe } = require("./coverage");

const url = "file://" + path.join(__dirname, "..", "..", "build", "saymake.html");

test("runs from a double-clicked file under its strict policy, with no requests out", async ({ page }) => {
  // Arrange
  const violations = [], requests = [];
  page.on("console", (m) => { if (/Content Security Policy|Refused/.test(m.text())) violations.push(m.text()); });
  page.on("request", (r) => { if (!/^(file|data|blob):/.test(r.url())) requests.push(r.url()); });
  await page.goto(url);
  await page.evaluate(() => document.addEventListener("securitypolicyviolation", (e) => console.log(`Refused ${e.violatedDirective}`)));
  await expect(page.getByRole("status")).toHaveText("Ready. Describe an object to start.", { timeout: 30000 });
  const waiting = page.waitForEvent("download");

  // Act
  await describe(page, "a tall fluted vase, 20 cm");
  await page.getByRole("button", { name: "Download STL (.stl)" }).click();

  // Assert
  expect((await waiting).suggestedFilename()).toBe("vase.stl");
  await expect(page.getByRole("img", { name: /3D preview of the vase/ })).toBeVisible();
  expect(violations).toEqual([]);
  expect(requests).toEqual([]);
  const csp = await page.locator('meta[http-equiv="Content-Security-Policy"]').getAttribute("content");
  expect(csp).toContain("connect-src 'none'");
  expect(csp).not.toContain("unsafe-inline");
  expect(await page.locator("link[rel=stylesheet], script[src]").count()).toBe(0);
  await expect(page).toHaveTitle("Kinglet Saymake: describe it, then print or cut it");
});
