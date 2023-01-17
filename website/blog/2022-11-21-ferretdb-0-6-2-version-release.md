---
slug: ferretdb-0-6-2-version-release
title: "New FerretDB release – 0.6.2: Now runs on Raspberry Pi!"
author: Alexander Fashakin
date: 2022-11-21
---

![FerretDB 0.6.2 release](https://www.ferretdb.io/wp-content/uploads/2022/11/ferret_rpi-1-1024x390.png)

<!--truncate-->

We are happy to announce FerretDB’s latest version release 0.6.2.
Even though this is not a major version release, it does come with a few exciting features, bug fixes, and enhancements, as well as improved documentation.

<!--truncate-->

## New features

For Raspberry Pi users, FerretDB now provides builds for `linux/arm/v7`.
This new feature further expands the range of environments that FerretDB supports.
We have also added a way for you to enable or disable telemetry at runtime.
Besides that, we’ve implemented a new feature for setting and getting telemetry status at runtime, enabling you to use the same as MongoDB for free monitoring.
Please [check our documentation for more information](https://docs.ferretdb.io/telemetry/#enable-telemetry "").

## Documentation

[Our documentation](https://docs.ferretdb.io "") has also been updated.
In the latest release, we have published our commands parity guide with MongoDB, where you can see the current list of features that we support.

## Bug Fixes

We’ve fixed issues with Unix socket listeners, where you get internal errors or panic when running FerretDB with a Unix socket listener.

## Other changes and enhancements

In other changes, we've made it easier to configure FerretDB in container and cloud environments through environment variables.
For identified errors, we’ve improved the accuracy of telemetry data in some cases where arguments or operators are present but not implemented, or return errors.
Furthermore, we have enabled the use of  `-` in collection names in line with real-life app usages.

Please find more details on the latest FerretDB version release [here on GitHub](https://github.com/FerretDB/FerretDB/releases "").
And if you have any questions, feel free to [contact us](https://docs.ferretdb.io/#community "").

Raspberry Pi is a trademark of Raspberry Pi Ltd.
