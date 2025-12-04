// Please do not merge changes in this file.

(function () {
  "use strict";

  const coll = db.test;

  coll.drop();

  const doc = {
    _id: "id",
    v: Timestamp(0, 0),
    d: {
      dv: Timestamp(0, 0),
    },
  };

  coll.insert(doc);

  const actual = coll.find({}).toArray()[0];
  assert.eq("id", actual._id);

  const now = Date.now() / 1000;

  assert.neq(Timestamp(0, 0), actual.v);
  assert.neq(0, actual.v.i);
  assert.between(now - 5, actual.v.t, now + 5);

  assert.eq(Timestamp(0, 0), actual.d.dv);
  assert.eq(0, actual.d.dv.i);

  print("test.js passed!");
})();
