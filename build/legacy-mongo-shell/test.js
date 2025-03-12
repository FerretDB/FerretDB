// Please do not merge changes in this file.

(function () {
  "use strict";

  // Update the following example with your test.

  const coll = db.test;

  coll.drop();

  const init = [
    {
      _id: "doc1",
      text: "apple banana cherry"
    }
  ];

  coll.insertMany(init);

  coll.createIndex({ text: "text" });

  const query = { $text: { $search: "banana" } };
  const projection = { text: 1, score: { $meta: "textScore" } };
  const sort = { score: { $meta: "textScore" } };

  const actual = coll.find(query, projection).sort(sort).toArray();

  const expected = [
    {
      _id: "doc1",
      text: "apple banana cherry",
      score: actual.length > 0 ? actual[0].score : null
    }
  ];

  assert.eq(expected, actual);

  print("test.js passed!");
})();
