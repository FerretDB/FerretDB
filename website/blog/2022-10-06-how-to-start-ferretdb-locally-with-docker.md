---
slug: how-to-start-ferretdb-locally-with-docker
title: "How to start FerretDB locally with Docker"
author: Patryk Kwiatek
image: ../static/img/blog/3g0okbkcve391.jpg
date: 2022-10-06
---

![how to start FerretDB locally](../static/img/blog/3g0okbkcve391.jpg)

<!--truncate-->

(image credit: [u/AssistantNava](https://www.reddit.com/user/AssistantNava/) / [Reddit](https://www.reddit.com/r/ferrets/comments/v3zv0z/ferret_on_da_computer/))

Even though FerretDB is still in development, it's possible to check out its capabilities and run it in your local environment!

The process of installing and running FerretDB is not complicated at all, and you can do it in just a few steps.  

## Prerequisites

To set up FerretDB locally using Docker, you will need to have the following installed:

* [Docker](https://docs.docker.com/get-docker/)
* [Docker Compose](https://docs.docker.com/compose/install/)
* Text editor of choice

## How to set up FerretDB environment locally

Once the prerequisites are installed, the next thing to do is to create a `docker-compose.yml` file, which will contain declarative configuration of all required Docker containers.

Note that, if you’ve already deployed a PostgreSQL compatible database in your environment, you can skip this step and use its address in later steps.

To begin with, we need to deploy our database of choice.
For our example, we will be using PostgreSQL 14, which is the stable version of the database.

Ensure to specify the `POSTGRES_DB` env variable with the same database name for FerretDB.
It will be used to store all MongoDB collections.

```js
version: "3"

services:
services:
  postgres:
    image: postgres:14
    container_name: postgres
    ports:
      - 5432:5432
    environment:
      - POSTGRES_USER=user
      - POSTGRES_DB=ferretdb
      - POSTGRES_HOST_AUTH_METHOD=trust

```

Please remember that the above deployment is not suitable for production environments.
It doesn't use any volumes to store data and the `POSTGRES_HOST_AUTH_METHOD=trust`, which means that there is no password required for the database.

We can now run `docker-compose up -d` to start our environment from the docker-compose file.
The `-d` flag will run containers in the background.
And If something isn't working as expected, it's worth it to remove the flag to check the logs.

Let’s check the current status of our docker environment.
Run the `docker-compose ps` command in the directory with `docker-compose.yml` file to easily check all running containers.

```js
> docker-compose ps
  Name                Command              State                    Ports
-------------------------------------------------------------------------------------------
postgres   docker-entrypoint.sh postgres   Up      0.0.0.0:5432->5432/tcp,:::5432->5432/tcp

```

Great!
The PostgreSQL container is now up and running.
It is available under the 5432 port on our local machine.

The next step is to serve the FerretDB container.
It will connect to the PostgreSQL database and act as a proxy.

```js
ferretdb:
  image: ghcr.io/ferretdb/ferretdb:latest
  container_name: ferretdb
  restart: on-failure
  ports:
    - 27017:27017
  command: ["--listen-addr=:27017", "--postgresql-url=postgres://user@postgres:5432/ferretdb"]

```

In the command section, you can specify custom values that will match your requirements.

The `--listen-addr` will set the port on which FerretDB will listen for requests.
The `--postgresql-url` is an address for FerretDB to connect to DB.
The user, hostname, port and database name after the slash should match the values provided in the PostgreSQL deployment.

It’s good practice to specify the PostgreSQL hostname by using the PostgreSQL container name.

You can also specify other flags, but at the moment most of them are created for development and testing:

```js
Flags:
  -h, --help                                   Show context-sensitive help.
      --version                                Print version to stdout (full version, commit, branch, dirty flag) and exit.
      --listen-addr="127.0.0.1:27017"          Listen address.
      --proxy-addr="127.0.0.1:37017"           Proxy address.
      --debug-addr="127.0.0.1:8088"            Debug address.
      --mode="normal"                          Operation mode: [normal proxy diff-normal diff-proxy].
      --test-record=""                         Directory of record files with binary data coming from connected clients.
      --handler="pg"                           Backend handler: dummy, pg, tigris.
      --postgresql-url="postgres://postgres@127.0.0.1:5432/ferretdb"
                                               PostgreSQL URL.
      --log-level="debug"                      Log level: debug, info, warn, error.
      --test-conn-timeout=0                    Test: set connection timeout.
      --tigris-client-id=""                    Tigris Client ID.
      --tigris-client-secret=""                Tigris Client secret.
      --tigris-token=""                        Tigris token.
      --tigris-url="http://127.0.0.1:8081/"    Tigris URL.

```

Now we can rerun `docker-compose up -d` to apply new changes.
Let’s see if the FerretDB container has started properly.

```js
> docker-compose ps
  Name                Command               State                      Ports
------------------------------------------------------------------------------------------------
ferretdb   /ferretdb --listen-addr=:2 ...   Up      0.0.0.0:27017->27017/tcp,:::27017->27017/tcp
postgres   docker-entrypoint.sh postgres    Up      0.0.0.0:5432->5432/tcp,:::5432->5432/tcp

```

The container is up and running, it also forwards local :27017 port to the same port on the container.
To be sure, let’s check FerretDB logs by using `docker logs` command with the container name or ID as an argument.

```js
> docker-compose logs ferretdb
Attaching to ferretdb
ferretdb    | 2022-10-05T23:45:35.989Z  INFO    ferretdb/main.go:111    Starting FerretDB v0.5.4...     {"version": "v0.5.4", "commit": "d2bdcb45ea319c657f44cc0a18783c145cb871c7", "branch": "main", "dirty": false, "-compiler": "gc", "-race": "true", "-tags": "ferretdb_testcover,ferretdb_tigris", "-trimpath": "true", "CGO_ENABLED": "1", "GOARCH": "amd64", "GOOS": "linux", "GOAMD64": "v1"}
ferretdb    | 2022-10-05T23:45:35.993Z  INFO    listener        clientconn/listener.go:78       Listening on [::]:27017 ...
ferretdb    | 2022-10-05T23:45:35.995Z  INFO    debug   debug/debug.go:60       Starting debug server on http://127.0.0.1:8088/

```

As we can see, FerretDB is waiting for incoming connections.

Let’s add a network to make a container communication separated from your environment.

```js
networks:
  default:
    name: ferretdb
```

Now we can try to connect to the FerretDB container using mongosh.

If you have mongosh on your machine you can just use it.
If not, you can create a simple MongoDB container and run the mongosh from there:

```js
> docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo:5 mongodb://ferretdb/

For mongosh info see: https://docs.mongodb.com/mongodb-shell/


To help improve our products, anonymous usage data is collected and sent to MongoDB periodically (https://www.mongodb.com/legal/privac
y-policy).
You can opt-out by running the disableTelemetry() command.

------
   The server generated these startup warnings when booting
   2022-10-05T23:58:59.045Z: Powered by 🥭 FerretDB v0.5.4 and PostgreSQL 14.5.
   2022-10-05T23:58:59.045Z: Please star us on GitHub: https://github.com/FerretDB/FerretDB
------

test>

```

### Testing commands in FerretDB

We’ve managed to create a sustainable environment for MongoDB drivers to run in!
Let’s see what will happen if we try to insert something to the collection.

```js
test> db.ferrets.insertOne({name: "Zippy", age: 4})
{
  acknowledged: true,
  insertedId: ObjectId("633e211b9e6442c3da4e1d21")
}
test>

```

With the insertOne() comand, one document record is inserted into the collection.
Let’s attempt to retrieve the documents in this collection using the find() method:

```js
test> db.ferrets.find({})
[
  { _id: ObjectId("633e211b9e6442c3da4e1d21"), name: 'Zippy', age: 4 }
]
test>

```

We are able to run MongoDB queries and store them successfully in the database!

### Checking FerretDB data in PostgreSQL

For our curiosity, let's see if the data is stored in the PostgreSQL database and in what way.
To do that, let’s execute `psql` command on the PostgreSQL container:

```js
> docker exec -ti postgres psql --user=user --db=ferretdb
psql (14.5 (Debian 14.5-1.pgdg110+1))
Type "help" for help.

ferretdb=#
```

Now, we are able to control and run queries on the postgres database.
Let’s check tables in `test` schema which was used by mongosh.

```js
List of relations
 Schema |        Name        | Type  | Owner
--------+--------------------+-------+-------
 test   | _ferretdb_settings | table | user
 test   | ferrets_b90ada46   | table | user
 test   | test_afd071e5      | table | user
(3 rows)

ferretdb=#

```

As we can see, there is a table `ferrets_b90ada46` created for our collection.
Now let’s print its content:

```js
ferretdb=# SELECT * FROM test.ferrets_b90ada46;
                                                _jsonb
------------------------------------------------------------------------------------------------------
 {"$k": ["name", "age", "_id"], "_id": {"$o": "633e211b9e6442c3da4e1d21"}, "age": 4, "name": "Zippy"}
(1 row)

```

From the output, we can conclude that the single MongoDB document is stored in a single row as a jsonb value.

## Recap

There you have it!
In just a few steps, we’ve been able to set up and run FerretDB locally using Docker.
We’ve also tested it out by inserting and retrieving a document in a collection, and exploring how it’s stored in PostgreSQL.

Even though FerretDB is still under development and not suitable for production environments, running it locally using Docker gives you a taste of what's to come.
But there’s still more to do.
You can start contributing to FerretDB by taking a look at [this article](https://blog.ferretdb.io/how-to-contribute-to-open-source-2022/).
To learn more about FerretDB, contribute to the project, or for any questions you might have, do reach out to us on [Slack](https://github.com/FerretDB/FerretDB#community) or [GitHub](https://github.com/FerretDB/FerretDB/discussions).
