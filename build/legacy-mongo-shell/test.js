// Please do not merge changes in this file.

(function () {
  "use strict";

  const coll = db.test;

  coll.drop();

  const init = [
    { _id: "decimal128", v: 42.13 }
  ];

  coll.insertMany(init);

  const query = { v: { $bitsAnySet: 6 } };

  const expected = [];

  const actual = coll.find(query).toArray();
  assert.eq(expected, actual);

  print("test.js passed!");
})();
