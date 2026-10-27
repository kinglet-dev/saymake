const { test, expect, ready, describe } = require("./coverage");

const turn = (page) => page.evaluate(() => document.querySelector("#view3d").dataset.turn);
const pixels = (page) => page.locator("#view3d").screenshot();

test.describe("The 3D preview", () => {
  test.beforeEach(async ({ page }) => {
    // Arrange (shared)
    await ready(page);
    await describe(page, "hexagonal vase");
    await expect(page.getByRole("img", { name: /3D preview/ })).toBeVisible();
  });

  test("draws the shape", async ({ page }) => {
    // Arrange
    const canvas = page.locator("#view3d");

    // Act
    const shot = await canvas.screenshot();

    // Assert: the shape is drawn (a blank white canvas compresses to almost nothing)
    expect(shot.length).toBeGreaterThan(5000);
    await page.setViewportSize({ width: 700, height: 900 }); // and it redraws at a new size
    await expect.poll(async () => (await canvas.screenshot()).equals(shot)).toBe(false);
  });

  test("turns when dragged", async ({ page }) => {
    // Arrange
    await page.locator("#view3d").scrollIntoViewIfNeeded();
    const box = await page.locator("#view3d").boundingBox();
    const before = await pixels(page);

    // Act
    await page.mouse.move(box.x + 100, box.y + 100);
    await page.mouse.down();
    await page.mouse.move(box.x + 220, box.y + 140, { steps: 4 });
    await page.mouse.up();
    await page.mouse.move(box.x + 10, box.y + 10);

    // Assert
    expect((await pixels(page)).equals(before)).toBe(false);
    expect(await turn(page)).not.toBe("-0.900 0.450");
  });

  test("turns with the arrow keys, back home with Home, and ignores other keys", async ({ page }) => {
    // Arrange
    const canvas = page.locator("#view3d");
    await canvas.focus();
    const home = await turn(page);

    // Act
    const seen = [];
    for (const key of ["ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown", "ArrowLeft"]) {
      await canvas.press(key);
      seen.push(await turn(page));
    }
    await canvas.press("x");
    const afterOther = await turn(page);
    await canvas.press("Home");

    // Assert
    expect(seen).toEqual(["-0.750 0.450", "-0.900 0.450", "-0.900 0.300", "-0.900 0.450", "-0.750 0.450"]);
    expect(afterOther).toBe("-0.750 0.450");
    expect(await turn(page)).toBe(home);
  });

  test("stops turning when the pointer is lost", async ({ page }) => {
    // Arrange
    const canvas = page.locator("#view3d");
    await canvas.dispatchEvent("pointerdown", { clientX: 10, clientY: 10, pointerId: 1, isPrimary: true });
    await canvas.dispatchEvent("pointercancel", { pointerId: 1 });

    // Act
    await canvas.dispatchEvent("pointermove", { clientX: 200, clientY: 200, pointerId: 1 });

    // Assert
    expect(await turn(page)).toBe("-0.900 0.450");
  });
});

test("says when the browser can't show 3D, and still offers downloads and outlines", async ({ page }) => {
  // Arrange: a browser without WebGL 2
  await page.addInitScript(() => {
    const real = HTMLCanvasElement.prototype.getContext;
    HTMLCanvasElement.prototype.getContext = function (kind, ...rest) {
      return kind === "webgl2" ? null : real.call(this, kind, ...rest);
    };
  });
  await ready(page);

  // Act
  await describe(page, "vase");

  // Assert
  await expect(page.locator("#no3d")).toHaveText("This browser can't show 3D previews (WebGL 2 is off or missing). The downloads still work.");
  await expect(page.locator("#view3d")).toBeHidden();
  await expect(page.getByRole("button", { name: "Download STL (.stl)" })).toBeVisible();
  await describe(page, "coaster");
  await expect(page.locator("#no3d")).toBeHidden();
  await expect(page.getByRole("img", { name: /^Outline of the coaster/ })).toBeVisible();
});
