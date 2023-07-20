---
slug: community-matters-fireside-chat-cockroachdb
title: 'Community Matters: Fireside Chat with Artem Ervits, CockroachDB'
authors:
  - name: Marcin Gwóźdź
    title: Director of Strategic Alliances at FerretDB
    url: https://www.linkedin.com/in/marcin-gwóźdź-277abaa9
    image_url: /img/blog/marcin-gwozdz.jpeg
  - name: Artem Ervits
    title: Solutions Engineer at CockroachDB
    url: https://www.linkedin.com/in/artemervits
    image_url: /img/blog/artem-ervits.jpeg
description: >
  We sat down with one of our earliest users, Artem Ervits from CockroachDB to discuss about open source, the database community, and what he thinks of FerretDB.
image: /img/blog/community-call-cockroachdb.jpg
tags: [open source, community, document databases, compatible applications]
---

![Community matters: fireside chat with Artem Ervits, CockroachDB](/img/blog/community-call-cockroachdb.jpg)

When we talk about open-source, we think about collaboration and freedom of choice: anyone can choose the tools that perfectly fit their needs.

<!--truncate-->

Open-source projects are even great additions to proprietary software and a lot of the proprietary solutions take advantage of the open source communities.

Community is what makes open-source software powerful, as well as the ease of access to its source code, so anybody can try it.
At [FerretDB](https://www.ferretdb.io/), we are constantly improving our software based on feedback from our active community.

One of our earliest users, Artem Ervits from [CockroachDB](https://www.cockroachlabs.com/), was playing with FerretDB during the early stages before the release of the first FerretDB GA, so we decided to ask him some questions:

**Why do you like open-source solutions?**

_I prefer to use open-source because of choice. There are a myriad of options available when you consider OSS vs. proprietary. Every time I get a new laptop, I dread clicking through the EULA. I don't know what I am signing and I don't have the patience to read through the small font. I stopped using Windows many years ago. LibreOffice is my go to when I need to edit a document. Who even buys a license anymore? When I joined Cockroach Labs I learned the founders Spencer Kimball and Peter Mattis are the original creators of Gimp, the photo editing software. Gimp, along with Wireshark and OpenOffice got me through college._

**What is important for you in an open-source project?**

_Friendly and welcoming community to new users and contributors and great documentation. I've participated in several open-source projects. Some projects were easy on the new users and others were unamicable. Some projects suffered from lack of ownership and new users fell victim to the internal bureaucracy. For others, lack of documentation hindered the adoption. It's a catch-22 and therefore, I always lean to communities with great content, community Slack and forums. It is empowering to receive a response from a project committer or project management committee. Those types of projects tend to stick around._

**How did you learn about FerretDB?**

_I follow the PostgreSQL ecosystem closely. When a new Mongo-compatible layer on top of PostgreSQL was introduced, I had to take a look for myself. The first thing I wanted to do was to run FerretDB with CockroachDB. I had doubts, but in the end, it worked out better than I anticipated._

**What is interesting for you in FerretDB? What do you like?**

_I really like the simplicity and stateless nature of FerretDB. It advertises itself as a Mongo-compatible proxy and it lives up to the promise as I'm able to get it up and running in a matter of minutes. I can spin up several instances of the proxy and instrument high availability, which serves itself well in combination with CockroachDB where high availability is non-negotiable. It offers an easy on-ramp for users of document databases to the relational world; to regain consistency and correctness._

**Do you see any pattern where document and relational databases are used together?**

_In my line of work, I speak to a lot of clients and their use cases vary. I come across clients using NoSQL databases as a caching layer atop an operational store. It's not something we do and we have to educate our users. The other place I see this scenario is where customers try to store large blobs of semi-structured data in a relational database. There are tradeoffs and we always advise to store these objects elsewhere._

_PostgreSQL has come a long way in supporting JSON and CockroachDB is right there with that support as well. But there are limitations and unless our customers adhere to the documented limitations, we have to walk away from those use cases in favor of something purpose-built for that._

**What use cases can you imagine for using CockroachDB and FerretDB together?**

_FerretDB is a great enabler for users coming from Mongo to PostgreSQL. Adopters of NoSQL more or less sacrificed cross-row transactions, secondary indexes, joins, constraints and SQL compatibility. From my first interaction with FerretDB, it was clear that I can take a Mongo archive and restore it to FerretDB backed by CockroachDB and have the data in Cockroach immediately available for querying via FerretDB, our JSON operators and directly via SQL and computed columns._

_MongoDB migrations are the first and primary use case where I see FerretDB bridge the path from document to relational world and I look forward to a time when FerretDB can be a drop-in replacement for a MongoDB Replica Set. Then, you don't even need to tolerate downtime by moving the data from Mongo to FerretDB in real time. But stepping back a bit, an opportunity to have a low lift migration to FerretDB from Mongo and also have the multiregion, consistency, correctness and resiliency story of CockroachDB has a significant edge compared to Mongo alone._

_Imagine you can leave your Mongo application as is and still benefit from availability zone and regional outage protection. Your uptime SLA story is much stronger and your application is no longer susceptible to planned and unplanned downtime. Then add to that ability to query the data with SQL and not just the Mongo API, the multiregion semantics, consistency guarantees. It's like having cake and eating it too!_

**Do you envision any specific CockroachDB-PostgreSQL differences that might pose compatibility challenges with FerretDB?**

_CockroachDB operates in SERIALIZABLE ISOLATION level and I imagine users coming from the Mongo world will see a larger than usual number of retry errors in CockroachDB. These errors do not surface as much in databases with weaker isolation; including PostgreSQL where default isolation is READ COMMITTED. CockroachDB does not have support for triggers, stored procedures, and no support for transactional schema changes. These gaps are the biggest hurdles for users coming from the PostgreSQL world, but fret not, our team is investing heavily in addressing these pain points._

**What are your expectations from FerretDB in the future?**

_I would like the FerretDB community to thrive and continue shipping features to close the gaps with Mongo compatibility. Eventually, I'd like to see deeper integration with CockroachDB to expose our native multiregion, spatial, and high availability capabilities. Both projects have unique capabilities that complement each other and unlocking this potential will be something I'd like to see come to fruition._

**How do you see the database community's future in general?**

_The database market is severely fragmented and looking within the PostgreSQL ecosystem, there are many one-off solutions that address partial pain points. Having the best-of-breed products to address the operational and analytical needs with a common API in a cloud-agnostic way is where I want the future to go. Usually, when I talk to customers, they say they are evaluating a handful of technologies, and on the surface, they all offer similar capabilities but going through the necessary cycles to validate is costly and painful. Too much choice is not always a good thing._

FerretDB has the ability to convert any Postgres-compatible database into a MongoDB-compatible document database.
At FerretDB, we fully understand the immense potential of collaborating with cloud-native, distributed database solutions like CockroachDB.

By combining CockroachDB and FerretDB, we can develop a seamlessly usable, high-speed, scalable, and resilient database solution that not only supports MongoDB workloads but also integrates with tools and frameworks from the MongoDB ecosystem.
Behind every software, there is always a community of users, so we wish to hear more voices from you!

Join our community!

- [FerretDB community Slack](https://docs.ferretdb.io/#community)
- [CockroachDB community Slack](https://www.cockroachlabs.com/join-community/)

### About The Speakers

- [Marcin Gwozdz - Director of Strategic Alliances at FerretDB](https://www.linkedin.com/in/marcin-gwóźdź-277abaa9)
- [Artem Ervits - Solutions Engineer at CockroachDB](https://www.linkedin.com/in/artemervits)

### About CockroachDB

CockroachDB is a distributed SQL database built on a transactional and strongly-consistent key-value store.
It scales horizontally; survives disk, machine, rack, and even data center failures with minimal latency disruption and no manual intervention; supports strongly-consistent ACID transactions; and provides a familiar SQL API for structuring, manipulating, and querying data.

### About FerretDB

FerretDB is a truly open-source alternative to MongoDB built on Postgres.
FerretDB allows you to use MongoDB drivers seamlessly with PostgreSQL as the database backend.
Use all tools, drivers, UIs, and the same query language and stay open-source.

Our mission is to enable the open-source community and developers to reap the benefits of an easy-to-use document database while avoiding vendor lock-in and faux pen licenses.
We are not affiliated, associated, authorized, endorsed by, or in any way officially connected with MongoDB Inc., or any of its subsidiaries or its affiliates.
