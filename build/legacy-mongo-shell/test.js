// Please do not merge changes in this file.

(function () {
  "use strict";

  // Update the following example with your test.

  const coll = db.test;

  coll.drop();

  const init = [
    { _id: "double", v: 42.13 },
    { _id: "double-whole", v: 42.0 },
    { _id: "double-zero", v: 0.0 },
  ];

  coll.insertMany(init);

  const created = coll.createIndexes([{ v: 1 } ]);
  assert.commandWorked(created);

  const dropped = db.runCommand({dropIndexes: 'test', index: {v:1}});
  assert.commandWorked(dropped);

  var failedTests = {};

  try {
    assert.eq(typeof dropped.ok, typeof 1.0);
  } catch (e) {
    failedTests['ok'] = e;
  };

  try {
    assert.eq(typeof dropped.nIndexesWas, typeof 1);
  } catch (e) {
    failedTests['nIndexesWas'] = e;
  };

  assert.eq(failedTests, {});

  print("test.js passed!");
})();
