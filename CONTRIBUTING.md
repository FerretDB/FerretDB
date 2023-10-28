# Contributing

Thank you for your interest in making FerretDB better!

## Finding something to work on

We are interested in all contributions, big or small, in code or documentation.
But unless you are fixing a very small issue like a typo,
we kindly ask you first to [create an issue](https://github.com/FerretDB/FerretDB/issues/new/choose),
or leave a comment on an existing issue if you want to work on it.
This way, you will get help from us and avoid wasted efforts if something can't be worked on right now
or someone is already working on it.
You can also [join our Slack chat](./README.md#community) and leave a message for us in the `#dev` channel.

You can find a list of good first issues for contributors [there](https://github.com/FerretDB/FerretDB/contribute).
Once you have some experience with contributing to FerretDB,
feel free to pick any issue
[that is not assigned to anyone and doesn't have `not ready` label](https://github.com/FerretDB/FerretDB/issues?q=is%3Aissue+is%3Aopen+no%3Aassignee+-label%3A%22not+ready%22).
Still, please let us know as described above.

## Setting up the environment

### Requirements

The supported way of contributing to FerretDB is to modify and run it on the host (Linux, macOS, or Windows)
with PostgreSQL and other dependencies running inside Docker containers via Docker Compose.
On Linux, `docker` (with `docker compose` subcommand a.k.a. [Compose V2](https://docs.docker.com/compose/compose-v2/),
not old `docker-compose` tool) should be installed on the host.
On macOS and Windows, [Docker Desktop](https://www.docker.com/products/docker-desktop/) should be used.
On Windows, it should be [configured to use WSL 2](https://docs.docker.com/desktop/windows/wsl/) without any distro;
all commands should be run on the host.

You will need Go 1.21 or later on the host.
If your package manager doesn't provide it yet,
please install it from [go.dev](https://go.dev/dl/).

You will also need `git` installed; the version provided by your package manager should do.
On Windows, the simplest way to install it might be <https://gitforwindows.org>.

Finally, you will also need [git-lfs](https://git-lfs.github.com) installed and configured (`git lfs install`).

### Making a working copy

Fork the [FerretDB repository on GitHub](https://github.com/FerretDB/FerretDB/fork).
After that, you can clone the repository and add the upstream repository as a remote:

```sh
git clone git@github.com:<YOUR_GITHUB_USERNAME>/FerretDB.git
cd FerretDB
git remote add upstream https://github.com/FerretDB/FerretDB.git
git fetch --all --tags
```

The last command would fetch all Git tags from our repository that GitHub won't copy to fork by default.
`git describe` command uses them to determine the FerretDB version during the build process.

To run development commands, you should first install the [`task`](https://taskfile.dev/) tool.
You can do this by changing the directory to `tools` (`cd tools`) and running `go generate -x`.
That will install `task` into the `bin` directory (`bin/task` on Linux and macOS, `bin\task.exe` on Windows).
You can then add `./bin` to `$PATH` either manually (`export PATH=./bin:$PATH` in `bash`)
or using something like [`direnv` (`.envrc` files)](https://direnv.net),
or replace every invocation of `task` with explicit `bin/task`.
You can also [install `task` globally](https://taskfile.dev/#/installation),
but that might lead to the version skew.

With `task` installed,
you can see all available tasks by running `task -l` in the root of the repository
(not in the `tools` directory).

After that, you should install development tools with `task init`
and download required Docker images with `task env-pull`.

If something does not work correctly,
you can reset the environment with `task env-reset`.

### Building a production release binaries

To build a production release binaries, run `task build-production`.
The results will be saved `tmp/bin`.

### Setting a GITHUB_TOKEN

Some of our development tools require access to public information on GitHub
at a rate higher than allowed for unauthenticated requests.
Those tools will report a problem in this case.
It could be solved by creating a new classic or fine-graned personal access token
[there](https://github.com/settings/tokens).
No scopes are needed for classic tokens, not even `public_repo`.
For fine-graned tokens, only read-only access to public repositories is needed without any additional permissions.
After generating a token, set the `GITHUB_TOKEN` environment variable:

```sh
export GITHUB_TOKEN=ghp_XXX
```

or

```sh
export GITHUB_TOKEN=github_pat_XXX
```

## Reporting a bug

We appreciate reporting a bug to us.
To help us accurately identify the cause, we encourage you to include a pull request with test script.
Please write the test script in [build/legacy-mongo-shell/test.js](build/legacy-mongo-shell/test.js).
You can find an overview of the available assertions [here](build/legacy-mongo-shell/README.md).
Use these assertions to validate your test's assumptions and invariants.
You can also find an example of how to prepare a test script in
[build/legacy-mongo-shell/test.example.js](build/legacy-mongo-shell/test.example.js).

With `task` installed (see above), you may test your script using following steps:

1. Start the development environment with `task env-up`.
2. Start FerretDB with `task run`.
3. Run the test script with `task testjs`.

Please create a pull request and include the link of the pull request in the bug issue.

## Contributing code

### Commands for contributing code

With `task` installed (see above), you may do the following:

1. Start the development environment with `task env-up`.
2. Run all tests in another terminal window with `task test` (see [below](#running-tests)).
3. Start FerretDB with `task run`.
   This will start it in a development mode where all requests are handled by FerretDB, but also routed to MongoDB.
   The differences in response are then logged and the FerretDB response is sent back to the client.
4. Fill collections in `test` database with data for experiments with `task env-data`.
5. Run `mongosh` with `task mongosh`.
   This allows you to run commands against FerretDB.
   For example, you can see what data was inserted by the previous command with `db.values.find()`.

### Operation modes

FerretDB can run in multiple operation modes.
Operation mode specify how FerretDB handles incoming requests.
FerretDB supports four operation modes: `normal`, `proxy`, `diff-normal`, `diff-proxy`,
see [Operation modes documentation page](https://docs.ferretdb.io/configuration/operation-modes/) for more details.

By running `task run` FerretDB starts on `diff-normal` mode, which routes all of
the sent requests both to the FerretDB and MongoDB.
While running, it logs a difference between both returned responses,
but sends the one from FerretDB to the client.
If you want to get the MongoDB response, you can run `task run-proxy` to start FerretDB in `diff-proxy` mode.

### Code overview

The directory `cmd` provides commands implementation.
Its subdirectory `ferretdb` is the main FerretDB binary; others are tools for development.

The package `tools` uses ["tools.go" approach](https://github.com/golang/go/issues/25922#issuecomment-402918061) to fix tools versions.
They are installed into `bin/` by `cd tools; go generate -x`.

The `internal` subpackages contain most of the FerretDB code:

- `types` package provides Go types matching BSON types that don't have built-in Go equivalents:
  we use `int32` for BSON's int32, but `types.ObjectID` for BSON's ObjectId.
- `types/fjson` provides converters from/to FJSON for built-in and `types` types.
  It is used for logging of BSON values and wire protocol messages.
- `bson` package provides converters from/to BSON for built-in and `types` types.
- `wire` package provides wire protocol implementation.
- `clientconn` package provides client connection implementation.
  It accepts client connections, reads `wire`/`bson` protocol messages, and passes them to `handlers`.
  Responses are then converted to `wire`/`bson` messages and sent back to the client.
- `handlers` contains a common interface for backend handlers that they should implement.
  Handlers use `types` and `wire` packages, but `bson` package details are hidden.
- `handlers/common` contains code shared by different handlers.
- `handlers/sjson` provides converters from/to SJSON for built-in and `types` types.
  SJSON adds some extensions to JSON for keeping object keys in order,
  preserving BSON type information in the values themselves, etc.
  It is used by backends and old `pg` handler.
- `handlers/sqlite` contains the implementation of the SQLite handler.
  It is being converted into universal handler for all backends.

#### Running tests

`internal` packages are tested by "unit" tests that are placed inside those packages.
Some of them are truly hermetic and test only the package that contains them;
you can run those "short" tests with `go test -short` or `task test-unit-short`,
but that's typically not required.
Other unit tests use real databases;
you can run those with `task test-unit` after starting the environment as described above.

We also have a set of "integration" tests in the `integration` directory.
They use the Go MongoDB driver like a regular user application.
They could test any MongoDB-compatible database (such as FerretDB or MongoDB itself) via a regular TCP or TLS port or Unix socket.
They also could test in-process FerretDB instances
(meaning that integration tests start and stop them themselves) with a given backend.
Finally, some integration tests (so-called compatibility or "compat" tests) connect to two systems
("target" for FerretDB and "compat" for MongoDB) at the same time,
send the same queries to both, and compare results.
You can run them with:

- `task test-integration-postgresql` for in-process FerretDB with PostgreSQL backend and MongoDB;
- `task test-integration-mongodb` for MongoDB only, skipping compat tests;
- or `task test-integration` to run all in parallel.

You may run all tests in parallel with `task test`.
If tests fail and the output is too confusing, try running them sequentially by using the commands above.

You can also run `task -C 1` to limit the number of concurrent tasks, which is useful for debugging.

To run a subset of integration tests and test cases, you may use Task variable `TEST_RUN`.
For example, to run all tests related to the `getMore` command implementation for in-process FerretDB with PostgreSQL backend
you may use `task test-integration-postgresql TEST_RUN='(?i)GetMore'`.

Finally, since all tests just run `go test` with various arguments and flags under the hood
(for example, `TEST_RUN` just provides the value for the [`-run` flag](https://pkg.go.dev/cmd/go#hdr-Testing_flags)),
you may also use all standard `go` tool facilities,
including [`GOFLAGS` environment variable](https://pkg.go.dev/cmd/go#hdr-Environment_variables).
For example:

- to run all tests related to the `getMore` command implementation in-process FerretDB with PostgreSQL backend
  with all subtests running sequentially,
  you may use `env GOFLAGS='-parallel=1' task test-integration-postgresql TEST_RUN='(?i)GetMore'`;
- to run all tests for in-process FerretDB with SQLite backend
  with [Go execution tracer](https://pkg.go.dev/runtime/trace) enabled,
  you may use `env GOFLAGS='-trace=trace.out' task test-integration-sqlite`.

> **Note**
>
> It is not recommended to set `GOFLAGS` and other Go environment variables with `export GOFLAGS=...`
> or `go env -w GOFLAGS=...` because they are invisible and easy to forget about, leading to confusion.

In general, we prefer integration tests over unit tests,
tests using real databases over short tests
and real objects over mocks.

(You might disagree with our terminology for "unit" and "integration" tests;
let's not fight over it.)

We have an additional integration testing system in another repository: <https://github.com/FerretDB/dance>.

#### Observability in tests

Integration tests start a debug handler with pprof profiles and execution traces on a random port
(to allow running multiple test configurations in parallel).
They also send telemetry traces to the local Jaeger instance that can be accessed at <http://127.0.0.1:16686/>.

### Code style and conventions

Above everything else, we value consistency in the source code.
If you see some code that doesn't follow some best practice but is consistent,
please keep it that way;
but please also tell us about it, so we can improve all of it.
If, on the other hand, you see code that is inconsistent without apparent reason (or comment),
please improve it as you work on it.

Our code follows most of the standard Go conventions,
documented on [CodeReviewComments wiki page](https://github.com/golang/go/wiki/CodeReviewComments)
and some other pages such as [Spelling](https://github.com/golang/go/wiki/Spelling).
Some of our idiosyncrasies:

1. We use type switches over BSON types in many places in our code.
   The order of `case`s follows this order: <https://pkg.go.dev/github.com/FerretDB/FerretDB/internal/types#hdr-Mapping>
   It may seem random, but it is only pseudo-random and follows BSON spec: <https://bsonspec.org/spec.html>
2. We generally pass and return `struct`s by pointers.
   There are some exceptions like `types.Path` that have value semantics, but when in doubt – use pointers.
3. Code comments:
   - All top-level declarations, even unexported, should have documentation comments.
   - In documentation comments do not describe the name in terms of the name itself (`// Registry is a registry of …`).
     Use other words instead; often, they could add additional information and make reading more pleasant (`// Registry stores …`).
   - In code comments, in general, do not describe _what_ the code does: it should be clear from the code itself
     (and when it doesn't and the code is tricky, simplify it instead).
     Instead, describe _why_ the code does that if it is not clear from the surrounding context, names, etc.
     There is no need to add comments just because there are no comments if everything is already clear without them.
   - For code comments, write either
     sentence fragments (do not start it with a capital letter, do not end it with a dot, use the simplified grammar) for short notes
     or full sentences (do start them with capital letters, do end them with dots, do check their grammar) when a longer (2+ sentences)
     explanation is needed (and the code could not be simplified).

#### Integration tests conventions

We prefer our integration tests to be straightforward,
branchless (with a few, if any, `if` and `switch` statements),
and backend-independent.
Ideally, the same test should work for both FerretDB with all backends and MongoDB.
If that's impossible without some branching, use helpers exported from the `setup` package,
such us `FailsForFerretDB`, `SkipForMongoDB`, etc.
The bar for using other ways of branching, such as checking error codes and messages, is very high.
Writing separate tests might be much better than making a single test that checks error text.

Also, we should use driver methods as much as possible instead of testing commands directly via `RunCommand`.

We have a lot of existing helpers to convert the driver's `bson.D` values to our `*types.Document` values,
to compare them, etc.
In most cases, they should be used instead of (deprecated) `bson.D.Map()`,
`bson.Unmarshal`, decoding into a struct with `bson` tags, comparing fields one-by-one, etc.
The bar for adding new helpers is very high.
Please check all existing ones.

If there's a need to use any large number to test corner cases,
we create constants for them with explanation what do they represent, and refer to them.
For example:

```go
const doubleMaxPrec = float64(1<<53 - 1) // 9007199254740991.0:    largest double values that could be represented as integer exactly
```

#### Integration tests naming guidelines

1. Test names should include the name of the command being tested.
   For instance, `TestDistinct` for testing the distinct command.
2. Compatibility tests should have `Compat` in the name, following the command.
   For example, `TestDistinctCompat`.
3. If the test doesn't use driver method but runs a command directly via `RunCommand`,
   the suffix `Command` should be added.
   For example, `TestDistinctCommand`.
4. If the test is both compat and runs a command, the suffix `CommandCompat` should be added.
   For example, `TestInsertCommandCompat`.
5. If the file consists of compatibility tests, add the `_compat` suffix.
   For example, `distinct_compat_test.go`.
6. Test names should be descriptive and provide information about the functionality or condition being tested.
   If the test is checking for a specific error scenario, include the error scenario in the name.
7. Keep test names concise, avoiding overly cryptic names.
   Use abbreviations when appropriate.
8. Avoid including test data in the name to maintain clarity and prevent excessively long names.
9. Test case names should follow `TitleCase` capitalization style.
   No spaces, dashes or underscores should be used neither for test names nor for test case names.
10. Keep the concatenation of test name segments (test, subtests) within 64 characters
    to satisfy the maximum limit for database names.

### Submitting code changes

#### Before submitting PR

Before submitting a pull request, please make sure that:

1. Your changes are committed to a [new branch](https://docs.github.com/en/get-started/quickstart/github-flow)
   created from the [current state](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/syncing-a-fork)
   of our main branch.
2. Tests are added or updated for new functionality or fixed bugs.
   Typical test cases include:
   - happy paths;
   - dot notation for existing and non-existent paths;
   - edge cases for invalid or unexpected values or types.
3. Comments are added or updated for all new or changed code.
   Please add missing comments for all (both exported and unexported)
   new and changed top-level declarations (`package`, `var`, `const`, `func`, `type`).
   Please also check that formatting is correct in the `task godocs` output.
4. `task all` passes.

#### Submitting PR

1. If the pull request is related to some issues,
   please mention the issue number in the pull request **description** like `Closes #{issue_number}.`
   or `Closes org/repo#{issue_number}.`
   (You can just follow the pull request template).
   Please do not use URLs like `https://github.com/org/repo/issue/{issue_number}`
   or paths like `org/repo/issue/{issue_number}` even if they are rendered the same on GitHub.
   If you propose a tiny fix, there is no needed to create a new issue.
2. There is no need to use draft pull requests.
   If you want to get feedback on something you are working on,
   please create a normal pull request, even if it is not "fully" ready yet.
3. In the pull request review conversations,
   please either leave a new comment or resolve (close) the conversation,
   which ensures the other people can read all comments.
   But **do not** do that simultaneously.
4. During your development,
   commit messages (both titles and bodies) are not important and can be "WIP" or anything else.
   All commits are always squashed on merge by GitHub,
   so there is **no need** to squash them manually, amend them, and/or do force pushes.
   But notice that the autogenerated GitHub's squash commit's body
   **should be** manually replaced by "Closes #{issue_number}.".
5. Please don't forget to click
   ["re-request review" buttons](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/requesting-a-pull-request-review)
   once PR is ready for re-review.

If you have interest in becoming or are a long-term contributor,
please read [PROCESS.md](.github/PROCESS.md) for more details.

## Contributing documentation

### Commands for contributing documentation

With `task` installed (see above), you may do the following:

1. Format and lint documentation with `task docs-fmt`.
2. Start Docusaurus development server with `task docs-dev`.
3. Build Docusaurus website with `task docs`.

### Submitting documentation changes

Before submitting a pull request, please make sure that:

1. Your changes are committed to a [new branch](https://docs.github.com/en/get-started/quickstart/github-flow)
   created from the [current state](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/syncing-a-fork)
   of our main branch.
2. Documentation is formatted, linted, and built with `task docs`.
3. Documentation is written according to our [writing guide](https://docs.ferretdb.io/contributing/writing-guide/).
