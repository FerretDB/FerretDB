// Move to test.js and run with `task mongo-test`.

(function() {
    "use strict";

    const col = db.test;

    col.drop();

    const init = [
        { _id: "double", v: 42.13 },
        { _id: "double-whole", v: 42.0 },
        { _id: "double-zero", v: 0.0 },
    ];

    col.insertMany(init);

    const query = { v: { $gt: 42.0 } };

    const expected = [
        { _id: "double", v: 42.13 },
    ];

    const actual = col.find(query).toArray();

    assert.eq(expected, actual);

    print("test.js passed!")
})();
