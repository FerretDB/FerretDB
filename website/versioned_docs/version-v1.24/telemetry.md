---
sidebar_position: 12
slug: /telemetry/ # referenced in many places; must not change
---

# Telemetry reporting

FerretDB collects basic anonymous usage data and sends it to our telemetry service ([FerretDB Beacon](https://beacon.ferretdb.com)).
It helps us understand FerretDB usage and how to increase compatibility and enhance our product further.
It also enables us to provide you with information about available updates.

Your privacy is important to us, and we understand how sensitive data collection can be.
We assure you that we are not collecting any personally identifying information.
We will never share or sell raw telemetry data to third parties.
We may share our findings and statistics based on telemetry data with third parties and the general public.

The following data is collected:

- Random FerretDB instance UUID
- [Autonomous system](<https://en.wikipedia.org/wiki/Autonomous_system_(Internet)>) number,
  cloud provider region and country derived from the IP address (but not the IP address itself)
- FerretDB version
- Backend (PostgreSQL or SQLite) version
- Installation type (Docker, package, cloud provider marketplace, self-built)
- Build configuration (Go version, build flags and tags)
- Uptime
- Command statistics:
  - protocol operation codes (e.g. `OP_MSG`, `OP_QUERY`);
  - command names (e.g., `find`, `aggregate`);
  - arguments (e.g., `sort`, `$count`);
  - error codes (e.g., `NotImplemented`, `InternalError`; or `ok`).

:::info
Argument values, data field names, successful responses, or error messages are never collected.
:::

## Version notifications

When a FerretDB update is available,
the telemetry service responds with information about the latest FerretDB version.
This information is logged in server logs and available via the `getLog` command with the `startupWarnings` argument, making it visible when connecting with various tools such as `mongosh`.

While you may not upgrade to the latest release immediately,
ensure you update early to take advantage of recent bug fixes, new features, and performance improvements.

## Configuration

The telemetry reporter has three state settings: `enabled`, `disabled`, and `undecided` (the default).
The latter acts as if it is `enabled` with two differences:

- When `enabled`, the first report is sent right after FerretDB starts.
  If `undecided`, the first report is delayed by one hour.
  That should give you enough time to disable it if you decide to do so.
- Similarly, when `enabled`, the last report is sent right before FerretDB shuts down.
  That does not happen when `undecided`.

:::info
`undecided` state does not automatically change into `enabled` or `disabled`.
Explicit user action is required (see below) to change an `undecided` state to `enabled` or `disabled`.
:::

Telemetry reporting is always disabled for [embedded FerretDB](https://pkg.go.dev/github.com/FerretDB/FerretDB/ferretdb)
and can't be configured.

### Disable telemetry

We urge you not to disable the telemetry reporter, as its insights will help us enhance our software.

While we are grateful for these usage insights, we understand that not everyone is comfortable with sending them.

:::caution
If you disable telemetry, automated version checks and information on updates will not be available.
:::

Telemetry can be disabled using any of the following options:

1. Pass the command-line flag `--telemetry` to the FerretDB executable with value:
   `0`, `f`, `false`, `n`, `no`, `off`, `disable`, `disabled`, `optout`, `opt-out`, `disallow`, `forbid`.

   ```sh
   --telemetry=disable
   ```

2. Set the environment variable `FERRETDB_TELEMETRY` with the same value as above:

   ```sh
   export FERRETDB_TELEMETRY=disable
   ```

3. Set the `DO_NOT_TRACK` environment variable with any of the following values:
   `1`, `t`, `true`, `y`, `yes`, `on`, `enable`, `enabled`.

   ```sh
   export DO_NOT_TRACK=true
   ```

4. Rename the FerretDB executable to include a `donottrack` string.

   :::caution
   If telemetry is disabled using this option, you cannot use the `--telemetry` flag or environment variables
   until the `donottrack` string is removed from the executable.
   :::

5. Use the `db.disableFreeMonitoring()` command on runtime.

   ```js
   db.disableFreeMonitoring()
   ```

   :::caution
   If the telemetry is set via a command-line flag, an environment variable, or a filename, it's not possible
   to modify its state via command.
   :::

### Enable telemetry

Telemetry can be explicitly enabled (see [above](#configuration)) with the command-line flag `--telemetry`
by setting one of the values:
`1`, `t`, `true`, `y`, `yes`, `on`, `enable`, `enabled`, `optin`, `opt-in`, `allow`.

```sh
--telemetry=enable
```

You can also use `FERRETDB_TELEMETRY` environment variable with the same values as above
or on runtime via `db.enableFreeMonitoring()` command.

```sh
export FERRETDB_TELEMETRY=enable
```

```js
db.enableFreeMonitoring()
```

One case when explicitly enabling telemetry is useful is if you want to help us improve compatibility
with your application by running its integration tests or manually testing it.
If you leave the telemetry state undecided and your test lasts less than an hour,
we will not have data about unimplemented commands and errors.

If you want to help us with that, please do the following:

1. Start FerretDB with [debug logging](configuration/flags.md) and telemetry explicitly enabled.
   Confirm that telemetry is enabled from the logs.
2. Test your application manually or with integration tests.
3. Gracefully stop FerretDB with `SIGTERM` or `docker stop` (not with `SIGKILL` or `docker kill`).
4. Optionally, locate instance UUID in the `state.json` file in the state directory
   (`/state` for Docker, current directory otherwise) and send it to us.
   That would allow us to locate your data and understand what FerretDB functionality
   should be implemented or fixed to improve compatibility with your application.
