// After every browser test: the browser-only modules must have run in full.
const fs = require("fs");
const { uncovered, outDir, expected } = require("./coverage");

module.exports = async () => {
  const { scripts, missing } = uncovered();
  fs.rmSync(outDir, { recursive: true, force: true });
  if (JSON.stringify(scripts) !== JSON.stringify(expected)) {
    throw new Error(`JS coverage saw ${scripts.join(", ") || "no scripts"}, expected ${expected.join(", ")}`);
  }
  if (missing.length > 0) throw new Error(`JS coverage below 100%:\n${missing.join("\n")}`);
  console.log(`JS coverage: 100% of ${expected.join(", ")}`);
};
