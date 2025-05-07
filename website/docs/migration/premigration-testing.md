---
sidebar_position: 2
---

# Pre-migration testing

To ensure a smooth and successful migration from MongoDB,
we offer several methods through which you can test your application with FerretDB.

## Performance testing

For performance testing and benchmarking, we recommend using the FerretDB production image or package rather than the development image.
For details on how to install FerretDB, [refer to the installation guide](../installation/ferretdb/docker.md).
Also, ensure to disable debug logging to avoid performance degradation.

Additional information on debugging and performance testing can be found on our [observability page](../configuration/observability.md).

## Operation modes

We offer multiple operation modes which help facilitate the testing of your application by enabling FerretDB to act as a proxy.
For more details, refer to the [operation modes](../configuration/operation-modes.md).

### Manual and automated testing with `diff-normal` mode

You can manually test your application or use integration tests, among other methods.
Afterward, you can inspect the differential output for errors or inconsistencies between responses that require your attention.

As an example, let us say that your application performs some complex query or operation and you'd like to test it in `diff-normal` mode.
You would do the following:

1. Start FerretDB in `diff-normal` mode.

   This can be achieved by setting the `--mode` [flag](../configuration/flags.md) or `FERRETDB_MODE` environment variable to `diff-normal`.
   By default, FerretDB starts in normal mode (`--mode=normal`/`FERRETDB_MODE=normal`).
   For more details, see [operation modes](../configuration/operation-modes.md).

   Ensure to specify `--listen-addr` and `--proxy-addr` flags or set the `FERRETDB_LISTEN_ADDR` and `FERRETDB_PROXY_ADDR` environment variables.
   Specify the address of your MongoDB instance for `--proxy-addr` flag or `FERRETDB_PROXY_ADDR` environment variable.
   [See docs for more details](../configuration/flags.md#interfaces). For example:

   ```sh
   ferretdb --mode=diff-normal \
     --proxy-addr=<mongodb-URI> \
     --listen-addr=<ferretdb-listen-address> \
     --postgresql-url=<postgres-connection>
   ```

   The `--listen-addr` flag or the `FERRERDB_LISTEN_ADDR` environment variable is set to `127.0.0.1:27017` by default.

2. Connect to the `--listen-addr` using the correct authentication credentials and then insert some documents.
3. Run a command to convert a collection to a capped collection.

   Please note that due to running in `diff-normal` mode, any error returned from FerretDB will be transmitted to the client, allowing us to promptly identify the issue.
   In the majority of cases, this does not necessitate additional scrutiny of the diff output.
   Nevertheless, if FerretDB does not handle the error, additional inspection becomes necessary.

   ```sh
   # run mongosh
   $ mongosh
   ```

   ```js
   // insert some documents
   db.locations.insertMany([
     { postId: '1', position: { type: 'Point', coordinates: [-73.97, 40.77] } },
     { postId: '2', position: { type: 'Point', coordinates: [-74.0, 40.75] } },
     { postId: '3', position: { type: 'Point', coordinates: [-73.95, 40.78] } },
     { postId: '4', position: { type: 'Point', coordinates: [-73.93, 40.76] } }
   ])

   // run the command
   db.runCommand({ convertToCapped: 'locations', size: 1024 })

   // the below error is returned to the client:
   // MongoServerError[CommandNotFound]: no such command: 'convertToCapped'
   ```

### Manual and automated testing with `diff-proxy` mode

Continuing with the same example above, we can further examine the diff output while in `diff-proxy` mode.

1. Run FerretDB in `diff-proxy` mode.
   This can again be achieved by using the `--mode` [flag](../configuration/flags.md) or by setting the `FERRETDB_MODE` environment variable to `diff-proxy`.
2. Follow the same instructions as the one for `diff-normal` above to run FerretDB in `diff-proxy` mode and re-run the command.

   ```js
   db.runCommand({ convertToCapped: 'locations', size: 1024 })
   ```

   ```text
   // the operation was handled by MongoDB, so the following response was returned:
   { ok: 1 }
   ```

In the diff output below, however, we have discovered that the command cannot be serviced by our application because the `convertToCapped` command is not implemented.

```text
--- res header
+++ proxy header
@@ -1 +1 @@
-length:   125, id:    5, response_to:   19, opcode: OP_MSG
+length:    38, id: 2621, response_to:   19, opcode: OP_MSG

Body diff:
--- res body
+++ proxy body
@@ -7,6 +7,3 @@
       `Document`: {
-        `ok`: 0.0,
-        `errmsg`: `no such command: 'convertToCapped'`,
-        `code`: 59,
-        `codeName`: `CommandNotFound`,
+        `ok`: 1.0,
       },
```

### Response metrics

Metrics are captured and written to standard output (`stdout`) upon exiting in [development builds](https://pkg.go.dev/github.com/FerretDB/FerretDB/v2/build/version#hdr-Development_builds).
This is a useful way to quickly determine what commands are not implemented for the client requests sent by your application.
In the metrics provided below, we can observe that the `convertToCapped` command was invoked once.

```text
# HELP ferretdb_client_accepts_total Total number of accepted client connections.
# TYPE ferretdb_client_accepts_total counter
ferretdb_client_accepts_total{error="0"} 11
# HELP ferretdb_client_duration_seconds Client connection lifetime in seconds.
# TYPE ferretdb_client_duration_seconds histogram
ferretdb_client_duration_seconds_bucket{error="0",le="1"} 2
ferretdb_client_duration_seconds_bucket{error="0",le="5"} 2
ferretdb_client_duration_seconds_bucket{error="0",le="10"} 2
ferretdb_client_duration_seconds_bucket{error="0",le="30"} 2
ferretdb_client_duration_seconds_bucket{error="0",le="60"} 4
ferretdb_client_duration_seconds_bucket{error="0",le="300"} 4
ferretdb_client_duration_seconds_bucket{error="0",le="600"} 4
ferretdb_client_duration_seconds_bucket{error="0",le="1800"} 4
ferretdb_client_duration_seconds_bucket{error="0",le="+Inf"} 4
ferretdb_client_duration_seconds_sum{error="0"} 101.054557752
ferretdb_client_duration_seconds_count{error="0"} 4
# HELP ferretdb_client_requests_total Total number of requests.
# TYPE ferretdb_client_requests_total counter
ferretdb_client_requests_total{command="aggregate",opcode="OP_MSG"} 1
ferretdb_client_requests_total{command="atlasVersion",opcode="OP_MSG"} 1
ferretdb_client_requests_total{command="buildInfo",opcode="OP_MSG"} 1
ferretdb_client_requests_total{command="convertToCapped",opcode="OP_MSG"} 1
ferretdb_client_requests_total{command="getLog",opcode="OP_MSG"} 1
ferretdb_client_requests_total{command="getParameter",opcode="OP_MSG"} 1
ferretdb_client_requests_total{command="hello",opcode="OP_MSG"} 12
ferretdb_client_requests_total{command="ping",opcode="OP_MSG"} 3
ferretdb_client_requests_total{command="unknown",opcode="OP_QUERY"} 9
# HELP ferretdb_client_responses_total Total number of responses.
# TYPE ferretdb_client_responses_total counter
ferretdb_client_responses_total{argument="unknown",command="aggregate",opcode="OP_MSG",result="ok"} 1
ferretdb_client_responses_total{argument="unknown",command="atlasVersion",opcode="OP_MSG",result="ok"} 1
ferretdb_client_responses_total{argument="unknown",command="buildInfo",opcode="OP_MSG",result="ok"} 1
ferretdb_client_responses_total{argument="unknown",command="convertToCapped",opcode="OP_MSG",result="ok"} 1
ferretdb_client_responses_total{argument="unknown",command="getLog",opcode="OP_MSG",result="ok"} 1
ferretdb_client_responses_total{argument="unknown",command="getParameter",opcode="OP_MSG",result="ok"} 1
ferretdb_client_responses_total{argument="unknown",command="hello",opcode="OP_MSG",result="ok"} 12
ferretdb_client_responses_total{argument="unknown",command="ping",opcode="OP_MSG",result="ok"} 3
ferretdb_client_responses_total{argument="unknown",command="unknown",opcode="OP_REPLY",result="ok"} 9
```

### Other tools

We also have a fork of the Amazon DocumentDB Compatibility Tool [here](https://github.com/FerretDB/amazon-documentdb-tools/tree/master/compat-tool).
The tool examines files to identify queries that use unsupported operators in FerretDB.
Please note that this tool is not highly accurate and may generate inaccurate reports, as it does not parse query syntax with contextual information about the originating command.
For example, an unsupported operator might appear within a `find` or `aggregate` command, which the tool does not differentiate.
Note that we also mark operators as unsupported if they are not supported in _all_ commands, which could result in false negatives.

Running the tool to check FerretDB compatibility:

```sh
# clone the repository and run the compat-tool
$ git clone https://github.com/FerretDB/amazon-documentdb-tools.git && cd amazon-documentdb-tools/compat-tool
$ python3 compat.py --directory=/path/to/myapp --version=FerretDB
```
