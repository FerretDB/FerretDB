// Please do not merge changes in this file.

(function () {
  "use strict";

  // Update the following example with your test.

  const coll = db.test;

  coll.drop();

  const init = [{ _id: "array", v: [42.0] }];

  coll.insertMany(init);

  const failedTests = {};

  try {
    const res = db.runCommand({ findAndModify: "test", maxTimeMS: "string" });
    assert.commandFailedWithCode(res, 2);
  } catch (e) {
    failedTests["WrongMaxTimeMS"] = e;
  }

  assert.eq(failedTests, {}, "tests failed");

  print("test.js passed!");
})();
