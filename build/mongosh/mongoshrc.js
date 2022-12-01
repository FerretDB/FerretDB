// Please do not commit changes in this file.

(function () {
    'use strict';
    const assert = require('assert'); // https://nodejs.org/api/assert.html

    const col = db.test;

    col.drop();

    const init = [
        { _id: 'double', v: 42.13 },
        { _id: 'double-whole', v: 42.0 },
        { _id: 'double-zero', v: 0.0 },
    ];

    col.insertMany(init);

    const query = { v: { $gt: 42.0 } };

    const expected = [
        { _id: 'double', v: 42.13 },
    ];

    let actual = [
        { _id: 'double', v: 42.13 },
    ];

    assert.deepStrictEqual(actual, expected);

    actual = col.find(query).toArray();

    // Object.setPrototypeOf(actual, Object.getPrototypeOf(expected));

    console.log(Object.getPrototypeOf(expected));
    console.log(Object.getPrototypeOf(actual));
    console.log(Object.getPrototypeOf(expected) === Object.getPrototypeOf(actual));

    console.log(Object.getPrototypeOf(expected[0]));
    console.log(Object.getPrototypeOf(actual[0]));
    console.log(Object.getPrototypeOf(expected[0]) === Object.getPrototypeOf(actual[0]));

    assert.deepStrictEqual(actual, expected);
})();
