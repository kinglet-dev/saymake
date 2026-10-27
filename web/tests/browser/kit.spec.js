// The kit kinglet.dev embeds (build/kit), in a host page standing in for the
// site (tests/fixtures/host.html): the site's policy, its own header and
// stylesheet, and rules with class names like the tool's own.
const { test, expect, ready, describe } = require("./coverage");

test("works inside another page under the site's policy, fetching only from it", async ({ page }) => {
  // Arrange
  const violations = [], outside = [];
  page.on("console", (m) => { if (/Content Security Policy|Refused/.test(m.text())) violations.push(m.text()); });
  page.on("request", (r) => { if (!r.url().startsWith("http://127.0.0.1:4377/kit/") && !/^(data|blob):/.test(r.url())) outside.push(r.url()); });
  await ready(page, "/kit/host.html");
  const waiting = page.waitForEvent("download");

  // Act
  await describe(page, "an oblique vase");
  await page.getByRole("button", { name: "Download STL (.stl)" }).click();

  // Assert
  expect((await waiting).suggestedFilename()).toBe("vase.stl");
  await expect(page.getByRole("img", { name: /3D preview of the vase/ })).toBeVisible();
  expect(violations).toEqual([]);
  expect(outside).toEqual([]);
});

test("keeps its own look whatever the host page's rules", async ({ page }) => {
  // Arrange
  await ready(page, "/kit/host.html");

  // Act
  const make = await page.getByRole("button", { name: "Make it" }).evaluate((b) => getComputedStyle(b).backgroundColor);
  const host = await page.locator("main > p.error").evaluate((p) => getComputedStyle(p).color);

  // Assert
  expect(make).toBe("rgb(196, 56, 30)"); // --button-bg, not the host's green .button
  expect(host).toBe("rgb(0, 0, 255)"); // and the host's own rules still apply to its own content
});

test("lays itself out by its own width, not the window's", async ({ page }) => {
  // Arrange
  await page.setViewportSize({ width: 1280, height: 900 });
  await ready(page, "/kit/host.html");
  await describe(page, "a tall fluted vase, 20 cm");
  const columns = () => page.locator(".sm-workspace").evaluate((w) => getComputedStyle(w).gridTemplateColumns.split(" ").length);
  const wide = await columns();

  // Act
  await page.evaluate(() => { document.body.style.maxWidth = "46rem"; }); // kinglet.dev's column

  // Assert
  expect(wide).toBe(2);
  await expect.poll(columns).toBe(1);
});
