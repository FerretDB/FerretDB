// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.compound_index_max_fields;
  t.drop();

  const spec = {};
  for (let i = 0; i < 32; i++) {
    spec['f' + i] = (i % 2 == 0) ? 1 : -1;
  }

  assert.commandWorked(t.createIndex(spec));

  // create an index that has one too many fields.
  spec['f32'] = 1;
  assert.commandFailedWithCode(t.createIndex(spec), 13103); // https://github.com/stbrody/mongo/blob/master/docs/errors.md#srcmongobsonorderingh

  print('test.js passed!');
})();
