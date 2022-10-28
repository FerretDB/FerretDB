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
* Autonomous system number, cloud provider region,  or country from IP address
* Uptime performance (seconds) - the amount of time it takes to execute a particular FerretDB command

## Version notification

The telemetry service sends periodic reports containing information about the latest FerretDB version.
This information is recorded in the server logs, startupWarnings, and serverStatus outputs.
While you may not upgrade to the latest release immediately, ensure that you update early to take advantage of recent bug fixes, new features, and performance improvements.

## Disable telemetry

We urge you not to disable this service, as its insights will help us enhance our software.
While we are grateful for these usage insights, we understand that not everyone is comfortable with sending them.

To disable telemetry, run the following command in your terminal:

```sh
disableFreeMonitoring()
```

:::note
If you disable telemetry, automated version checks and updates will not be available.
:::

## Enable Telemetry

The telemetry service is enabled by default.
If telemetry is disabled, enable telemetry by running the following command in your terminal:

```sh
enableFreeMonitoring()
```
