// Please do not merge changes in this file.

(function() {
  'use strict';
    
  let res = db.runCommand({"isMaster": 1, "client": {"application": "foobar"}});
  assert.commandFailed(res);
  assert.eq(res.code, ErrorCodes.ClientMetadataCannotBeMutated, tojson(res));

  print('test.js passed!');
})();
