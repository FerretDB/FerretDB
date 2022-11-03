---

sidebar_position: 1

---

# Telemetry reporting

FerretDB collects anonymous usage data to help us achieve compatibility and enhance our product.

It enables us to provide automated checks and updates on available versions.

Your privacy is important to us, and we understand how sensitive data collection can be, which is why we will not collect any personally-identifying information or share any of the data with third parties.

The following usage data will be collected:

* Installation UUID
* FerretDB version
* Backend - PostgreSQL or Tigris
* Build configuration - including build tags and installation type (Docker, package, self-built)
* Query errors - error code, command name, and query operator name
* Autonomous system number, cloud provider region, and country
* Uptime performance (seconds) - the amount of time it takes to execute a particular FerretDB command

## Version notification

The telemetry service sends periodic reports containing information about the latest FerretDB version.

This information is recorded in the server logs, startupWarnings, and serverStatus outputs.

While you may not upgrade to the latest release immediately, ensure that you update early to take advantage of recent bug fixes, new features, and performance improvements.

## Configure telemetry

The telemetry service has three state settings: enabled, disabled, and undecided (default).

:::info
When the state setting is `undecided`, users have a delay period of one hour after startup before telemetry data is reported.
:::

### Disable telemetry

We urge you not to disable this service, as its insights will help us enhance our software.

While we are grateful for these usage insights, we understand that not everyone is comfortable with sending them.

:::caution
If you disable telemetry, automated version checks and updates will not be available.
:::

Telemetry can be disabled using any of the following options:

1. Use a command-line flag: To disable telemetry, run the command-line flag `--telemetry` and assign any of these values to it: `0`, `f`, `disable`, `no`, `false`, `n`, `disabled`, `optout`, `opt-out`, `disallow`, `forbid`.

   ```sh
   –telemetry=disable
   ```

2. Use an environment variable: To disable telemetry, assign any of these values to the `DO_NOT_TRACK` environment variable: `1`, `t`, `true`, or `enable`.

   ```sh
   DO_NOT_TRACK=true
   ```

3. Rename FerretDB executable to include a `donottrack` string.

   :::caution
   If telemetry is disabled using this option, you cannot run the `--telemetry` flag until the `donottrack` string is removed.
   :::

### Enable Telemetry

If telemetry is disabled, enable telemetry with the command-line flag `--telemetry` and assign any of these values to it: `1`, `t`, `enable`, `yes`, `true`, `y`, `enabled`, `optin`, `opt-in`, `allow`, `forbid`.

```sh
--telemetry=enable
```

If telemetry is disabled with a `donottrack` string in the executable, remove the `donottrack` string to use the command-line flag and values again.
