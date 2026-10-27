import { test } from "node:test";
import assert from "node:assert/strict";
import { latest } from "../../src/js/lib/schedule.mjs";

function fakeTimers() {
  let next = 1;
  const pending = new Map();
  return {
    setTimeout: (fn) => { const id = next++; pending.set(id, fn); return id; },
    clearTimeout: (id) => pending.delete(id),
    flush: () => { const fns = [...pending.values()]; pending.clear(); fns.forEach((f) => f()); },
    count: () => pending.size,
  };
}

test("latest runs once, with the last arguments, after a burst of calls", () => {
  // Arrange
  const timers = fakeTimers();
  const seen = [];
  const run = latest((v) => seen.push(v), 100, timers);

  // Act
  run(1); run(2); run(3);
  const waiting = timers.count();
  timers.flush();

  // Assert
  assert.equal(waiting, 1);
  assert.deepEqual(seen, [3]);
});

test("latest runs again for calls after it ran", () => {
  // Arrange
  const timers = fakeTimers();
  const seen = [];
  const run = latest((v) => seen.push(v), 100, timers);

  // Act
  run(1); timers.flush(); run(2); timers.flush();

  // Assert
  assert.deepEqual(seen, [1, 2]);
});

test("latest uses the global timers when none are given", async () => {
  // Arrange
  const seen = [];
  const run = latest((v) => seen.push(v), 1);

  // Act
  run(5);
  await new Promise((r) => setTimeout(r, 20));

  // Assert
  assert.deepEqual(seen, [5]);
});
