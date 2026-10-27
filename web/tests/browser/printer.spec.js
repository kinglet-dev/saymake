const { test, expect, ready, describe } = require("./coverage");

// openPrinter opens the collapsed Printer section, as a person would.
const openPrinter = (page) => page.locator("summary", { hasText: "Printer" }).click();

test.describe("The printer size", () => {
  test.beforeEach(async ({ page }) => {
    // Arrange (shared)
    await ready(page);
  });

  test("starts at 256 mm each way", async ({ page }) => {
    // Arrange
    const sizes = ["Printer width", "Printer depth", "Printer height"];

    // Act
    const values = [];
    for (const s of sizes) values.push(await page.getByLabel(s).inputValue());

    // Assert
    expect(values).toEqual(["256", "256", "256"]);
  });

  test("warns when the shape won't fit, as soon as the size changes", async ({ page }) => {
    // Arrange
    await describe(page, "vase 20 cm");
    await expect(page.getByRole("list", { name: "Printing checks" })).toBeHidden();

    // Act
    await openPrinter(page);
    await page.getByLabel("Printer height").fill("150");

    // Assert
    await expect(page.getByRole("list", { name: "Printing checks" })).toContainText("doesn't fit a 256 × 256 × 150 mm printer");
    await expect(page.getByRole("button", { name: "Download STL (.stl)" })).toBeVisible();
  });

  test("uses the printer size when reading a description too", async ({ page }) => {
    // Arrange
    await openPrinter(page);
    await page.getByLabel("Printer width").fill("50");

    // Act
    await describe(page, "bowl");

    // Assert
    await expect(page.getByRole("list", { name: "Printing checks" })).toContainText("doesn't fit a 50 × 256 × 256 mm printer");
  });

  test("says what's wrong with a size out of range", async ({ page }) => {
    // Arrange
    await describe(page, "vase");

    // Act
    await openPrinter(page);
    await page.getByLabel("Printer depth").fill("5");

    // Assert
    await expect(page.getByRole("alert")).toHaveText("the printer size must be three numbers from 10 to 2000 mm");
  });

  test("does nothing before there is a shape", async ({ page }) => {
    // Arrange
    await openPrinter(page);
    const height = page.getByLabel("Printer height");

    // Act
    await height.fill("150");

    // Assert
    await expect(page.getByRole("alert")).toBeHidden();
    await expect(page.locator("#result")).toBeHidden();
  });
});

test("shows words for things it can't make yet as their own chips", async ({ page }) => {
  // Arrange
  await ready(page);

  // Act
  await describe(page, "a vase with a lid");

  // Assert
  const chip = page.getByRole("list", { name: "How your description was read" }).getByRole("listitem").filter({ hasText: "lid: not supported yet" });
  await expect(chip).toHaveClass("sm-unsupported");
  await expect(chip).toHaveAttribute("title", "Lids aren't supported yet.");
  await expect(page.getByRole("list", { name: "Notes" })).toContainText("Lids aren't supported yet.");
});

test("shows the Printer and Settings sections as ones to open, with their arrows", async ({ page }) => {
  // Arrange
  await ready(page);
  await describe(page, "a tall fluted vase, 20 cm");

  // Act
  const shown = await page.locator("summary").evaluateAll((all) => all.map((s) => getComputedStyle(s).display));

  // Assert
  expect(shown).toEqual(["list-item", "list-item"]);
});
