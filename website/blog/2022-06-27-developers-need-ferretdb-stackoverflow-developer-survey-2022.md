---
slug: developers-need-ferretdb-stackoverflow-developer-survey-2022
title: "Developers need FerretDB: StackOverflow Developer Survey 2022"
author: Peter Farkas
date: 2022-06-28
---

The annual [StackOverflow Developer Survey](https://survey.stackoverflow.co/2022/)is a great way to check the current adoption of different technologies out there.
This is a survey which is filled out by aspiring and professional engineers, voting on the technologies they use, or those they are most interested in.
Let's take a look at how FerretDB's plans align with the latest trends!

<!--truncate-->

When it comes to databases, this year, we see close to 70k responses, and the survey concludes that PostgreSQL is most loved database (and relational database) among Professional developers, and second most popular overall.

![Database survey](https://www.ferretdb.io/wp-content/uploads/2022/06/stackoverflow.jpg)

When it comes to document, or NoSQL databases, MongoDB is still the undisputed leader, and is heavily contesting the popularity of relational databases.

It is hard to think that any of the excitement comes from them adopting the [SSPL license](https://ssplisbad.com/) or that MongoDB’s managed offering, MongoDB Atlas, is probably mostly running on proprietary code these days.

The excitement undoubtedly comes from MongoDB getting developer experience right.
We cannot praise MongoDB enough for changing the database industry by focusing on developer experience itself.
Developer experience with some of the relational databases of the past decades were questionable, at best.
I hear the uproar of many DBAs as I am writing this, but let’s face it: most developers are not interested in operating a database, and definitely not interested in debugging, interfacing, replication, and the like.

However,  one needs to remember that MongoDB’s great developer experience comes with the high cost of vendor lock-in.
There is simply no real open source alternative to MongoDB, and absolutely no compatible alternative to MongoDB Atlas.

Back in 2016, shortly after launching MongoDB Atlas, MongoDB [published a blog post on the subject](https://www.mongodb.com/blog/post/avoiding-the-dark-side-of-the-cloud-platform-lock-in ""), and argues that MongoDB Atlas does not come with lock-in risks, since it runs on different cloud providers.
They also claimed back then that MongoDB Atlas is running the same software as MongoDB.

We don’t think that their claims hold true anymore.
You may be able to use different cloud providers, but you will still be in the mercy of MongoDB Inc, who licenses MongoDB to the cloud providers themselves.
Note that at the time of publishing the above referenced blog post, MongoDB was still licensed under Apache 2.0, which is no longer the case.
Nowadays, MongoDB Atlas does a lot more than on prem MongoDB, and the source code for those features are nowhere to be seen.

With Atlas, when you are in it, you will have to stay in it, and MongoDB is going to set the price on the privilege of storing and using your own data.
To us, this sounds scary, and we bet that if there would be an alternative offering more freedom, developers would be willing to consider it.

**So what is the solution to MongoDB's vendor lock-in?**

At FerretDB, we are working on combining the best of two worlds: the robustness and freedom of PostgreSQL, and the unmatched developer experience of MongoDB.

How we interpret the StackOverflow Developer Surveyfrom the perspective of FerretDB is that we are on the right track: we are working on a solution which is based on PostgreSQL, and we are providing an open source alternative to MongoDB.
Those service providers or companies currently running PostgreSQL will be able to cater to the needs of developers seeking for the ease of use of MongoDB

However, we are not going to stop at PostgreSQL: we are planning to support more databases, so developer experience can be standardized across many different platforms.

Working with partners, like [Tigris Data](https://tigrisdata.io), we can even provide an alternative to MongoDB Atlas itself.
On Monday, [we released FerretDB](https://github.com/FerretDB/FerretDB/releases/tag/v0.4.0)[0.0.4](https://github.com/FerretDB/FerretDB/releases/tag/v0.4.0), which comes with functionality supporting Tigris - a work in progress DBaaS, aspiring to become an alternative to MongoDB Atlas, and the people behind it are one of the finest experts in the world at providing database services at scale.

We are also looking to partner with the PostgreSQL community and PostgreSQL DBaaS providers, who are interested in working with us providing an alternative to MongoDB.
Feel free to let us know if you are interested!

![FerretDB team](https://www.ferretdb.io/wp-content/uploads/2022/06/group-1024x679.jpg)
