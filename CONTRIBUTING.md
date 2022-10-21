# Contributing

Thank you for your interest in making FerretDB better!

## Finding something to work on

We are interested in all contributions, big or small, in code or documentation.
But unless you are fixing a very small issue like a typo,
we kindly ask you first to [create an issue](https://github.com/FerretDB/FerretDB/issues/new/choose),
to leave a comment on an existing issue if you want to work on it,
or to [join our Slack chat](./README.md#community) and leave a message for us there.
This way, you will get help from us and avoid wasted efforts if something can't be worked on right now
or someone is already working on it.

You can find a list of good issues for first-time contributors [there](https://github.com/FerretDB/FerretDB/contribute).

## Setting up the environment

### Requirements

The supported way of contributing to FerretDB is to modify and run it on the host (Linux, macOS, or Windows)
with PostgreSQL and other dependencies running inside Docker containers via Docker Compose.
On Linux, `docker` and `docker-compose` should be installed on the host.
On macOS and Windows, [Docker Desktop](https://www.docker.com/products/docker-desktop/) should be used.
On Windows, it should be [configured to use WSL 2](https://docs.docker.com/desktop/windows/wsl/) without any distro;
all commands should be run on the host.

You will need Go 1.19 or later on the host.
If your package manager doesn't provide it yet,
please install it from [go.dev](https://go.dev/dl/).

You will also need `git` installed; the version provided by your package manager should do.
On Windows, the simplest way to install it might be <https://gitforwindows.org>.

Finally, you will also need [git-lfs](https://git-lfs.github.com) installed and configured (`git lfs install`).

### Making a working copy

Fork the [FerretDB repository on GitHub](https://github.com/FerretDB/FerretDB/fork).
To have all the tags in the repository and what they point to, copy all branches by removing checkmark for `copy the main branch only` before forking.

After forking FerretDB on GitHub, you can clone the repository:

```sh
git clone git@github.com:<YOUR_GITHUB_USERNAME>/FerretDB.git
cd FerretDB
git remote add upstream https://github.com/FerretDB/FerretDB.git
```

To run development commands, you should first install the [`task`](https://taskfile.dev/) tool.
You can do this by changing the directory to `tools` (`cd tools`) and running `go generate -x`.
That will install `task` into the `bin` directory (`bin/task` on Linux and macOS, `bin\task.exe` on Windows).
You can then add `./bin` to `$PATH` either manually (`export PATH=./bin:$PATH` in `bash`)
or using something like [`direnv` (`.envrc` files)](https://direnv.net),
or replace every invocation of `task` with explicit `bin/task`.
You can also [install `task` globally](https://taskfile.dev/#/installation),
but that might lead to the version skew.

With `task` installed,
you should install development tools with `task init`
and download required Docker images with `task env-pull`.

If something does not work correctly,
you can reset the environment with `task env-reset`.

You can see all available `task` tasks with `task -l`.

## Contributing code

### Commands for contributing code

With `task` installed (see above), you may do the following:

1. Start the development environment with `task env-up`.
2. Run all tests in another terminal window with `task test`.
3. Start FerretDB with `task run`.
   This will start it in a development mode where all requests are handled by FerretDB, but also routed to MongoDB.
   The differences in response are then logged and the FerretDB response is sent back to the client.
4. Fill `values` collection in `test` database with data for experiments with `task env-data`.
5. Run `mongosh` with `task mongosh`.
   This allows you to run commands against FerretDB.
   For example, you can see what data was inserted by the previous command with `db.values.find()`.

### Code overview

The directory `cmd` provides commands implementation.
Its subdirectory `ferretdb` is the main FerretDB binary; others are tools for development.

The package `tools` uses ["tools.go" approach](https://github.com/golang/go/issues/25922#issuecomment-402918061) to fix tools versions.
They are installed into `bin/` by `cd tools; go generate -x`.

The `internal` subpackages contain most of the FerretDB code:

* `types` package provides Go types matching BSON types that don't have built-in Go equivalents:
  we use `int32` for BSON's int32, but `types.ObjectID` for BSON's ObjectId.
* `types/fjson` provides converters from/to FJSON for built-in and `types` types.
  FJSON adds some extensions to JSON for keeping object keys in order,
  preserving BSON type information in the values themselves, etc.
  It is used for logging of BSON values and wire protocol messages.
* `bson` package provides converters from/to BSON for built-in and `types` types.
* `wire` package provides wire protocol implementation.
* `clientconn` package provides client connection implementation.
  It accepts client connections, reads `wire`/`bson` protocol messages, and passes them to `handlers`.
  Responses are then converted to `wire`/`bson` messages and sent back to the client.
* `handlers` contains a common interface for backend handlers that they should implement.
  Handlers use `types` and `wire` packages, but `bson` package details are hidden.
* `handlers/common` contains code shared by different handlers.
* `handlers/dummy` contains a stub implementation of that interface that could be copied into a new package
  as a starting point for the new handlers.
* `handlers/pg` contains the implementation of the PostgreSQL handler.
* `handlers/pg/pjson` provides converters from/to PJSON for built-in and `types` types.
  PJSON adds some extensions to JSON for keeping object keys in order,
  preserving BSON type information in the values themselves, etc.
  It is used by `pg` handler.
* `handlers/tigris` contains the implementation of the Tigris handler.
* `handlers/tigris/tjson` provides converters from/to TJSON with JSON Schema for built-in and `types` types.
  BSON type information is preserved either in the schema (where possible) or in the values themselves.
  It is used by `tigris` handler.

Those packages are tested by "unit" tests that are placed inside those packages.
Some of them are truly hermetic and test only the package that contains them;
you can run those "short" tests with `go test -short` or `task test-unit-short`,
but that's typically not required.
Other unit tests use real databases;
you can run those with `task test-unit` after starting the environment as described above.

We also have a set of "integration" tests in the `integration` directory.
They use the Go MongoDB driver like a regular user application.
They could test target any MongoDB-compatible database (such as FerretDB or MongoDB itself) via a regular TCP port.
They also could test target in-process FerretDB instances
(meaning that integration tests start and stop them themselves) with a given handler.
Some tests (so-called compatibility or "compat" tests) connect to two systems ("target" and "compat") at the same time,
send the same queries to both, and compare results.
You can run them with:

* `task test-integration-pg` for in-process FerretDB with `pg` handler and MongoDB on port 37017 (as in our development environment);
* `task test-integration-tigris` for in-process FerretDB with `tigris` handler and MongoDB on port 37017;
* `task test-integration-mongodb` for MongoDB running on port 37017 only;
* or `task test-integration` to run all in parallel.

You may run all tests in parallel with `task test`.
If tests fail and the output is too confusing, try running them sequentially by using the commands above.

You can also run `task -C 1` to limit the number of concurrent tasks, which is useful for debugging.

Finally, since all tests just run `go test` with various arguments and flags under the hood,
you may also use all standard `go` tool facilities,
including [`GOFLAGS` environment variable](https://pkg.go.dev/cmd/go#hdr-Environment_variables).
For example, to run a single test case for `pg` handler with all subtests running sequentially,
you may use `env GOFLAGS='-run=TestName/TestCaseName -parallel=1' task test-integration-pg`.

In general, we prefer integration tests over unit tests,
tests using real databases over short tests
and real objects over mocks.

(You might disagree with our terminology for "unit" and "integration" tests;
let's not fight over it.)

We have an additional integration testing system in another repository: <https://github.com/FerretDB/dance>.

### Code style and conventions

Above everything else, we value consistency in the source code.
If you see some code that doesn't follow some best practice but is consistent,
please keep it that way;
but please also tell us about it, so we can improve all of it.
If, on the other hand, you see code that is inconsistent without apparent reason (or comment),
please improve it as you work on it.

Our code most of the standard Go conventions,
documented on [CodeReviewComments wiki page](https://github.com/golang/go/wiki/CodeReviewComments).
Some of our idiosyncrasies:

1. We use type switches over BSON types in many places in our code.
   The order of `case`s follows this order: <https://pkg.go.dev/github.com/FerretDB/FerretDB/internal/types#hdr-Mapping>
   It may seem random, but it is only pseudo-random and follows BSON spec: <https://bsonspec.org/spec.html>

### Submitting code changes

Before submitting a pull request, please make sure that:

1. Tests are added for new functionality or fixed bugs.
2. `task all` passes.
3. Comments are added or updated for all new and changed top-level declarations (functions, types, etc).
   Both exported and unexported declarations should have comments.
4. Comments are rendered correctly in the `task godocs` output.

## Contributing documentation

### Commands for contributing documentation

With `task` installed (see above), you may do the following:

1. Format and lint documentation with `task docs-fmt`.
2. Start Docusaurus development server with `task docs-dev`.
3. Build Docusaurus website with `task docs`.

### Submitting documentation changes

Before submitting a pull request, please make sure that:

1. Documentation is formatted, linted, and built with `task docs`.
