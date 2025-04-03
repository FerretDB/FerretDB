---
slug: run-ferretdb-postgres-documentdb-extension-cnpg-kubernetes
title: 'How to deploy FerretDB with CloudNativePG on Kubernetes'
authors: [alex]
description: >
  Learn how to deploy FerretDB with PostgreSQL using CloudNativePG on Kubernetes.
image: /img/blog/ferretdb-cnpg.jpg
tags: [compatible applications, tutorial, cloud, postgresql tools, open source]
---

![Deploy FerretDB with CloudNativePG on Kubernetes](/img/blog/ferretdb-cnpg.jpg)

Running FerretDB on Kubernetes typically requires managing a PostgreSQL backend.
With [CloudNativePG](https://cloudnative-pg.io/), a Kubernetes-native operator for PostgreSQL, you can automate provisioning, scaling, and failover while keeping your infrastructure declarative and production-ready.

<!--truncate-->

We previously covered how to run FerretDB on Kubernetes using CloudNativePG (CNPG) as the PostgreSQL operator.
The setup process has changed since FerretDB v1.x.
With v2.x, FerretDB introduces new features, improved performance, and better MongoDB compatibility – which now rely on the PostgreSQL with the DocumentDB extension.

This guide will walk you through the steps to deploy FerretDB and PostgreSQL with DocumentDB extension using CloudNativePG on Kubernetes.

## Install CNPG with Helm

Install the CNPG operator into a separate namespace.
This operator will watch for `Cluster` resources and manage PostgreSQL accordingly.

```sh
helm repo add cnpg https://cloudnative-pg.github.io/charts
helm upgrade --install cnpg \
  --namespace cnpg \
  --create-namespace \
  cnpg/cloudnative-pg
```

## Create the PostgreSQL cluster

We'll now define a PostgreSQL cluster using the `Cluster` resource from CNPG.
This cluster will run 3 Postgres instances using FerretDB's special `postgres-documentdb` image that includes the required extensions.
CNPG will manage replication, failover, and lifecycle.

We explicitly enable `enableSuperuserAccess` so that we can connect with the default `postgres` user.
We also set the `postgresUID` and `postgresGID` to 999, which is the UID and GID used in the FerretDB image.

We also load the required extensions and enable shared preload libraries needed by FerretDB.

Save this as `pg-cluster.yaml`:

```yaml
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: postgres-cluster
  namespace: cnpg
spec:
  instances: 3
  imageName: 'ghcr.io/ferretdb/postgres-documentdb:17-0.102.0-ferretdb-2.0.0'
  postgresUID: 999
  postgresGID: 999
  enableSuperuserAccess: true

  storage:
    size: 1Gi

  postgresql:
    shared_preload_libraries:
      - pg_cron
      - pg_documentdb_core
      - pg_documentdb
      - pg_stat_statements
    parameters:
      cron.database_name: 'postgres'

  bootstrap:
    initdb:
      postInitSQL:
        - 'CREATE EXTENSION IF NOT EXISTS documentdb CASCADE;'
```

Apply it:

```sh
kubectl apply -f pg-cluster.yaml -n cnpg
```

CNPG will handle cluster creation, persistent storage, and generate a password for the `postgres` superuser in a secret.

You can check the status of the cluster with:

```sh
kubectl get cluster -n cnpg
```

You should see the cluster in `Running` state.

```text
NAME                STATUS   REPLICAS   READY   AGE
postgres-cluster   Running  3          3       1m
```

To get the generated password for the `postgres` user:

```sh
kubectl get secret -n cnpg postgres-cluster-superuser -o jsonpath='{.data.password}' | base64 -d && echo
```

## Deploy FerretDB

Now that the PostgreSQL backend is ready, we deploy FerretDB itself.
It expects a connection string to PostgreSQL via the `FERRETDB_POSTGRESQL_URL` environment variable.

Save this as `ferretdb.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ferretdb
  namespace: cnpg
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ferretdb
  template:
    metadata:
      labels:
        app: ferretdb
    spec:
      containers:
        - name: ferretdb
          image: ghcr.io/ferretdb/ferretdb:2.0.0
          ports:
            - containerPort: 27017
          env:
            - name: FERRETDB_POSTGRESQL_URL
              value: 'postgresql://postgres:<paste-password-here>@postgres-cluster-rw.cnpg.svc.cluster.local:5432/postgres'

---
apiVersion: v1
kind: Service
metadata:
  name: ferretdb-service
  namespace: cnpg
spec:
  selector:
    app: ferretdb
  ports:
    - protocol: TCP
      port: 27017
      targetPort: 27017
  type: NodePort
```

Then apply FerretDB:

```sh
kubectl apply -f ferretdb.yaml -n cnpg
```

This will create a service named `ferretdb-service` that FerretDB can use to connect to the Postgres instance.
Check the status of the FerretDB pod to ensure that it is running:

```sh
kubectl get pods -n cnpg
```

## Connect to FerretDB

Expose FerretDB locally by port-forwarding its service:

```sh
kubectl port-forward svc/ferretdb-service -n cnpg 27017:27017
```

Then in another terminal:

```sh
mongosh "mongodb://postgres:<password>@localhost/postgres"
```

You're now connected to FerretDB.

### Test CRUD operations

Try a quick test:

```js
db.test.insertOne({ hello: 'world' })
db.test.find()
```

This verifies that FerretDB is up and the entire setup is functioning as expected.

## Conclusion

You now have FerretDB v2 running on Kubernetes, backed by a production-grade PostgreSQL cluster managed by CloudNativePG.
This setup gives you MongoDB-compatible functionality with the reliability and tooling of PostgreSQL.
For production environments, consider adding backups, monitoring, authentication, and horizontal scaling.

Have any questions?
Reach out to us on [our community channels](https://docs.ferretdb.io/#community).
