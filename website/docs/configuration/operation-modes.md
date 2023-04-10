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

```diff
Header diff:
--- res header
+++ proxy header
@@ -1 +1 @@
-length:    63, id:   14, response_to:   24, opcode: OP_MSG
+length:    64, id:  229, response_to:   24, opcode: OP_MSG

Body diff:
--- res body
+++ proxy body
@@ -10,3 +10,3 @@
         ],
-        "you": "127.0.0.1:57079",
+        "you": "172.19.0.1:59824",
         "ok": {
```
