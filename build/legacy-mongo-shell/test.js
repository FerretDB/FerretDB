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

  const query = { v: { $gt: 42.0 } };

  const expected = [{ _id: "double", v: 42.13 }];

  const actual = coll.find(query).toArray();
  assert.eq(expected, actual);

  print("test.js passed!");
})();
