---
sidebar_position: 1
---

# Evaluation

We provide evaluation images that come with FerretDB and PostgreSQL with DocumentDB extension.

- [`ghcr.io/ferretdb/ferretdb-eval:2`](https://ghcr.io/ferretdb/ferretdb-eval:2) image for quick testing and experiments.
- [`ghcr.io/ferretdb/ferretdb-eval-dev:2`](https://ghcr.io/ferretdb/ferretdb-eval-dev:2) image for debugging, with features that make it slower.

You'll need [Docker](https://docs.docker.com/get-docker/) installed to run it.

Run this command to start FerretDB with PostgreSQL + DocumentDB extension,
ensure to update `<username>` and `<password>`.

```sh
docker run -d --rm --name ferretdb -p 27017:27017 --platform linux/amd64 \
  -e POSTGRES_USER=<username> \
  -e POSTGRES_PASSWORD=<password> \
  -v ./data:/var/lib/postgresql/data \
  ghcr.io/ferretdb/ferretdb-eval:2
```

This command will start a container with FerretDB, pre-packaged PostgreSQL with DocumentDB extension, and MongoDB Shell for quick testing and experiments.

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
