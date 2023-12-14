// Please do not merge changes in this file.

(function() {
  'use strict';

  const coll = db.test;

  coll.drop();

  coll.createIndexes([ { foo: 1 }, { bar: 1 }] )

  coll.createIndexes([ { foo: 1 }, {var: 1}, { bar: 1 }] ) // FerretDB panics here
})();
