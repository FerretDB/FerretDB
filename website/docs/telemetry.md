---
sidebar_position: 7
slug: /telemetry/
---

# Telemetry reporting

FerretDB collects anonymous usage data and sends them to our telemetry service ([FerretDB Beacon](https://beacon.ferretdb.io)),
which helps us understand its usage, and how we can further increase compatibility and enhance our product.
It also enables us to provide you information about available updates.

Your privacy is important to us, and we understand how sensitive data collection can be,
which is why we are not collecting any personally-identifying information
or share any of the data with third parties.

The following data will be collected:

* FerretDB version
* Random instance UUID
* [Autonomous system](https://en.wikipedia.org/wiki/Autonomous_system_(Internet)) number,
  cloud provider region, or country derived from IP address (but the IP address itself)
* Uptime
* Backend (PostgreSQL or Tigris) version
* Build configuration and installation type (Docker, package, self-built)
* Command statistics:
  * protocol operation codes (e.g. `OP_MSG`, `OP_QUERY`);
  * command names (e.g. `update`, `aggregate`);
  * arguments (e.g. `$set`, `$count (stage)`).
  * error codes (e.g. `NotImplemented`, `InternalError`; or `ok`);

:::info
Argument values, data field names, successful responses, or error messages are never collected.
:::

## Version notification

When a FerretDB update is available, the telemetry service sends periodic reports containing information about the latest FerretDB version.

This information is logged in the server logs, `startupWarnings`, and `serverStatus` command output.

While you may not upgrade to the latest release immediately,
ensure that you update early to take advantage of recent bug fixes, new features, and performance improvements.

## Configure telemetry

The telemetry service has three state settings: `enabled`, `disabled`, and `undecided` (default).

:::info
When the state setting is `undecided`, users have a delay period of one hour after startup before telemetry data is reported.
:::

### Disable telemetry

We urge you not to disable this service, as its insights will help us enhance our software.

While we are grateful for these usage insights, we understand that not everyone is comfortable with sending them.

:::caution
If you disable telemetry, automated version checks and information on updates will not be available.
:::

Telemetry can be disabled using any of the following options:

1. Pass the command-line flag `--telemetry` to the FerretDB executable with value:
   `0`, `f`, `false`, `n`, `no`, `off`, `disable`, `disabled`, `optout`, `opt-out`, `disallow`, `forbid`:

   ```sh
   --telemetry=disable
   ```

2. Set the environment variable `FERRETDB_TELEMETRY`:

   ```sh
   export FERRETDB_TELEMETRY=disable
   ```

3. Set the `DO_NOT_TRACK` environment variable with any of the following values:
   `1`, `t`, `true`, `y`, `yes`, `on`, `enable`, `enabled`:

   ```sh
   export DO_NOT_TRACK=true
   ```

4. Rename FerretDB executable to include a `donottrack` string.

   :::caution
   If telemetry is disabled using this option, you cannot use the `--telemetry` flag or environment variables
   until the `donottrack` string is removed.
   :::

5. Use the `db.disableFreeMonitoring()` command on runtime.

   ```js
   db.disableFreeMonitoring()
   ```

   :::caution
   If the telemetry is set via a command-line flag, an environment variable or a filename, it's not possible
   to modify its state via command.
   :::

### Enable Telemetry

If telemetry is disabled, enable telemetry with the command-line flag `--telemetry` and assign any of these values to it:
`1`, `t`, `true`, `y`, `yes`, `on`, `enable`, `enabled`, `optin`, `opt-in`, `allow`:

```sh
--telemetry=enable
```

You can also use `FERRETDB_TELEMETRY` environment variable with same values.

If telemetry is disabled with a `donottrack` string in the executable,
remove the `donottrack` string to use the command-line flag and values again.

It's also possible to enable telemetry on runtime via `db.enableFreeMonitoring()` command.
