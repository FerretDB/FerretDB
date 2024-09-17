---
slug: run-ferretdb-cybertec-postgres-kubernetes
title: 'Run FerretDB on Kubernetes with CyberTec Postgres'
authors: [alex]
description: >
  Learn how to run FerretDB on Kubernetes using CyberTec Postgres operator for a truly open source alternative to MongoDB.
image: /img/blog/ferretdb-cybertec-postgres.jpg
tags: [compatible applications, tutorial, cloud, postgresql tools, open source]
---

![Run FerretDB and Postgres Cluster using CyberTec Postgres on Kubernetes](/img/blog/ferretdb-cybertec-postgres.jpg)

<!--truncate-->

Are you looking to replace your MongoDB instance with an open source solution?
You can run FerretDB on Kubernetes with CyberTec Postgres.
As a truly open source alternative to MongoDB, FerretDB removes any risk of vendor lock-in, unlimited flexibility, lower costs, and the high reliance of Postgres.

With [CyberTec Postgres operator](https://github.com/cybertec-postgresql/CYBERTEC-pg-operator), you can create enterprise/production ready Postgres database environment with automatic backups, rolling update procedures, auto-failover and self-healing.

In this blog post, you'll learn to set up FerretDB with CyberTec Postgres as the backend on Kubernetes.

This is one of a series of a series of Postgres operator solutions you can use to setup a Postgres cluster on Kubernetes for your instance.
Check out some of the others:

## Prerequisites

- Running cluster (on minikube)
- kubectl
- Helm
- psql

## Guide to setup CyberTech Postgres operator

Start by downloading/cloning the the CyberTec project

```sh
GITHUB_USER='USERNAME'
git clone https://github.com/$GITHUB_USER/CYBERTEC-operator-tutorials.git
cd CYBERTEC-operator-tutorials
```

Ensure you have a cluster running.
Next, create a namespace `cpo` for the project.

```sh
kubectl create namespace cpo
```

Use Helm to install the CyberTec Postgres Operator

```sh
 helm install cpo -n cpo setup/helm/operator/.
```

Within the cloned repository `cluster-tutorials`, use the following command to set up a single postgres cluster and apply it.

```sh
 kubectl apply -f cluster-tutorials/single-cluster/postgres.yaml -n cpo
```

Ensure to check to see that all the pods are running:

```sh
kubectl get pods -n cpo
```

It may take a few minutes for the pods to be ready.
Ensure that the pods are in a running state before proceeding.

```text
$ kubectl get pods -n cpo
NAME                                 READY   STATUS              RESTARTS   AGE   IP       NODE       NOMINATED NODE   READINESS GATES
postgres-operator-86f6fb46bd-s9qgz   0/1     ContainerCreating   0          71s   &lt;none>   minikube   &lt;none>           &lt;none>
```

## Enable traffic to Postgres cluster

Awesome!
Now enable traffic to the PostgreSQL server so that you can connect to the database.
You can do that by patching svc to allow traffic via `NodePort`.

```sh
kubectl patch svc cluster-1 -n cpo -p '{"spec": {"type": "NodePort"}}'
```

```sh
kubectl get svc -n cpo
```

Output should look like this:

```text
NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
cluster-1               NodePort    10.97.229.173   <none>        5432:30399/TCP   17m
cluster-1-clusterpods   ClusterIP   None            <none>        <none>           17m
cluster-1-repl          ClusterIP   10.97.246.53    <none>        5432/TCP         17m
```

Now that the Postgres clusters are set up, you need the user credential secret to connect to the database instance:

```sh
kubectl get secret -n cpo postgres.cluster-1.credentials.postgresql.cpo.opensource.cybertec.at -o jsonpath='{.data}' | jq '.|map_values(@base64d)'
```

Output should look like this:

```text
{
  "password": "Upp9khrJDyeC7XNn6chhtiFyrpTVF5g1udtBihdd2WSrAliw9b2t8takWwHqy4pd",
  "username": "postgres"
}
```

## Create and deploy FerretDB pods and service

With the Postgres instance running, you need to create the FerretDB instance.
The following `yaml` file sets up a FerretDB deployment and connects to the Postgres database using the `FERRETDB_POSTGRESQL_URL`.
The host name for the Postggres instance is `cluster-1` following the naming convention of the cluster-name.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ferretdb
  namespace: cpo
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
          image: ghcr.io/ferretdb/ferretdb:latest
          ports:
            - containerPort: 27017
          env:
            - name: FERRETDB_POSTGRESQL_URL
              value: postgres://postgres@cluster-1:5432/postgres

---
apiVersion: v1
kind: Service
metadata:
  name: ferretdb-service
  namespace: cpo
spec:
  selector:
    app: ferretdb
  ports:
    - name: mongo
      protocol: TCP
      port: 27017
      targetPort: 27017
```

Apply the deployment by running:

```sh
kubectl apply -f ferretdb-deployment.yaml
```

## Access FerretDB instance via `mongosh`

Create a temp `mongosh` pod:

```sh
kubectl run -it --rm --image=mongo:latest mongo-client -- bash
```

Connect with credentials:

```sh
mongosh "mongodb://postgres:password@host:27017/postgres?authMechanism=PLAIN"
```

Let's run some CRUD commands to see how FerretDB enables you to replace MongoDB and run your familiar queries and operations.

## View data in Postgres via psql

If you wnat to know how the data is stored on Postgres, you can connect to the database via psql using the database and user credentials:

```sh
PGPASSWORD=password psql -h 127.0.0.1 -p 5431 -U postgres
```

## Conclusion

You can further optimize the CyberTec Postgres setup however way you prefer or integrate monitoring tools such as Grafana, etc.
To learn how to do that, follow the CyberTec Postgres documentation page.

If you have any question about FerretDB, reach out to us on any of our community pages.
