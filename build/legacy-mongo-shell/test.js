// Please do not merge changes in this file.

(function () {
  "use strict";

  // Update the following example with your test.

  const coll = db.test;

  coll.drop();


  const init = [
    { v: 1 },
  ];

  coll.insertMany(init);

  const actual = db.runCommand({dbStats:1});

  assert.eq(NumberLong(1), actual.objects,"objects not equal");
  assert.eq(33, actual.avgObjSize,"avgObjSize not equal");

  print("test.js passed!");
})();
