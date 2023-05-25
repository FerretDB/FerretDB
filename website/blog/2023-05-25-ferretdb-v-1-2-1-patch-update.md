---
slug: ferretdb-v-1-2-1-patch-update
title: "FerretDB v1.2.1. Patch Update"
authors: [alex]
description: >
  FerretDB v1.2.1. is an important patch update that addresses critical issues in the latest release.
image: /img/blog/ferretdb-1-2-1-patch-update.jpg
tags: [release, patch]
---

![FerretDB v.1.2.1](/img/blog/ferretdb-1-2-1-patch-update.jpg)

We've just released a patch update that addresses critical issues in the last FerretDB release.

<!--truncate-->

This patch targets two important problems experienced by our users, concerning both update notifications and error handling for authentication issues.

Previously, when an update was no longer available, [FerretDB](https://www.ferretdb.io/) would still report an update being available.
In this patch update, we've fixed this issue and now FerretDB should accurately reflect the status of available updates correctly via [telemetry](https://docs.ferretdb.io/telemetry/).

In the previous, we noticed that error messages related to authentication were quite unclear.
This patch addresses this by returning more insightful error messages, and also provide a link to [our documentation](https://docs.ferretdb.io/) for more clarity.

Please ensure to apply this patch as soon as possible to take advantage of these crucial fixes.
Thank you.

If you have any questions or need further assistance, please [contact us here](https://docs.ferretdb.io/#community).
