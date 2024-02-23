// Please do not merge changes in this file.

(function() {
  'use strict';

  db.getSiblingDB('admin').system.users.remove({});

  db.getSiblingDB('admin').createUser({user: 'username', pwd: 'password', roles: []});

  mongoClient = function(uri) {
    return new Mongo(uri);
  }

  const uri = 'mongodb://username:password@localhost:27017/?authMechanism=SCRAM-SHA-1';

  try {
    mongoClient(uri);
  } catch (e) {
    print('test.js failed: ' + e);
    return;
  }

  print('test.js passed!');
})();
