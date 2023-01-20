---
slug: 0-5-0-release-is-out-embedding-ferretdb-into-go-programs
title: "New release - embedding FerretDB 0.5.0 into Go programs"
author: Alexey Palazhchenko
description: FerretDB v0.5.0 includes a new exciting feature – the ability to use it as a regular Go library package.
image: /img/blog/group-of-ferrets-on-white.jpg
date: 2022-07-11
---

[FerretDB v0.5.0, released today](https://github.com/FerretDB/FerretDB/releases/tag/v0.5.0), includes a new exciting feature – the ability to use it as a regular Go library package.

![Image credit: allthingsnature.org](/img/blog/group-of-ferrets-on-white.jpg)

<!--truncate-->

It can be embedded into a program and deployed as a single artifact without a need to run FerretDB as a separate process.
Then any MongoDB client application could connect to it and use it normally, while data will be stored in PostgreSQL.
Even the program that embeds FerretDB could connect to it if there is a need to do that.

Let's see how the [ferretdb package](https://pkg.go.dev/github.com/FerretDB/FerretDB/ferretdb) could be used.
First, we need to add a Go module to dependencies as usual:

```js
go get github.com/FerretDB/FerretDB/ferretdb@latest
```

Then we create a new instance of embedded FerretDB that would use the specified PostgreSQL database for storage:

```js
f, _ := ferretdb.New(&ferretdb.Config{
    Handler:       "pg",
    PostgreSQLURL: "postgres://username:password@127.0.0.1:5432/database",
})

```

We make it run in the background:

```js
go f.Run(context.Background())
```

And then, we use a method to get a MongoDB URI that can be used with any client:

```js
fmt.Println(f.MongoDBURI())

```

For example, we can connect to it with Mongo Shell:

```js
$ mongosh mongodb://127.0.0.1:27017/

    Current Mongosh Log ID: 62cb2d6a37f455b3cd5f0004
    Connecting to:  mongodb://127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+1.5.0
    Using MongoDB:  5.0.42
    Using Mongosh:  1.5.0

    For mongosh info see: https://docs.mongodb.com/mongodb-shell/

    ------
    The server generated these startup warnings when booting
    2022-07-10T19:50:02.183Z: Powered by 🥭 FerretDB unknown and PostgreSQL 14.4.
    2022-07-10T19:50:02.183Z: Please star us on GitHub: https://github.com/FerretDB/FerretDB
    ------

test>
```

And that's it!
With just a few lines of code, you can avoid the need to run a separate FerretDB or MongoDB process, all while using an open-source library.
Of course, that's only the first step for that functionality.
Some configuration options are missing, and some additional APIs might be needed.
Please [join our community](https://github.com/FerretDB/FerretDB#community) and tell us what you think, what works great and what doesn't, and what additional functionality is needed.
We will be happy to hear from you!

*A slightly bigger example can be seen in this repo: [https://github.com/FerretDB/embedded-example](https://github.com/FerretDB/embedded-example)*
