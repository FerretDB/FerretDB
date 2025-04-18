---
sidebar_position: 1
---

# Evaluation

We provide an evaluation image
[`ghcr.io/ferretdb/ferretdb-eval:2`](https://ghcr.io/ferretdb/ferretdb-eval:2)
for quick testing and experiments.

You'll need [Docker](https://docs.docker.com/get-docker/) installed to run it.

Run this command to start FerretDB with PostgreSQL + DocumentDB extension,
ensure to update `<username>` and `<password>`.

```sh
docker run -d --rm --name ferretdb -p 27017:27017 --platform linux/amd64 \
  --env POSTGRES_USER=<username> \
  --env POSTGRES_PASSWORD=<password> \
  ghcr.io/ferretdb/ferretdb-eval:2
```

This command will start a container with FerretDB, pre-packaged PostgreSQL with DocumentDB extension, and MongoDB Shell for quick testing and experiments.

However, it is unsuitable for production use cases because it keeps all data inside and loses it on shutdown.
See other [installation guides](..) for instructions that don't have those problems.

With that container running, you can:

- Connect to it with any MongoDB client application using MongoDB URI.

  ```text
  mongodb://<username>:<password>@127.0.0.1:27017/
  ```

- Connect to it using MongoDB Shell by just running the command below.

  ```sh
  mongosh mongodb://<username>:<password>@127.0.0.1:27017/
  ```

  If you don't have it installed locally, you can run the command below.

  ```sh
  docker exec -it ferretdb mongosh mongodb://<username>:<password>@127.0.0.1:27017/
  ```

- For PostgreSQL, connect to it by running the command below.

  ```sh
  docker exec -it ferretdb psql -U <username> postgres
  ```

In the [next step](usage/concepts.md), we will show you how to perform basic CRUD operations on FerretDB.

You can stop the container with `docker stop ferretdb`.
