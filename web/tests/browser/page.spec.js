const { test, expect, ready, describe } = require("./coverage");

test.describe("Describing an object", () => {
  test.beforeEach(async ({ page }) => {
    // Arrange (shared)
    await ready(page);
  });

  test("shows the reading, a 3D preview, checks, downloads and settings", async ({ page }) => {
    // Arrange
    const text = "a tall fluted vase, 20 cm";

    // Act
    await describe(page, text);

    // Assert
    await expect(page.getByRole("status")).toHaveText(/^Made a vase, .* × .* × 200 mm\.$/);
    await expect(page.getByRole("heading", { level: 2 })).toHaveText("Vase");
    const chips = page.getByRole("list", { name: "How your description was read" }).getByRole("listitem");
    await expect(chips).toHaveText(["Vessel › Vase", "tall", "fluted", "Height 200 mm"]);
    await expect(chips.first()).toHaveAttribute("title", "The object: “vase”");
    await expect(page.getByRole("img", { name: /^3D preview of the vase, .* × .* × 200 mm/ })).toBeVisible();
    await expect(page.getByRole("button", { name: "Download STL (.stl)" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Download 3MF (.3mf)" })).toBeVisible();
    await expect(page.getByRole("slider", { name: "Height", exact: true })).toHaveAttribute("aria-valuetext", "200 mm");
    await expect(page.getByRole("list", { name: "Notes" })).toBeHidden();
    await expect(page.getByRole("list", { name: "Printing checks" })).toBeHidden();
  });

  test("shows a flat shape as outlines, with notes and words it didn't use", async ({ page }) => {
    // Arrange
    const text = "fluted round coaster 2000 mm with a border and glitter";

    // Act
    await describe(page, text);

    // Assert
    await expect(page.getByRole("img", { name: /^Outline of the coaster, 600 × 600 mm/ })).toBeVisible();
    await expect(page.locator("#view2d path.sm-cut")).toHaveCount(1);
    await expect(page.locator("#view2d path.sm-score")).toHaveCount(1);
    await expect(page.getByRole("list", { name: "Notes" })).toContainText("Length 2000 mm is more than the largest allowed");
    const chips = page.getByRole("list", { name: "How your description was read" }).getByRole("listitem");
    await expect(chips.filter({ hasText: "fluted: not used" })).toHaveClass("sm-unused");
    await expect(chips.filter({ hasText: "glitter: not understood" })).toHaveClass("sm-unknown");
    await expect(page.getByRole("button", { name: "Download SVG (.svg)" })).toBeVisible();
    await expect(page.locator("#view3d")).toBeHidden();
  });

  test("asks for a description when there is none", async ({ page }) => {
    // Arrange
    const text = "   ";

    // Act
    await describe(page, text);

    // Assert
    await expect(page.getByRole("alert")).toHaveText("Type a description first, like “a tall fluted vase, 20 cm”.");
  });

  test("says what it can make when it doesn't know the object", async ({ page }) => {
    // Arrange
    await describe(page, "a vase");

    // Act
    await describe(page, "a lamp");

    // Assert
    await expect(page.getByRole("alert")).toContainText("I couldn't tell what to make. Name an object, such as");
    await expect(page.locator("#result")).toBeHidden();
    await expect(page.getByRole("status")).toHaveText("");
  });

  test("shows the settings to fix a shape that can't be made", async ({ page }) => {
    // Arrange
    await describe(page, "bud vase with 8 mm thick walls");
    await expect(page.getByRole("alert")).toContainText("The wall is too thick");
    await expect(page.getByRole("status")).toHaveText("Read your description, but couldn't make the shape; see the message below.");
    await expect(page.getByRole("button", { name: /Download/ })).toHaveCount(0);

    // Act
    await page.getByRole("slider", { name: "Wall", exact: true }).fill("2");

    // Assert
    await expect(page.getByRole("status")).toHaveText(/^Updated: /);
    await expect(page.getByRole("alert")).toBeHidden();
    await expect(page.getByRole("button", { name: "Download STL (.stl)" })).toBeVisible();
  });

  test("rebuilds as a setting moves, and says when the settings can't be made", async ({ page }) => {
    // Arrange
    await describe(page, "bud vase");
    const wall = page.getByRole("slider", { name: "Wall", exact: true });

    // Act
    await wall.fill("10");

    // Assert
    await expect(page.getByRole("alert")).toContainText("The wall is too thick");
    await expect(page.getByRole("status")).toHaveText("Couldn't make that; see the message below.");
    await expect(page.locator("#size")).toHaveText("Not updated: these settings can't be made.");
    await expect(wall).toHaveAttribute("aria-valuetext", "10 mm");
  });

  test("shows the checks' warnings", async ({ page }) => {
    // Arrange
    await describe(page, "vase");

    // Act
    await page.getByRole("slider", { name: "Wall", exact: true }).fill("0.5");

    // Assert
    const checks = page.getByRole("list", { name: "Printing checks" });
    await expect(checks).toContainText("CheckThe thinnest wall is about");
    await expect(checks.getByRole("listitem")).toHaveClass("sm-warning-level");
  });

  test("downloads the file, named after the object", async ({ page }) => {
    // Arrange
    await describe(page, "round coaster");
    const waiting = page.waitForEvent("download");

    // Act
    await page.getByRole("button", { name: "Download SVG (.svg)" }).click();

    // Assert
    const file = await waiting;
    expect(file.suggestedFilename()).toBe("coaster.svg");
    const body = await (await file.createReadStream()).toArray();
    expect(Buffer.concat(body).toString()).toContain('<g id="cut"');
    await expect(page.getByRole("status")).toHaveText("Downloading coaster.svg.");
    await page.waitForTimeout(1100); // the page lets go of the file's temporary address after a second
  });

  test("says why a download failed", async ({ page }) => {
    // Arrange: the tool refuses this export
    await describe(page, "round coaster");
    await page.evaluate(() => {
      const real = globalThis.saymake.call;
      globalThis.saymake.call = (req) => (JSON.parse(req).op === "export"
        ? { json: '{"ok":false,"error":"The disk is full."}', mesh: null, file: null }
        : real(req));
    });

    // Act
    await page.getByRole("button", { name: "Download SVG (.svg)" }).click();

    // Assert
    await expect(page.getByRole("alert")).toHaveText("The disk is full.");
  });

  test("shows an error finding as one that stops printing", async ({ page }) => {
    // Arrange: the tool reports a blocking check
    await page.evaluate(() => {
      const real = globalThis.saymake.call;
      globalThis.saymake.call = (req) => {
        const r = real(req);
        const a = JSON.parse(r.json);
        a.shape.findings = [{ level: "error", message: "Broken." }];
        a.shape.formats = null; // as the tool sends it: no formats when a check fails
        return { ...r, json: JSON.stringify(a) };
      };
    });

    // Act
    await describe(page, "vase");

    // Assert
    const item = page.getByRole("list", { name: "Printing checks" }).getByRole("listitem");
    await expect(item).toHaveText("Can't printBroken.");
    await expect(item).toHaveClass("sm-error-level");
    await expect(page.getByRole("button", { name: /Download/ })).toHaveCount(0);
  });
});

test("names itself, with its version, license and source", async ({ page }) => {
  // Arrange
  await ready(page);

  // Act
  const footer = page.getByRole("contentinfo");

  // Assert
  await expect(page).toHaveTitle("Kinglet Saymake: describe it, then print or cut it");
  await expect(page.getByRole("banner")).toContainText("Saymake");
  await expect(footer).toHaveText("Kinglet Saymake dev · MIT License · Source on GitHub");
  await expect(footer.getByRole("link", { name: "Source on GitHub" })).toHaveAttribute("href", "https://github.com/kinglet-dev/saymake");
});
