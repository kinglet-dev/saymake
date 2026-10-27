// WCAG 2.2 A and AA checks with axe-core (MPL-2.0; approved 2026-10-18 for
// test use only: it runs in the test browser and never ships with the page).
const fs = require("fs");
const { test, expect, ready, describe } = require("./coverage");

const axeSource = fs.readFileSync(require.resolve("axe-core/axe.min.js"), "utf8");

async function violations(page) {
  await page.evaluate(axeSource);
  const results = await page.evaluate(() =>
    window.axe.run(document, { runOnly: ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"] }));
  return results.violations.map((v) => `${v.id}: ${v.help} (${v.nodes.map((n) => n.target.join(" ")).join("; ")})`);
}

for (const colorScheme of ["light", "dark"]) {
  test.describe(`Accessibility in the ${colorScheme} theme`, () => {
    test.use({ colorScheme });

    for (const [name, text] of [
      ["the start page", null],
      ["a solid with warnings and notes", "vase 2000 mm, 0.5 mm thick walls, with glitter"],
      ["a flat shape", "round coaster with a border and a hanging hole"],
      ["a shape that can't be made", "bud vase with 8 mm thick walls"],
      ["an object it doesn't know", "a lamp"],
    ]) {
      test(`${name} meets WCAG 2.2 AA`, async ({ page }) => {
        // Arrange
        await ready(page);
        if (text) await describe(page, text);
        await page.waitForFunction(() => document.fonts.check("16px Outfit"));

        // Act
        const broken = await violations(page);

        // Assert
        expect(broken).toEqual([]);
      });
    }
  });
}

test("works at 320 px wide without scrolling sideways, and at 200% text", async ({ page }) => {
  // Arrange
  await page.setViewportSize({ width: 320, height: 800 });
  await ready(page);
  await describe(page, "a tall fluted vase, 20 cm");

  // Act
  const wide = await page.evaluate(() => document.documentElement.scrollWidth);
  await page.evaluate(() => { document.documentElement.style.fontSize = "200%"; });
  const wideAtDouble = await page.evaluate(() => document.documentElement.scrollWidth);

  // Assert
  expect(wide).toBeLessThanOrEqual(320);
  expect(wideAtDouble).toBeLessThanOrEqual(320);
});

test("can be used with the keyboard alone, in reading order", async ({ page }) => {
  // Arrange
  await ready(page);

  // Act
  await page.keyboard.type("hexagonal planter");
  await page.keyboard.press("Enter");
  await expect(page.getByRole("heading", { level: 2 })).toHaveText("Planter");
  const order = [];
  for (let i = 0; i < 5; i++) { // the description field has focus first
    await page.keyboard.press("Tab");
    order.push(await page.evaluate(() => document.activeElement.id || document.activeElement.textContent.trim()));
  }

  // Assert
  expect(order).toEqual(["make", "Printer", "view3d", "Download STL (.stl)", "Download 3MF (.3mf)"]);
});
