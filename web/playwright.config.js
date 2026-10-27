// Browser tests against build/dev (run `node build.mjs` first), served by
// Python's standard-library web server; plus the single file over file://.
const { defineConfig, devices } = require("@playwright/test");

const port = 4377;

module.exports = defineConfig({
  testDir: "tests/browser",
  globalTeardown: require.resolve("./tests/browser/coverage-teardown.js"),
  forbidOnly: true,
  retries: 0,
  workers: 2,
  reporter: [["list"]],
  timeout: 60000,
  use: {
    baseURL: `http://127.0.0.1:${port}`,
    ...devices["Desktop Chrome"],
    // Software WebGL, so the 3D preview renders without a GPU.
    launchOptions: { args: ["--use-gl=angle", "--use-angle=swiftshader", "--enable-unsafe-swiftshader"] },
  },
  webServer: {
    command: `python3 -m http.server ${port} --bind 127.0.0.1 --directory build/dev 2>/dev/null`,
    url: `http://127.0.0.1:${port}/`,
    reuseExistingServer: false,
  },
});
