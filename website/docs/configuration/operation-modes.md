---
sidebar_position: 2
slug: /configuration/operation-modes/ # referenced in README.md
---

# Operation modes

To simplify the development and debugging of FerretDB, we support different operation modes.
Operation modes specify how FerretDB handles incoming requests.
They are useful for testing, debugging, or bug reporting.

You can specify modes by using the `--mode` flag or `FERRETDB_MODE` variable,
which accept following types of values: `normal`, `proxy`, `diff-normal`, `diff-proxy`.

By default FerretDB always run on `normal` mode, which means that all client requests
are processed only by FerretDB and returned to the client.

## Proxy

Proxy is another MongoDB-compatible database, accessible from the machine.
You can specify its connection URL with `--proxy-addr` flag or with the `FERRETDB_PROXY_ADDR` variable.

To forward all requests to proxy and return them to the client, use `proxy` operation mode.

## Diff modes

Diff modes (`diff-normal`, `diff-proxy`) forward requests to both databases, and log the difference between them.

The `diff-normal` afterwards returns the response from FerretDB and `diff-proxy` - from the specified proxy handler.

Example diff output:

```sh
--- res header
+++ proxy header
@@ -1 +1 @@
-length:    87, id:    3, response_to:   69, opcode: OP_MSG
+length:    87, id:  645, response_to:   69, opcode: OP_MSG

Body diff:
--- res body
+++ proxy body
@@ -14,4 +14,4 @@
         "estimate": false,
-        "size": 0,
-        "numObjects": 0,
+        "size": 93,
+        "numObjects": 3,
         "millis": 0,
```
