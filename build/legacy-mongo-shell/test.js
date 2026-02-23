// Please do not merge changes in this file.

(function () {
  "use strict";

  const command = {
    aggregate: 1,
    pipeline: [{ $currentOp: {} }],
    cursor: {},
  };

  const actual = db.getSiblingDB("admin").runCommand(command);
  assert.commandWorked(actual);

  const op = actual.cursor.firstBatch.find((o) => o.command.aggregate != null);
  assert.eq(false, op == null);

  var failedTests = {};

  try {
    assert.eq("string", typeof op.currentOpTime);
  } catch (e) {
    failedTests["currentOpTime"] = e;
  }

  try {
    assert.eq("number", typeof op.opid);
  } catch (e) {
    failedTests["opid"] = e;
  }

  try {
    assert.eq(1, op.command.aggregate);
  } catch (e) {
    failedTests["commandAggregate"] = e;
  }

  try {
    assert.eq([{ $currentOp: {} }], op.command.pipeline);
  } catch (e) {
    failedTests["commandPipeline"] = e;
  }

  try {
    assert.eq({}, op.command.cursor);
  } catch (e) {
    failedTests["commandCursor"] = e;
  }

  try {
    assert.eq("admin", op.command.$db);
  } catch (e) {
    failedTests["commandDB"] = e;
  }

  assert.eq({}, failedTests);

  print("test.js passed!");
})();
