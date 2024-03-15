/* eslint-disable max-len */
// Please do not merge changes in this file.

(function() {
  'use strict';

  const t = db.foo;
  t.drop();

  let res = db.runCommand({ping: 1});
  assert.eq(res.ok, 1, 'ping failed');

  res = t.insert({});
  assert.writeOK(res, 'insert failed');

  let port = 27017;

  const roles = [];

  if (db.getSiblingDB('admin').runCommand({getParameter: '*'}).wiredTigerConcurrentReadTransactions !== undefined) {
    roles.push({role: 'read', db: 'admin'});
    port = 47017;
  };

  // db.getSiblingDB('admin').system.users.remove({});

  db.getSiblingDB('admin').createUser({user: 'username2', pwd: 'password2', roles: roles});

  const mongoClient = function(uri) {
    return new Mongo(uri);
  };

  const uri = 'mongodb://username2:password2@host.docker.internal:' + port + '/?authMechanism=SCRAM-SHA-1';

  try {
    mongoClient(uri);
  } catch (e) {
    throw new Error('test.js failed: ' + e);
  }

  print('test.js passed!');
})();
