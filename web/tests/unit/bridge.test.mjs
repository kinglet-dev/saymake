import { test } from "node:test";
import assert from "node:assert/strict";
import { createBridge } from "../../src/js/lib/bridge.mjs";

function fakeCall(answer) {
  const sent = [];
  const call = (req) => {
    sent.push(JSON.parse(req));
    return answer;
  };
  return { call, sent };
}

test("read sends the description and returns the parsed answer with the mesh", () => {
  // Arrange
  const mesh = new Uint8Array([1, 2]);
  const { call, sent } = fakeCall({ json: '{"ok":true,"reading":{"concept":"vase"}}', mesh, file: null });
  const b = createBridge(call);

  // Act
  const got = b.read("a vase", [256, 256, 256]);

  // Assert
  assert.deepEqual(sent, [{ op: "read", text: "a vase", printer: [256, 256, 256] }]);
  assert.equal(got.ok, true);
  assert.equal(got.reading.concept, "vase");
  assert.equal(got.mesh, mesh);
  assert.equal(got.file, null);
});

test("build sends the family and values", () => {
  // Arrange
  const { call, sent } = fakeCall({ json: '{"ok":false,"error":"no"}', mesh: null, file: null });
  const b = createBridge(call);

  // Act
  const got = b.build("tray", { length: 300 }, [200, 200, 200]);

  // Assert
  assert.deepEqual(sent, [{ op: "build", family: "tray", values: { length: 300 }, printer: [200, 200, 200] }]);
  assert.deepEqual([got.ok, got.error], [false, "no"]);
});

test("exportFile asks for a format and returns the file", () => {
  // Arrange
  const file = new Uint8Array([9]);
  const { call, sent } = fakeCall({ json: '{"ok":true,"file":{"name":"vase.stl","mediaType":"model/stl"}}', mesh: null, file });
  const b = createBridge(call);

  // Act
  const got = b.exportFile("vessel", { height: 200 }, "stl", "vase", [256, 256, 256]);

  // Assert
  assert.deepEqual(sent, [{ op: "export", family: "vessel", values: { height: 200 }, format: "stl", name: "vase", printer: [256, 256, 256] }]);
  assert.equal(got.file, file);
  assert.equal(got.fileInfo.name, "vase.stl");
});

test("an answer that isn't JSON becomes an error answer", () => {
  // Arrange
  const { call } = fakeCall({ json: "{", mesh: null, file: null });
  const b = createBridge(call);

  // Act
  const got = b.read("x");

  // Assert
  assert.equal(got.ok, false);
  assert.match(got.error, /couldn't understand the tool's answer/);
});
