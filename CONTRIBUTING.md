# Contributing

FerretDB is currently developed in either Linux or macOS, everything else is running inside Docker Compose.

You will need Go 1.18 (for [fuzzing](https://go.dev/blog/fuzz-beta) and [generics](https://go.dev/blog/generics-proposal)) that is not yet released.
[Compile it yourself](https://golang.org/doc/install/source) or use [`gotip download`](https://pkg.go.dev/golang.org/dl/gotip).

To verify your Go version:
```
$ go version
go version devel go1.18-[...]
```
**Note:** If using gotip and `go version` does not return version 1.18, you may need to symbolic link `gotip` to `go`.
```sh
$ ln -s `which gotip` /usr/local/bin/go
```
Use `which gotip` to get the path to gotip.

## Cloning the Repository

After forking FerretDB you can clone the repository - for the best practices use the following instructions:

```sh
$ git clone git@github.com:FerretDB/FerretDB.git
$ cd FerretDB
$ git remote rename origin upstream
$ git remote set-url --push upstream NO_PUSH
$ git remote add origin git@github.com:<YOUR_GITHUB_USERNAME>/FerretDB.git
```

## Helpful Commands

1. Install tools with `make init`.
2. Start the development environment with `make env-up`.
   This will start PostgreSQL and MongoDB; filling them with identical sets of test data.
3. You may then run tests in another window with `make test`.
4. You can start FerretDB with `make run`.
   This will start it in a development mode where all requests are handled by FerretDB, but also routed to MongoDB.
   The differences in response are then logged and the FerretDB response is sent back to the client.
5. Run `mongosh` with `make mongosh`.
   This allows you to run commands against FerretDB.

You can see all available "make" commands with `make help`.

## Code Overview

Package `cmd` provides commands implementation. `ferretdb` is the main FerretDB binary; others are tools for development.

Package `tools` uses "tools.go" approach to fix tools versions. They are installed into `bin/` by `make init`.

`internal` subpackages contain most of the FerretDB code:
* `types` package provides Go types matching BSON types that don't have built-in Go equivalents: we use `int32` for BSON's int32, but types.ObjectID for BSON's ObjectId.
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
