---
slug: new-ferretdb-release-v116
title: FerretDB releases v1.16.0!
authors: [alex]
description: >
  We are delighted to announce the release of FerretDB v1.16.0, where we update our documentation to be compatible with Docusaurus v3, enable support for `DeleteAll` for capped collections, and more.
image: /img/blog/ferretdb-v1.16.0.jpg
tags: [release]
---

![FerretDB releases v.1.16.0](/img/blog/ferretdb-v1.16.0.jpg)

We are delighted to announce the release of FerretDB v1.16.0, where we update our documentation to be compatible with Docusaurus v3, enable support for `DeleteAll` for capped collections, and more.

<!--truncate-->

FerretDB is an open-source document database that provides MongoDB compatibility to other database backends, including [Postgres](https://www.postgresql.org/) and [SQLite](https://www.sqlite.org/), and more are on the way.
Let's check out some of the latest changes in this release, and also give a round-up of some of our achievements this year.

## Latest changes

In this release, we have focused on improving FerretDB and setting up more integration tests, which should make it easier to support other backends in the future.
Similarly, we're also enhancing pushdowns by removing unsafe pushdowns to ensure full compatibility with MongoDB, and this will also help to enable easy support for new backends.

As part of our ongoing work to support capped collections, we've enabled' DeleteAll' support.
We've also added TLS support to proxy mode; you can observe the communication between TSL- and authentication-enabled client and server.

We made a few changes in our documentation and blog pages to be compatible with [Docusaurus v3](https://docusaurus.io/blog/preparing-your-site-for-docusaurus-v3), and now we've succesfully upgraded to the new version.

Check out the [release notes for other changes in this release.](https://github.com/FerretDB/FerretDB/releases/tag/v1.16.0)

## Highlights of the year

It's the season of good cheer, with festivities and a new year just around the corner.
As every company rounds up on the year, we're happy to highlight some of our incredible milestones this year.

- In April 2022, we released [the first production-ready version of FerretDB](https://blog.ferretdb.io/ferretdb-1-0-ga-opensource-mongodb-alternative/), based on Postgres and since then, we've seen many companies go on to adopt FerretDB in their applications.

- Besides adding a production-ready support for the SQLite backend, we've also redesigned our backend architecture to ensure better performance and, most importantly, make it easy to support more database backends - see [here for more details](https://blog.ferretdb.io/ferretdb-v1-10-production-ready-sqlite/#the-new-architecture).

- For folks interested in running a managed FerretDB service, this is now possible with [Scaleway](https://www.scaleway.com/en/betas/#managed-document-database), [Civo](https://www.civo.com/marketplace/FerretDB), and [Vultr](https://www.vultr.com/docs/ferretdb-managed-database-guide/).

- Community is at the center of everything we do, and in the past few months we've been a part of some truly amazing conferences and open source gatherings: Civo Navigate 2023 in London, Postgres Ibiza 2023, Percona University, ATO 2023, are just a few of them.

- We've also seen a sharp rise in the number of contributors to FerretDB; we were especially proud to welcome so many new contributors during the last Hacktoberfest celebrations.

Of course, there's still so many highlights in 2023 that we can't fit in this post, but we plan to cover them in another blog post - so stay tuned!

None of this would be possible without the unwavering support of the open source community and the FerretDB users.
Thank for believing in open source and supporting a true open-source MongoDB alternative suitable for many use cases.

If you're just discovering FerretDB, we are always happy to have new members join our community – check [us out on GitHub](https://github.com/FerretDB/FerretDB).
