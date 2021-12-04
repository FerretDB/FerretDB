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
```
ls -s path_to_gotip /user/local/bin/go
```
Use `which gotip` to get the path to gotip.


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
