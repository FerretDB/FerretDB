// Please do not merge changes in this file.

(function() {
  'use strict';

  const maxBsonObjectSize = db.hello().maxBsonObjectSize;
  const docOverhead = Object.bsonsize({_id: new ObjectId(), x: ''});
  const maxStrSize = maxBsonObjectSize - docOverhead;
  const maxStr = 'a'.repeat(maxStrSize);
  const coll = db.max_doc_size;

  assert.commandWorked(coll.insert({_id: new ObjectId(), x: maxStr}));

  const largerThanMaxString = maxStr + 'a';

  coll.drop();
  assert.commandFailedWithCode(coll.insert({_id: new ObjectId(), x: largerThanMaxString}), 2);

  print('test.js passed!');
})();
