/* eslint-disable max-len */
// Please do not merge changes in this file.

(function() {
  'use strict';

  let client = db.runCommand({whatsmyuri: 1}).you;

  print('connected to: ' + client);

  const t = db.foo;
  t.drop();

  const admin = db.getSiblingDB('admin');

  let res = db.runCommand({ping: 1});
  assert.eq(res.ok, 1, 'ping failed');

  res = t.insert({});
  assert.writeOK(res, 'insert failed');

  let port = 27017;

  const roles = [];

  if (admin.runCommand({getParameter: '*'}).wiredTigerConcurrentReadTransactions !== undefined) {
    roles.push({role: 'read', db: 'admin'});
    port = 47017;
  };


  admin.system.users.remove({user: 'user'});
  admin.createUser({user: 'user', pwd: '1234', roles: roles});

  const mongoClient = function(uri) {
    return new Mongo(uri);
  };

  const uri = 'mongodb://user:1234@host.docker.internal:' + port + '/?authMechanism=SCRAM-SHA-1';

  try {
    client = mongoClient(uri);
  } catch (e) {
    throw new Error('test.js failed: ' + e);
  }

  print('connected to: ' + client.getDB('test').runCommand({whatsmyuri: 1}).you);

  print('test.js passed!');
})();
