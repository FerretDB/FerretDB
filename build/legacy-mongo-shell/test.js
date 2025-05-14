// Please do not merge changes in this file.

(function () {
  "use strict";

  // Update the following example with your test.

  const coll = db.test;

  coll.drop();

  const init = [{ _id: "double", v: 42.13 }];

  coll.insertMany(init);

  const res = db.test.runCommand({ collStats: "test" });
  assert.commandWorked(res);

  var failedTests = {};

  try {
    assert.eq(res.numOrphanDocs, NumberInt(0));
  } catch (e) {
    failedTests["numOrphanDocs"] = e;
  }

  try {
    assert.eq(res.capped, false);
  } catch (e) {
    failedTests["capped"] = e;
  }

  try {
    assert.eq(Object.keys(res.indexDetails), ["_id_"]);
  } catch (e) {
    failedTests["indexDetails"] = e;
  }

  try {
    assert.eq(typeof res.indexSizes._id_, "number");
  } catch (e) {
    failedTests["indexSizesFields"] = e;
  }

  assert.eq(failedTests, {});
  print("test.js passed!");
})();
