// Please do not merge changes in this file.

(function () {
  "use strict";

  // Update the following example with your test.

  const coll = db.test;

  coll.drop();

  coll.insertOne({ _id: "timestamp-zero", v: Timestamp(0,0)});

  const expected = [{ _id: "timestamp-zero", v: Timestamp(0,0)}];

  const actual = coll.find({}).toArray();
  assert.eq(expected, actual);

  print("test.js passed!");
})();
