// Please do not merge changes in this file.

(function () {
  "use strict";

  const coll = db.test;

  coll.drop();

  const init = [
    { _id: "decimal128", v: NumberDecimal("42.13") },
    { _id: "decimal128-inf", v: NumberDecimal("Infinity") },
    { _id: "decimal128-nan", v: NumberDecimal("NaN") },
  ];

  coll.insertMany(init);


	let expected = [{ "_id" : "decimal128", "v" : NumberDecimal("42.13") }];
	let actual = coll.find({ v: { $bitsAnySet: 6 }}).toArray();
	assert.eq(expected, actual);

	expected = [];
	actual = coll.find({ v: { $bitsAnyClear: 6 }}).toArray();
	assert.eq(expected, actual);

	expected = [];
	actual = coll.find({ v: { $bitsAllClear: 6 }}).toArray();
	assert.eq(expected, actual);

  print("test.js passed!");
})();
