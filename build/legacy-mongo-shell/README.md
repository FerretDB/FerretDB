# Assertions in the `mongo` shell

See [Assertions 101](https://github.com/mongodb/mongo/wiki/Write-Tests-For-Server-Code#assertions-101) for a very brief overview.

The legacy `mongo` shell uses its own assertions which are defined [here](https://github.com/mongodb/mongo/blob/master/src/mongo/shell/assert.js).
These deviate from `mongosh`, which uses the standard Node.js [assert](https://nodejs.org/api/assert.html) module.
When you write a small reproducible test and call `task testjs` the legacy `mongo` shell will be invoked.

A useful function is `assert.commandFailedWithCode` which asserts that the command failed with the expected code as the name implies.
One should pass the result of a call to the `db.runCommand()` helper as this provides a result type that the function can parse.
This is the preferred method to issue database commands, as it provides a consistent interface between the shell and drivers.

For example, to check if the `findAndModify` command failed with the error code `ImmutableField` you would do the following:

```js
> const res = db.runCommand({findAndModify: "foo", query: {}, update: {$set: {_id: 1}}});
> assert.commandFailedWithCode(res, ErrorCodes.ImmutableField);
```

It is not always necessary use the `db.runCommand()` helper as some write methods wrap a `writeError` which the helpers can parse.
For example, `insert`, `update`, and `remove` will all return a `WriteResult` so the function can parse the result and look for a `writeError`.

See the `WriteResult` object below:

```js
// WriteResult object
> assert.commandWorked(db.foo.insert({a: 1}));
WriteResult({ "nInserted" : 1 })
> const res = db.foo.insert({a: 1});
> res.hasWriteError();
false
> Object.keys(res);
[
  "ok",
  "nInserted",
  "nUpserted",
  "nMatched",
  "nModified",
  "nRemoved",
  "getUpsertedId",
  "getRawResponse",
  "getWriteError",
  "hasWriteError",
  "getWriteConcernError",
  "hasWriteConcernError",
  "tojson",
  "toString",
  "shellPrint"
]
> res instanceof WriteResult
true
>
```

The `find()` method will return an object, the properties of which will not be parsed by the various assert functions:

```js
> const res = db.foo.find({a: 1});
> res._filter
{ "a" : 1 }
> res._db
test
> res._batchSize
0
// all properties are private but the shell will iterate the cursor object when called
> Object.keys(res);
[
  "_mongo",
  "_db",
  "_collection",
  "_ns",
  "_filter",
  "_projection",
  "_limit",
  "_skip",
  "_batchSize",
  "_options",
  "_additionalCmdParams",
  "_cursor",
  "_numReturned"
]
>
```

But a `runCommand()` result will return the correct object that can be parsed:

```js
> // use a runCommand instead
> const res = db.runCommand({find: "foo", filter: {a: 1}});
> res.ok
1
> res._commandObj
{
  "find" : "foo",
  "filter" : {
    "a" : 1
  },
  "lsid" : {
    "id" : UUID("cb64879f-6f69-4656-a38e-ce6dbe3ccebc")
  }
}
> assert.commandFailed(res);
uncaught exception: Error: command worked when it should have failed: {
  "cursor" : {
    "firstBatch" : [
      {
        "_id" : ObjectId("6458ec7adb858f891c0b8c68"),
        "a" : 1
      },
      {
        "_id" : ObjectId("6458ec9ddb858f891c0b8c6a"),
        "a" : 1
      }
    ],
    "id" : NumberLong(0),
    "ns" : "test.foo"
  },
  "ok" : 1
} :
_getErrorWithCode@src/mongo/shell/utils.js:24:13
doassert@src/mongo/shell/assert.js:18:14
_assertCommandFailed@src/mongo/shell/assert.js:819:25
assert.commandFailed@src/mongo/shell/assert.js:877:16
@(shell):1:8
```

## Some useful functions

`assert.eq(a, b, msg)`

throws if two values are not equal (tested without strict equality).

`assert.isnull(what, msg)`

throws if `what` is not null.

`assert.commandFailed(res, msg)`

throws if the command did not fail.

`assert.commandWorked(res)`

throws if the result contained an error.

`assert.commandFailedWithCode(res, expectedCode, msg)`

throws if the command did not fail with the expected code.

`assert.sameMembers(aArr, bArr, msg, compareFn = _isDocEq)`

throws if the two arrays do not have the same members, in any order.
By default, nested arrays must have the same order to be considered equal.
Optionally accepts a `compareFn` to compare values instead of using `docEq`.

`assert.writeOK(res, msg)`

throws if write result contained an error.

`assert.docEq(expectedDoc, actualDoc, msg)`

throws if `actualDoc` object is not equal to `expectedDoc` object.
The order of fields
(properties) within objects is disregarded.
Throws if object representation in BSON exceeds 16793600 bytes.

`assert.retry(func, msg, num_attempts, intervalMS)`

calls the given function `func` repeatedly at time intervals specified by
`intervalMS` (milliseconds) until either `func()` returns true or the number of
attempted function calls is equal to `num_attempts`.
Throws an exception with
message `msg` after all attempts are used up.
If no `intervalMS` argument is passed, it defaults to 0.

## All assert functions

```js
assert.adminCommandWorkedAllowingNetworkError
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
assert.includes
assert.isnull
assert.lt
assert.lte
assert.neq
assert.noAPIParams
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

## Error Codes

`ErrorCodes` is an object that is generated from various source files.
It provides error names that correspond to their respective codes.

```js
// tab in the shell after typing the below to display all 443 error codes
> ErrorCodes.
Display all 443 possibilities? (y or n)
ErrorCodes.APIDeprecationError                                          ErrorCodes.NetworkInterfaceExceededTimeLimit
ErrorCodes.APIMismatchError                                             ErrorCodes.NetworkTimeout
ErrorCodes.APIStrictError                                               ErrorCodes.NewReplicaSetConfigurationIncompatible
ErrorCodes.APIVersionError                                              ErrorCodes.NoConfigPrimary
ErrorCodes.AlarmAlreadyFulfilled                                        ErrorCodes.NoMatchParseContext
ErrorCodes.AlreadyInitialized                                           ErrorCodes.NoMatchingDocument
ErrorCodes.AmbiguousIndexKeyPattern                                     ErrorCodes.NoProgressMade
ErrorCodes.AtomicityFailure                                             ErrorCodes.NoProjectionFound
ErrorCodes.AuditingNotEnabled                                           ErrorCodes.NoQueryExecutionPlans
ErrorCodes.AuthSchemaIncompatible                                       ErrorCodes.NoReplicationEnabled
ErrorCodes.AuthenticationAbandoned                                      ErrorCodes.NoShardingEnabled
ErrorCodes.AuthenticationFailed                                         ErrorCodes.NoSuchKey
ErrorCodes.AuthenticationRestrictionUnmet                               ErrorCodes.NoSuchReshardCollection
ErrorCodes.BSONObjectTooLarge                                           ErrorCodes.NoSuchSession
ErrorCodes.BackgroundOperationInProgressForDatabase                     ErrorCodes.NoSuchTenantMigration
ErrorCodes.BackgroundOperationInProgressForNamespace                    ErrorCodes.NoSuchTransaction
ErrorCodes.BackupCursorOpenConflictWithCheckpoint                       ErrorCodes.NodeNotElectable
ErrorCodes.BadPerfCounterPath                                           ErrorCodes.NodeNotFound
ErrorCodes.BadValue                                                     ErrorCodes.NonConformantBSON
ErrorCodes.BalancerInterrupted                                          ErrorCodes.NonExistentPath
ErrorCodes.BrokenPromise                                                ErrorCodes.NonRetryableTenantMigrationConflict
ErrorCodes.CallbackCanceled                                             ErrorCodes.NotAReplicaSet
ErrorCodes.CanRepairToDowngrade                                         ErrorCodes.NotARetryableWriteCommand
ErrorCodes.CannotApplyOplogWhilePrimary                                 ErrorCodes.NotExactValueField
ErrorCodes.CannotBackfillArray                                          ErrorCodes.NotImplemented
ErrorCodes.CannotBackup                                                 ErrorCodes.NotPrimaryNoSecondaryOk
ErrorCodes.CannotBuildIndexKeys                                         ErrorCodes.NotPrimaryOrSecondary
ErrorCodes.CannotConvertIndexToUnique                                   ErrorCodes.NotSecondary
ErrorCodes.CannotCreateCollection                                       ErrorCodes.NotSingleValueField
ErrorCodes.CannotCreateIndex                                            ErrorCodes.NotWritablePrimary
ErrorCodes.CannotDowngrade                                              ErrorCodes.NotYetInitialized
ErrorCodes.CannotDropShardKeyIndex                                      ErrorCodes.OBSOLETE_BalancerLostDistributedLock
```

`ErrorCodeStrings` is an object with the reverse mapping of the above.

```js
> ErrorCodeStrings[0]
OK
> ErrorCodeStrings[2]
BadValue
> ErrorCodeStrings[66]
ImmutableField
```
