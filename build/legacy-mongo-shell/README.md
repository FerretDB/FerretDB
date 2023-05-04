### Useful assertions

The legacy `mongo` shell uses its own assertions defined [here](https://github.com/mongodb/mongo/blob/master/src/mongo/shell/assert.js). This deviates from `mongosh` this which uses the standard Node.js [assert](https://nodejs.org/api/assert.html) module.

A useful function is `assert.commandFailedWithCode` which asserts that the command failed with the expected code as the name implies. One should pass the result of a call to the `db.runCommand()` helper as this provides a result type that the function can parse.

For example to check if the `findAndModify` command failed with the error code `ImmutableField` you would do the following:

```js
> const res = db.runCommand({findAndModify: "foo", query: {}, update: {$set: {_id: 1}}});
> assert.commandFailedWithCode(res, ErrorCodes.ImmutableField);
```

Here `ErrorCodes` is an object that is generated from various source files. The keys are the error names and the values correspond to their respective error codes.


`assert.commandFailedWithCode(res, expectedCode, msg)`

throws if the result did not contain the expected code.

`assert.commandWorked(res)`

throws if the result contained an error.

`assert.sameMembers(aArr, bArr, msg, compareFn = _isDocEq)`

throws if the two arrays do not have the same members, in any order. By default, nested arrays must have the same order to be considered equal. Optionally accepts a compareFn to compare values instead of using docEq.

`assert.writeOK(res, msg, {ignoreWriteConcernErrors} = {})`

throws if write result contained an error.

Here are the available functions:

```
assert.adminCommandWorkedAllowingNetworkError
assert.apply
assert.automsg
assert.between
assert.betweenEx
assert.betweenIn
assert.bind
assert.call
assert.close
assert.closeWithinMS
assert.commandFailed
assert.commandFailedWithCode
assert.commandWorked
assert.commandWorkedIgnoringWriteConcernErrors
assert.commandWorkedIgnoringWriteErrors
assert.commandWorkedIgnoringWriteErrorsAndWriteConcernErrors
assert.commandWorkedOrFailedWithCode
assert.contains
assert.containsPrefix
assert.docEq
assert.doesNotThrow
assert.dropExceptionsWithCode
assert.eq
assert.gt
assert.gte
assert.hasFields
assert.hasOwnProperty
assert.includes
assert.isnull
assert.lt
assert.lte
assert.neq
assert.noAPIParams
assert.propertyIsEnumerable
assert.retry
assert.retryNoExcept
assert.sameMembers
assert.setEq
assert.soon
assert.soonNoExcept
assert.throws
assert.throwsWithCode
assert.time
assert.toLocaleString
assert.toString
assert.valueOf
assert.writeError
assert.writeErrorWithCode
assert.writeOK
```
