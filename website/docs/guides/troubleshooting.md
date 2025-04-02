---
sidebar_position: 5
slug: /troubleshooting/ # referenced in code in error messages
description: Learn about common issues and how to resolve them.
---

# Troubleshooting

If you experience issues with FerretDB, this troubleshooting guide will help you resolve the most common ones.

## Connectivity

<!-- Do not change header above as it is referenced in code in error messages -->

Do you have trouble setting up or connecting to FerretDB?
Find solutions to common connectivity issues below.

### Extension initialization

If you get an error when initializing PostgreSQL with the DocumentDB extension in Docker,
it may be due to an existing PostgreSQL data directory or volume.
This error occurs because the previous PostgreSQL data directory was created without the DocumentDB extension.

Log error may look like this:

```text
schema "documentdb_api" does not exist
```

To resolve this issue, delete the existing PostgreSQL data directory if unused
or change the data directory path in your Docker setup.
For example, if the path to the data directory of your PostgreSQL with DocumentDB extension instance is `./data`,
change it to `./postgres-data`.
You may need to export or migrate your data to the new PostgreSQL data directory.
Follow our [migration guide](../migration/migrating-from-mongodb.md) for more details.

For more details on setting up PostgreSQL with the DocumentDB extension in Docker,
see the [Docker installation guide](../installation/documentdb/docker.md).

### The authentication mechanism is not enabled

If you get an error when connecting to FerretDB with the `PLAIN` authentication mechanism
(e.g. `mongodb://username:password@127.0.0.1:27017/ferretdb?authMechanism=PLAIN`),
it is because `PLAIN` authentication is no longer supported in FerretDB v2.x.

Log error may look like this:

```text
Received authentication for mechanism %s which is not enabled
```

Note that FerretDB v2.x uses the `SCRAM-SHA-256` authentication mechanism,
and authentication is enabled by default.
To resolve this issue, connect to FerretDB without specifying the `PLAIN` mechanism in the connection string
(e.g. `mongodb://username:password@127.0.0.1:27017/`).

Learn more about [FerretDB authentication](../security/authentication.md).

### Salt length

<!-- Do not change header above as it is referenced in code in error messages -->

## Compatibility

For any compatibility issues or concerns,
read our [pre-migration testing guide](../migration/premigration-testing.md).
The guide will help you identify any potential compatibility issues before migrating your data to FerretDB.

## Performance

If you experience performance issues or have concerns about your FerretDB setup,
debugging and observability tools can help.
Our [observability guide](../configuration/observability.md) provides insights into logging,
OpenTelemetry traces, debug handlers, metrics, and health probes,
which can help diagnose these issues and optimize performance effectively.

## Other issues

If your issues persist or you encounter other problems,
please check your logs for details and share them with us on any of
[our community channels](../introduction.md#community) to get help resolving them.
