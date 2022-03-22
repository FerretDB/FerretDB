# Contributing

FerretDB is currently developed in either Linux or macOS, everything else is running inside Docker Compose.

You will need Go 1.18 as FerretDB extensively uses ([fuzzing](https://go.dev/doc/tutorial/fuzz)) and [generics](https://go.dev/doc/tutorial/generics)).
If your package manager does not provide it yet, please install it from [go.dev](https://go.dev/dl/).

## Cloning the Repository

After forking FerretDB on GitHub, you can clone the repository:

```sh
git clone git@github.com:<YOUR_GITHUB_USERNAME>/FerretDB.git
cd FerretDB
git remote add upstream https://github.com/FerretDB/FerretDB.git
```

## Helpful Commands

In order to run development commands you should get [task](https://taskfile.dev/).
You can do this with `cd tools; go generate -x`.
After this `task` will be available with `bin/task` on Linux and `bin\task.exe` on Windows.

1. Install development tools with `bin/task init`.
2. Start the development environment with `bin/task env-up`.
   This will start PostgreSQL and MongoDB; filling them with identical sets of test data.
3. You may then run tests in another window with `bin/task test`.
4. You can start FerretDB with `bin/task run`.
   This will start it in a development mode where all requests are handled by FerretDB, but also routed to MongoDB.
   The differences in response are then logged and the FerretDB response is sent back to the client.
5. Run `mongosh` with `bin/task mongosh`.
   This allows you to run commands against FerretDB.

You can see all available "task" commands with `bin/task -l`.

## Code Overview

Package `cmd` provides commands implementation. `ferretdb` is the main FerretDB binary; others are tools for development.

Package `tools` uses "tools.go" approach to fix tools versions. They are installed into `bin/` by `cd tools; go generate -x`.

`internal` subpackages contain most of the FerretDB code:
* `types` package provides Go types matching BSON types that don't have built-in Go equivalents: we use `int32` for BSON's int32, but `types.ObjectID` for BSON's ObjectId.
* `fjson` provides converters from/to FJSON for built-in and `types` types.
  FJSON adds some extensions to JSON for keeping object keys in order, preserving BSON type information, etc.
  FJSON is used by `jsonb1` handler/storage.
* `bson` package provides converters from/to BSON for built-in and `types` types.
* `wire` package provides wire protocol implementation.
* `clientconn` package provides client connection implementation.
  It accepts client connections, reads `wire`/`bson` protocol messages, and passes them to `handlers`.
  Responses are then converted to `wire`/`bson` messages and sent back to the client.
* `handlers` handle protocol commands.
  They use `fjson` package for storing data in PostgreSQL in jsonb columns, but they don't use `bson` package – all data is represented as built-in and `types` types.


## Useful links

* https://www.postgresql.org/docs/current/functions-json.html
* https://www.postgresql.org/docs/current/datatype-json.html
* https://doxygen.postgresql.org/jsonb_8c_source.html
* https://www.postgresql.org/docs/current/indexes-expressional.html
* https://www.postgresql.org/docs/current/pageinspect.html
* escaping https://github.com/lib/pq/issues/782
* jsonb https://www.postgresql.eu/events/pgconfeu2019/sessions/session/2555/slides/221/jsonpath-pgconfeu-2019.pdf
* jsnpath guide https://github.com/obartunov/sqljsondoc/blob/master/jsonpath.md

## Agreements

* in where and projection, let functions be functions, not methods
* use placeholders, must not use sanitizers / bare values
* formatting a message should be caller's responsibility