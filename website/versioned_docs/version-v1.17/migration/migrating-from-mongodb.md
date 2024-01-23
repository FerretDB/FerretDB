---
sidebar_position: 2
---

# Migrating from MongoDB

Before you begin this section of the migration process, go through the pre-migration process so as to ensure a successful migration from MongoDB to FerretDB:

- [Pre-migration testing](premigration-testing.md)

This guide will help you migrate your data from MongoDB – locally or online – to FerretDB.

As an open-source MongoDB alternative, FerretDB is built to work with many MongoDB tools.
In that case, you can migrate your data using MongoDB native tools such as `mongodump`/`mongorestore` and `mongoexport`/`mongoimport`.

Before you go forward with the migration, you need to have the following:

- MongoDB connection URI
- FerretDB connection URI
- MongoDB native tools

## Backup your MongoDB data

To backup your MongoDB instance using `mongodump` or `mongoexport`, you'll need to set the connection string to your MongoDB instance (e.g. `"mongodb://127.0.0.1:27017"`) to run the following command:

```sh
mongodump --uri="mongodb://<yourusername>:<yourpassword>@<host>:<port>"
```

The `mongodump` command will create a dump of all the data in the instance, consisting of BSON files of all the collections.
Also, you can migrate a specific database (e.g. `--db=test`) or collection (e.g. `--collection=supply`) using their respective parameters after the `--uri` connection string.

:::caution
If you include the database in your connection string, there's no need to specify a database name for the backup or restore process.
:::

```sh
mongoexport --uri="mongodb://<yourusername>:<yourpassword>@<host>:<port>" --db=<database-name> --collection=<collection-name> --out=<collection>.json
```

On the other hand, `mongoexport` does not provide a direct way to export all the collections at once, like `mongodump` does.

Instead, you need to set the connection string to connect with your preferred database and then run the command together with the parameters for the collection (`--collection=myCollection`) and the directory you want to export to (e.g. `--out=collection-name.json`).

## Restore your data to FerretDB

To restore or import your backed-up data to FerretDB, set the connection string to your FerretDB instance and use `mongorestore` and `mongoimport`.

Run the following command in your terminal, from your `dump` folder:

```sh
mongorestore --uri="mongodb://<yourusername>:<yourpassword>@<host>:<port>/?authMechanism=PLAIN"
```

With this command, you can restore all the data in `dump` into your FerretDB instance.
You can also specify the database and collection (`dump/<database>/<collection>`) you want to restore from the `dump` folder, according to your preferences.

To import your database using `mongoimport`, run the command from the terminal directory where you exported your data:

````sh

```sh
mongoimport--uri="mongodb://<yourusername>:<yourpassword>@<host>:<port>/?authMechanism=PLAIN" --db=<database-name> --collection=<collection-name> --file=<collection>.json
````

The command will import the specified collection you exported from your MongoDB instance to FerretDB.
