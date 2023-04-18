// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.test;

  coll.drop();

  const init = [
    {_id: 'exists'},
  ];

  coll.insertMany(init);

  var actualErr = null
  try {
    coll.findOneAndUpdate({_id:"not_exists"}, {$set: {_id:"exists"}}, {upsert:true})
  }
  catch(err) {
    actualErr = err
  }

  const expectedErrCode = 66

  assert.eq(expectedErrCode, actualErr.code);

  print('test.js passed!');
})();
