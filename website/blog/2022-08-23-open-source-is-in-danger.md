---
slug: open-source-is-in-danger
title: "Open Source is in Danger as a Result Of Complacency, Plain Stupidity and Greed"
author: Peter Farkas
description: For those of us working in the IT industry, we should be well aware of the many benefits that open source software provides.
image: /img/blog/Maintainer@5x-1-300x300.png
date: 2022-08-23
---

For those of us working in the IT industry, we should be well aware of the many benefits that open source software provides.

<!--truncate-->

In the early years of our careers, open source software gave us the opportunity to experiment, learn and build things.
It gave us the excitement and privilege of understanding how things work under the hood.

Open Source gave us the confidence to build on the technology we love, because no one can really take it away from us.
And many of today's developers started out with a simple contribution to their favorite Open Source project.

As for myself, I can't imagine a world without Open Source software.

However, as with many things we benefit from in life, over time, we tend to take them for granted.
And open source itself is definitely not a shiny new toy—we have Linux, SQLite, PostgreSQL, and an endless amount of tools that make our lives and work easier.
We also have Open Source hardware.

In 2018, when we heard about[MongoDB's changeover to the non-open source Server Side Public License, or SSPL](https://www.mongodb.com/blog/post/mongodb-now-released-under-the-server-side-public-license), our first thought was "wow, how stupid of MongoDB".
They will lose contributors and users.
As it turned out, this knee-jerk reaction is not an accurate representation of the implications of MongoDB's move on Open Source.
For some, this move may have not resulted in significant differences.

Anyone, including MongoDB, Elastic or Graylog should be allowed to release their products under whatever license they would like to use, and the community should be free to continue if they want.
Life happens, bad or strange decisions like this happen —- we can move on.

However, in terms of what MongoDB did, the impact is just different.
No one should decide to come up with their own terms on what to call Open Source, and  this is exactly what is happening.
The SSPL license was supposed to "save" open source companies from Amazon or GCP, so they can't make money on them without giving back to the community.

Instead of "saving open source", however, what really happened, is that if you use MongoDB on a cloud provider, you will forever be at the mercy of MongoDB Inc. and the fees they will be charging you, through your infrastructure provider.
Your infrastructure provider will forever be on the hook to pay license fees to MongoDB.

And if you decide to provide a cloud service that uses MongoDB, you will need to buy a license of your own.
If you would like to move away - bummer, you don't have an alternative.

Does this sound like open source?

Telling users to go on-prem if they don't like the fees is cynical and not feasible, especially since most companies using open source will no longer have their own infrastructure to do so.
Moreover, I am curious about the amount of source code which is available for MongoDB Atlas, which you would theoretically need to run in order to replace it.

When MongoDB requested the OSI to recognize SSPL as an Open Source license, they were not taken aback by the fact that it was [not recognized as such](https://opensource.org/node/1099).
However, [MongoDB proceeded to continue calling theselves Open Source](https://www.mongodb.com/why-use-mongodb), just because they could.
They still do.
They decided that they understand better what the definition of open source is.

As [OpenUK](https://openuk.uk/)'s Amanda Brock[put it](https://www.computing.co.uk/analysis/4027028/elastic-stretched-patience-open-source) when Elastic adopted SSPL: "Let's be really clear - it's a move from open to proprietary as a consequence of a failed business model decision"

The compounding effect of this, in the long run, is the inevitable dilution of the true meaning, the true definition of Open Source.
New generations of users will have a very different understanding of what open source is.

The reason why FerretDB supports the Open Source Initiative as a [Maintainer level sponsor](https://opensource.org/corporate-sponsors-support) is because we believe the integrity of Open Source should be preserved for future generations, and for all of us benefiting from FOSS software on a daily basis.
This is one of the many ways we are planning to give back to the community.

![Open source initiative](/img/blog/Maintainer@5x-1-300x300.png)

FerretDB is working on a true [Open Source MongoDB-compatible database replacement](https://github.com/FerretDB/FerretDB), released under Apache 2.0, to make sure that the Open Source community, and anyone wanting to avoid a vendor lock-in situation, will have an alternative.
