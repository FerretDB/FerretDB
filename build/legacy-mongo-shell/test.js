// Please do not merge changes in this file.

(function() {
    'use strict';

    const coll = db.test;

    const expected = { _id: 1, a: 1 };

    coll.drop();

    coll.insertOne({_id:1});

    const actual = coll.findOneAndReplace({_id:1},{_id:1,a:2},{upsert:true,returnDocument:'after'})

    assert.eq(expected, actual);

    print('test issue 2830.js passed!');
  })();