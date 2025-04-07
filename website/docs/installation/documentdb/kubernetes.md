---
sidebar_position: 3
---

# Kubernetes

FerretDB uses PostgreSQL with [DocumentDB extension](https://github.com/microsoft/documentdb) as a database engine.
You need to have a running Kubernetes cluster to set up PostgreSQL with DocumentDB extension on Kubernetes.

You can deploy PostgreSQL with DocumentDB extension using any of our provided images.
Please see the [Docker installation docs](../documentdb/docker.md) to learn more about the available images.

:::tip
We strongly recommend specifying the full image tag (e.g., `17-0.102.0-ferretdb-2.1.0`)
to ensure consistency across deployments.
For more information on the best FerretDB image to use, see the [DocumentDB release notes](https://github.com/FerretDB/documentdb/releases/).
:::

The following `postgres.yaml` file creates a PostgreSQL instance with the DocumentDB extension:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: postgres
  labels:
    app: postgres
spec:
  containers:
    - name: postgres
      image: ghcr.io/ferretdb/postgres-documentdb:17-0.102.0-ferretdb-2.1.0
      ports:
        - containerPort: 5432
      env:
        - name: POSTGRES_USER
          value: <username>
        - name: POSTGRES_PASSWORD
          value: <password>
        - name: POSTGRES_DB
          value: postgres
      volumeMounts:
        - name: data
          mountPath: /var/lib/postgresql/data
  volumes:
    - name: data

---
apiVersion: v1
kind: Service
metadata:
  name: postgres
spec:
  selector:
    app: postgres
  ports:
    - port: 5432
      targetPort: 5432
```

Ensure to update the `<username>` and `<password>`.
Apply the `postgres.yaml` file to create the PostgreSQL instance:

```sh
kubectl apply -f postgres.yaml
```

This will create a service named `postgres` that FerretDB can use to connect to the Postgres instance.
Check the status of the pods to ensure that the PostgreSQL instance is running:

```sh
kubectl get pods
```

See [FerretDB Kubernetes installation](../ferretdb/kubernetes.md) for more details on connecting to FerretDB.
