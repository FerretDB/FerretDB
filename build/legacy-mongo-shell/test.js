// Please do not merge changes in this file.

(function() {
  'use strict';

  let port = 27017;

  let roles = [];

  if (db.getSiblingDB('admin').runCommand({getParameter: '*'}).wiredTigerConcurrentReadTransactions !== undefined) {
    roles.push({role: 'read', db: 'admin'});
    port = 47017;
  };

  db.getSiblingDB('admin').system.users.remove({});

  db.getSiblingDB('admin').createUser({user: 'username', pwd: 'password', roles: roles});

  const mongoClient = function(uri) {
    return new Mongo(uri);
  }

  const uri = 'mongodb://username:password@host.docker.internal:' + port + '/?authMechanism=SCRAM-SHA-1';

  try {
    mongoClient(uri);
  } catch (e) {
    throw new Error('test.js failed: ' + e);
  }

  print('test.js passed!');
})();
