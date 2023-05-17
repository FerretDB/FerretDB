---
slug: introducing-ferretdb-beacon-telemetry-service
title: "Introducing FerretDB Beacon: Help Us Increase Compatibility"
author: Peter Farkas
description: Introducing FerretDB Beacon – a service to help us effectively identify compatibility issues and prioritize solutions based on user interaction with FerretDB.
image: /img/blog/allen-cai-Y4RxCIaYaSk-unsplash-1024x683.jpg
date: 2022-10-03
---

Introducing FerretDB Beacon – a service to help us effectively identify compatibility issues and prioritize solutions based on user interaction with FerretDB.

![FerretDB Beacon](/img/blog/allen-cai-Y4RxCIaYaSk-unsplash-1024x683.jpg)

<!--truncate-->

(image credit: Allen Cai @aycai / [Unsplash](https://unsplash.com/photos/Y4RxCIaYaSk))

At [FerretDB](https://www.ferretdb.io/), we are on a journey to build the de-facto open source alternative to MongoDB, and this means we must ensure compatibility and support with all the necessary MongoDB drivers and protocols, as well as all your favorite commands.
But this may be hard to do without knowing what’s missing or which particular issues our users are experiencing.

That is the problem we are currently facing – how do we make FerretDB a better and more complete open source replacement for MongoDB?

The feedback and comments from our Slack and GitHub community have been invaluable to us in many ways.
However, we think there’s so much more to learn!

FerretDB Beacon is a telemetry service that can help us prioritize issues and improvements effectively by understanding how users interact with FerretDB.
We would like to gather usage data on how our features are being used, failing MongoDB commands, and above all compatibility issues.

For instance, CLA assistant may appear to be fully supported by FerretDB.
However, at least one of the rarely used features would require aggregation pipelines to be supported, which is not currently the case.
These scenarios can only be discovered by collecting information on unsupported commands.

With FerretDB Beacon, the most common unsupported commands can be prioritized based on real world needs.
This is the reason why we are currently working on FerretDB Beacon, which is planned to be introduced with our upcoming Alpha release at the beginning of October.
In addition, FerretDB Beacon users will be notified when new versions are available.

However, we understand that data collection can be a sensitive subject for many in the open source community.
Because of this, we're implementing this functionality with complete transparency in mind, and anyone will be free to check the short source code responsible for data collection.

You will have the chance to opt-out whenever you want (even before starting FerretDB), but even so, we urge you to support us in this endeavor – the correct calibration and logging of bugs and usage data could substantially speed up our processes towards improving FerretDB.
And if you decide to opt out, we hope that you can join our community and provide feedback manually via issues, features requests, and discussions.

Most importantly, we remain committed to the principles of open source and community, recognizing that we cannot do this without your help and support.
In light of this, we’d like to get your opinion on some of the concerns you might have regarding the use of FerretDB Beacon and how we can ease those fears.

If you have any questions or concerns you would like to raise, kindly reach out to us on [Slack chat](https://join.slack.com/t/ferretdb/shared_invite/zt-zqe9hj8g-ZcMG3~5Cs5u9uuOPnZB8~A) and [GitHub Discussions](https://github.com/FerretDB/FerretDB/discussions) for more information.
You can also join our [open office hours meeting](https://calendar.google.com/event?action=TEMPLATE&amp;tmeid=NjNkdTkyN3VoNW5zdHRiaHZybXFtb2l1OWtfMjAyMTEyMTNUMTgwMDAwWiBjX24zN3RxdW9yZWlsOWIwMm0wNzQwMDA3MjQ0QGc&amp;tmsrc=c_n37tquoreil9b02m0740007244%40group.calendar.google.com&amp;scp=ALL) every Monday at 18:00 UTC at [Google Meet](https://meet.google.com/mcb-arhw-qbq).
