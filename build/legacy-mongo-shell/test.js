// Please do not merge changes in this file.

(function () {
  "use strict";

  db.coll.drop();

  const l = 1939;
  // const l = 1938;

  const v = 'x'.repeat(l);

  db.coll.insertOne({v: v});

  const actual = db.coll.findOneAndDelete({});

  assert.eq(v, actual.v);

  print("test.js passed!");
})();
