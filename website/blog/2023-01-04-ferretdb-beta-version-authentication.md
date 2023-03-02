---
title: FerretDB v0.8.0 - The Beta Version
slug: ferretdb-beta-version-authentication
author: Alexander Fashakin
description: The FerretDB beta version (v.0.8.0) includes exciting new features, including authentication for PostgreSQL, `$min` operator support, and much more.
image: /img/blog/FerretDB-is-now-Beta.-1-980x551.png
date: 2023-01-04
---

The FerretDB beta version (v.0.8.0) includes exciting new features, including authentication for PostgreSQL, `$min` operator support, and much more.

![FerretDB 0.8.0 release](/img/blog/FerretDB-is-now-Beta.-1-980x551.png)

<!--truncate-->

FerretDB - the open source alternative to MongoDB - is thrilled to announce release of our  Beta version (0.8.0), which includes various new features, bug fixes, improved documentation, and, most importantly, the implementation of authentication for PostgreSQL.
This new and exciting release is no ordinary milestone; it's a culmination of FerretDB's journey as we work on bringing you the ultimate open-source alternative to MongoDB by converting MongoDB protocol queries to SQL, with PostgreSQL as the database engine.

While we do not aim to cover all the features of MongoDB, our goal with the Beta is to provide a solid foundation on which to build targeted features that'll enable FerretDB to support more and more real-world use cases.
Be sure to [check our roadmap for further details](https://github.com/orgs/FerretDB/projects/2).

:::caution
Please note that this particular change breaks backward compatibility.
To dump and restore the data, you’ll need connection strings to connect to FerretDB and PostgreSQL.
There are numerous ways to dump and restore your data.
For example, you can follow the following steps:

1. Backup FerretDB databases with `mongodump`.
Set your FerretDB connection string in `-—uri` and run:
   `mongodump --uri="mongodb://127.0.0.1:27017"`
   This command will create a directory with dumps for each FerretDB database on the given server.
Later, we will use this dump to restore the database.
Instead of `mongodump`, you can also use `mongoexport` for this.

2. Backup PostgreSQL `ferretdb` database.
If the migration goes well, we will not use this backup.
But we recommend to do this step in case you need to rollback.
Specify your host and port.
`pg_dump -h 127.0.0.1 -p 5432 -U username ferretdb > ferretdb.sql`

3. Stop FerretDB (This depends on your operating system and the way you run FerretDB)

4. Connect to PostgreSQL and drop `ferretdb` database as it's not needed anymore:
   * `psql -h 127.0.0.1 -p 5432 -U username postgres`
   * `DROP DATABASE ferretdb`

5. Upgrade FerretDB and run FerretDB 0.8 (Please refer to [our documentation](https://docs.ferretdb.io/category/quickstart/) where we describe how to update and start FerretDB)

6. Restore database using `mongorestore --uri="mongodb://127.0.0.1:27017"`

7. While you can rollback with `mongorestore`, in case something doesn't work and you need to rollback to FerretDB 0.7.1:
   * Stop ferretdb
   * Delete ferretdb database (repeat step 4)
   * Restore `ferretdb` PostgreSQL db from the dump we created on the step 2:  `psql -h 127.0.0.1 -p 5432 -U username ferretdb -f ferretdb.sql`
   * Start ferretdb 0.7.1
:::

In this blog post, we'll be sharing detailed information about the FerretDB beta release (0.8.0).

## What's new?

Since the last release (0.7.2), we've made great strides in adding new features to FerretDB, especially the introduction of authentication for PostgreSQL.
With authentication, users can now connect securely using passwords, ensuring that only authorized connections are established.
You can do this by specifying your username and password in the FerretDB connection string as `mongodb://username:password@ferretdb:27018/?tls=true&authMechanism=PLAIN`.
See more details [in our documentation](https://docs.ferretdb.io/security/#authentication).

But that's not all - in addition to the `$max` update operator, FerretDB beta now includes support for the `$min` update operator:

```js
db.collection.update(
    {},
    {
        $min: {
            <field1>: <value1>,
            ...
        }
    }
)
```

We've also added support for `ordered` inserts, which allows you to insert data in the exact order it comes in.

```js
db.collection.insert({
        <field1>: <value1>,
        ...
    },
    {
        ordered: true
    }
)
```

## Bug fixes and enhancements

This release includes a bug fix that addresses an issue with the `$inc` operator when used with `unset` documents.
We found that, in certain cases, an invalid value of the `$inc` operator was updating `unset` documents with non-numeric values when it should have instead returned an error.
This release should prevent unexpected updates on `unset` documents, and likewise, improve the overall stability and reliability of FerretDB.

The release includes one enhancement, which is updating our building documentation.
We have documented our build process and Go build tags for those who use embedded packages or build FerretDB themselves.
Please [check here](https://github.com/FerretDB/FerretDB/blob/main/README.md#building-and-packaging) for further details.

## Documentation

Our documentation is not left out on the list of improvements!
Specifically, we've updated it to include descriptions and examples for element query operators, array query operators, and comparison and logical query operators.

These changes help make our documentation more comprehensive and user-friendly, which should hugely benefit our community and users alike.
Please don't hesitate to let us know what you think about it.

## Other changes

In this latest release, we've made some changes to our supported features, which include discontinuing support for `$elemMatch` and `$slice` projection operators due to existing technical issues.
However, this change is not a cause for concern, merely a signal that we will be focusing on other features and improvements that are of greater priority in [our roadmap](https://github.com/orgs/FerretDB/projects/2).
This change will not affect the support for the `$elemMatch` query operator.
For more information on the supported operators and commands, please check out our [updated documentation](https://docs.ferretdb.io/reference/supported_commands/).

Please find more details about the new features and changes in the beta version [here](https://github.com/FerretDB/FerretDB/releases/tag/v0.8.0).

## And to our amazing community and users

Once again, we want to take a moment to express our gratitude to everyone who has been part of the FerretDB journey so far - either as a user, contributor, supporter, or admirer.
Not to mention the growing number of partnerships, compatible applications, and all the contributions from our incredible community.
In 2022, we had:

* 👨🏻‍💻 Over 40 code contributors with more than 130 merged pull requests from our community of contributors
* ⭐️ 5.1k Stars on GitHub(Have you given us a Star yet? Be sure to [Star us on GitHub](https://github.com/FerretDB/FerretDB)
* ⏫ More than 100 Docker image downloads

We appreciate your continued and invaluable support and feedback, as we strive to make FerretDB even better.
Stay tuned for more exciting developments and updates from us!

Please remember, if you have questions about FerretDB, feel free to [contact us](https://docs.ferretdb.io/#community).
We'd love to hear from you!
