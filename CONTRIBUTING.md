# Contributing

MangoDB development is done on the host running Linux or macOS, with everything else running inside Docker Compose.

You will need Go 1.18 (for [fuzzing](https://go.dev/blog/fuzz-beta) and [generics](https://go.dev/blog/generics-proposal)) that is not released yet.
[Compile it yourself](https://golang.org/doc/install/source) or use [`gotip download`](https://pkg.go.dev/golang.org/dl/gotip).
Verify Go version:
```
$ go version
go version devel go1.18-[...]
```

1. Install tools with `make init`.
2. Start the development environment with `make env-up`.
   This will start PostgreSQL and MongoDB, filling them with identical sets of test data.
3. Run tests in the other window with `make test`.
4. See all available targets with `make help`.
5. Start MangoDB with `make run`.
   That will start it in a development mode where all requests are handled by MangoDB and also routed to MongoDB.
   The response differences are then logged, and the MangoDB response is sent back to the client.
6. Run `monogsh` with `make mongosh`.
