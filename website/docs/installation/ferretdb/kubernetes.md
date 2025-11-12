---
sidebar_position: 2
---

# Kubernetes

To deploy FerretDB on Kubernetes, you need to have a running Kubernetes cluster and PostgreSQL instance with DocumentDB extension.
Please see the [DocumentDB installation docs](../documentdb/kubernetes.md) for more information on how to deploy it on Kubernetes.

We provide different FerretDB images for various deployments.
Please see the [Docker installation docs](docker.md) to learn more on the available images.

:::tip
We strongly recommend specifying the full image tag (e.g., `2.8.0`)
to ensure consistency across deployments.
Ensure to [enable telemetry](../../telemetry.md) to receive notifications on the latest versions.

For more information on the best DocumentDB version to use, see the [corresponding release notes for the FerretDB version](https://github.com/FerretDB/FerretDB/releases/).
:::

Create a `ferretdb.yaml` file with the following content:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ferretdb
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
          image: ghcr.io/ferretdb/ferretdb:2.8.0
          ports:
            - containerPort: 27017
          env:
            - name: FERRETDB_POSTGRESQL_URL
              value: postgres://<username>:<password>@postgres:5432/postgres

---
apiVersion: v1
kind: Service
metadata:
  name: ferretdb
  labels:
    app: ferretdb
spec:
  selector:
    app: ferretdb
  ports:
    - port: 27017
      targetPort: 27017
```

Ensure to update the `<username>` and `<password>`.

Apply the `ferretdb.yaml` file:

```sh
kubectl apply -f ferretdb.yaml
```

Check the status of the FerretDB pod to ensure that it is running:

```sh
kubectl get pods -l app=ferretdb
kubectl get svc -l app=ferretdb
```

Use `kubectl port-forward` to connect to FerretDB from your local machine:

```sh
kubectl port-forward svc/ferretdb 27017:27017
```

If you have `mongosh` installed, you can connect to FerretDB from another terminal with the following command:

```sh
mongosh mongodb://<username>:<password>@127.0.0.1:27017/
```
