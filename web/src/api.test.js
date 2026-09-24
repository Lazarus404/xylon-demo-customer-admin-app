import test from "node:test";
import assert from "node:assert/strict";
import { parsePresenceRoster } from "./api.js";

test("parsePresenceRoster map shape", () => {
  const rows = parsePresenceRoster({
    participants: { alice: { role: "host", joined_at: 1 }, bob: { role: "guest" } }
  });
  assert.equal(rows.length, 2);
  assert.deepEqual(rows[0], { id: "alice", role: "host", meta: '{"joined_at":1}' });
  assert.deepEqual(rows[1], { id: "bob", role: "guest", meta: "" });
});

test("parsePresenceRoster array shape", () => {
  const rows = parsePresenceRoster({
    participants: [{ participant_id: "p1", role: "host" }]
  });
  assert.deepEqual(rows, [{ id: "p1", role: "host", meta: "" }]);
});

test("parsePresenceRoster empty", () => {
  assert.deepEqual(parsePresenceRoster(null), []);
  assert.deepEqual(parsePresenceRoster({}), []);
  assert.deepEqual(parsePresenceRoster({ participants: {} }), []);
});
