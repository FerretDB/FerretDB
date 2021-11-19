# Contributing

MangoDB is currently developed in either Linux or macOS, everything else is running inside Docker-Compose. 

You will need Go 1.18 (for [fuzzing](https://go.dev/blog/fuzz-beta) and [generics](https://go.dev/blog/generics-proposal)) that is not yet released.
[Compile it yourself](https://golang.org/doc/install/source) or use [`gotip download`](https://pkg.go.dev/golang.org/dl/gotip).

To verify your Go version:
```
$ go version
go version devel go1.18-[...]
```
## Helpful Commands

1. Install tools with `make init`.
2. Start the development environment with `make env-up`.
   This will start PostgreSQL and MongoDB; filling them with identical sets of test data.
3. You may then run tests in the other window with `make test`. 

You can see all available "make" commands with `make help`.

You can start MangoDB standalone with `make run`. This will start it in a development mode where all requests are handled by MangoDB but also routed to MongoDB. The differences in response are then logged and the MangoDB response is sent back to the client.

Run `mongosh` with `make mongosh`.
