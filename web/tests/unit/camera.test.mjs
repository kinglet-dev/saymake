import { test } from "node:test";
import assert from "node:assert/strict";
import { perspective, lookAt, multiply, orbit, frame, transform } from "../../src/js/lib/camera.mjs";

const near = (a, b, tol = 1e-6) => a.forEach((x, i) => assert.ok(Math.abs(x - b[i]) < tol, `${i}: ${x} vs ${b[i]}`));

test("perspective builds the standard OpenGL projection, column-major", () => {
  // Arrange
  const f = 1 / Math.tan(Math.PI / 8);

  // Act
  const p = perspective(Math.PI / 4, 2, 1, 3);

  // Assert
  near([...p], [f / 2, 0, 0, 0, 0, f, 0, 0, 0, 0, -2, -1, 0, 0, -3, 0]);
});

test("lookAt puts the target straight ahead, on the negative z axis", () => {
  // Arrange
  const v = lookAt([0, -10, 0], [0, 0, 0], [0, 0, 1]);

  // Act
  const target = transform(v, [0, 0, 0]);
  const above = transform(v, [0, 0, 1]);

  // Assert
  near(target, [0, 0, -10]);
  near(above, [0, 1, -10]);
});

test("multiply by the identity changes nothing, and composes in order", () => {
  // Arrange
  const id = new Float32Array([1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1]);
  const move = new Float32Array([1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 5, 0, 0, 1]);
  const twice = new Float32Array([2, 0, 0, 0, 0, 2, 0, 0, 0, 0, 2, 0, 0, 0, 0, 1]);

  // Act
  const same = multiply(id, move);
  const moveThenScale = multiply(twice, move);

  // Assert
  near([...same], [...move]);
  near(transform(moveThenScale, [1, 0, 0]), [12, 0, 0]);
});

test("orbit turns with the drag and stops short of looking straight down or up", () => {
  // Arrange
  const start = { yaw: 0, pitch: 0.3 };

  // Act
  const turned = orbit(start, 100, 0);
  const down = orbit(start, 0, 10000);
  const up = orbit(start, 0, -10000);

  // Assert
  assert.ok(Math.abs(turned.yaw - -1) < 1e-9);
  assert.ok(Math.abs(down.pitch - 1.4) < 1e-9);
  assert.ok(Math.abs(up.pitch - -1.4) < 1e-9);
});

test("frame centers the shape and keeps all of it in front of the camera", () => {
  // Arrange
  const min = [-50, -50, 0], max = [50, 50, 200];

  // Act
  const f = frame(min, max, { yaw: 0.5, pitch: 0.3 }, 1.5);

  // Assert
  const center = transform(f.view, [0, 0, 100]);
  near(center.slice(0, 2), [0, 0], 1e-4);
  const radius = Math.hypot(100, 100, 200) / 2;
  assert.ok(-center[2] - radius > f.near, "the whole shape is past the near plane");
  assert.ok(-center[2] + radius < f.far, "and before the far plane");
  assert.equal(f.projection.length, 16);
});

test("frame looks at a flat shape too", () => {
  // Arrange
  const min = [-10, -10, 0], max = [10, 10, 0];

  // Act
  const f = frame(min, max, { yaw: 0, pitch: 0.5 }, 1);

  // Assert
  assert.ok(f.far > f.near && f.near > 0);
});
