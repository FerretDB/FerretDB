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

Are you looking to replace your MongoDB instance with an open source solution?
You can do that by running FerretDB on Kubernetes with CyberTec Postgres.

<!--truncate-->

[FerretDB](https://www.ferretdb.com/) offers a truly open-source document database alternative to MongoDB, removing the risks associated with vendor lock-in and the cost implications of proprietary licenses.
By using Postgres as the underlying database, you gain the reliability, scalability, and extensive feature set of one of the most reliable databases available today.
With [CyberTec Postgres operator](https://www.cybertec-postgresql.com/en/), you can enable production-grade features such as auto-failover, rolling updates, and automated backups, suitable for enterprise-level workloads.

In this blog post, you'll learn to set up FerretDB with CyberTec Postgres as the backend on Kubernetes.

## Prerequisites

- Kubernetes cluster (use [Minikube](https://minikube.sigs.k8s.io/docs/start/) for local development and testing)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Helm](https://helm.sh/docs/intro/install/)
- [psql](https://www.postgresql.org/docs/current/app-psql.html)

## Guide to setup CyberTech Postgres operator

FerretDB is a truly open source MongoDB alternative database that uses Postgres as a backend.
To run FerretDB on Kubernetes, you need to set up a Postgres cluster, and you can do that using the CyberTec Postgres operator.

### Download the CyberTec project

When cloning the CyberTec Postgres operator repository, you can either fork it (if you intend to make customizations) or clone it directly for use in your environment.

Start by forking the [CyberTec project](https://github.com/cybertec-postgresql/CYBERTEC-operator-tutorials.git).

Then clone the forked repository:

```sh
GITHUB_USER='USERNAME'
git clone https://github.com/$GITHUB_USER/CYBERTEC-operator-tutorials.git
cd CYBERTEC-operator-tutorials
```

This folder is where the core Helm chart for installing the Postgres operator resides, and you will use it in the subsequent steps to install the operator and configure your Postgres cluster.

### Create a namespace for the project

Ensure you have a Kubernetes cluster running.
Next, create a namespace `cpo` for the project.

```sh
kubectl create namespace cpo
```

### Install the CyberTec Postgres operator

Use Helm to install the CyberTec Postgres Operator.

```sh
helm install cpo -n cpo setup/helm/operator/
```

This action will install the CyberTec Postgres operator in the `cpo` namespace.

Next, you need to set up a single Postgres cluster.

### Create a single Postgres cluster

The required `yaml` file is already present in the `cluster-tutorials/single-cluster` directory.
kubectl apply -f cluster-tutorials/single-cluster/postgres.yaml -n cpo

```sh
kubectl apply -f cluster-tutorials/single-cluster/postgres.yaml -n cpo
```

When you apply the `postgres.yaml` file, the CyberTec Postgres operator automates the deployment and lifecycle management of the Postgres cluster.
The operator continuously monitors the cluster and performs self-healing actions if nodes go down.
For further customization, you can edit the `postgres.yaml` file to adjust parameters like replica count, resource limits, or backup schedules.

Ensure to check to see that all the pods are running:

```sh
kubectl get pods -n cpo
```

The pods should be in a running state, like this:

```text
NAME                                 READY   STATUS    RESTARTS   AGE
cluster-1-0                          1/1     Running   0          5m10s
postgres-operator-78d4fdc97b-mp49x   1/1     Running   0          5m51s
```

## Enable external traffic to Postgres cluster

The Postgres cluster is running in a private network and you need to enable traffic to the Postgres server so that you can connect to the database.

You can do that by patching `svc` to allow traffic via `NodePort`.

```sh
kubectl patch svc cluster-1 -n cpo -p '{"spec": {"type": "NodePort"}}'
```

By default, services in Kubernetes are only accessible within the cluster's internal network.
Patching the service to `NodePort` exposes it to traffic from outside the cluster on a specific port.
This allows you to connect to the Postgres instance from outside the Kubernetes environment.
In a production setup, you might want to consider more scalable options like using a `LoadBalancer` service type, or configuring Ingress for managing external traffic in a secure way.

Check the service to see the port that is open:

```sh
kubectl get svc -n cpo
```

Output should look like this:

```text
NAME                    TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
cluster-1               NodePort    10.105.156.194   <none>        5432:31263/TCP   9m43s
cluster-1-clusterpods   ClusterIP   None             <none>        <none>           9m43s
cluster-1-repl          ClusterIP   10.101.77.119    <none>        5432/TCP         9m43s
```

### Get the Postgres user credentials

Now that the Postgres clusters are set up, you need the user credential stored in `Secret` to connect to the database instance:

```sh
kubectl get secret -n cpo postgres.cluster-1.credentials.postgresql.cpo.opensource.cybertec.at -o jsonpath='{.data}' | jq '.|map_values(@base64d)'
```

Output should look like this:

```json
{
  "password": "naS0UMX4ajDUtFJZ2Zntwxscn5tnBnLsrDolSXqKOcxvaYkjAdjWRCRQhybbyORN",
  "username": "postgres"
}
```

The Postgres instance is ready for use as a backend for FerretDB.

## Create and deploy FerretDB pods and service

With the Postgres instance running, you need to create the FerretDB instance.
The following `yaml` file sets up a FerretDB deployment and connects to the Postgres database using the `FERRETDB_POSTGRESQL_URL`.

The host name for the Postgres instance is `cluster-1`.

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

This will create a FerretDB pod and service in the `cpo` namespace.

Run `kubectl get pods -n cpo` to see that the pod is running.

The output should look like this:

```text
NAME                                 READY   STATUS    RESTARTS   AGE
cluster-1-0                          1/1     Running   0          17m
ferretdb-8c5468db-mmqtr              1/1     Running   0          38s
postgres-operator-78d4fdc97b-mp49x   1/1     Running   0          18m
```

You can also view the service by running `kubectl get svc -n cpo`.

The output should look like this:

```text
% kubectl get svc -n cpo
NAME                    TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
cluster-1               ClusterIP   10.105.156.194   <none>        5432/TCP    36m
cluster-1-clusterpods   ClusterIP   None             <none>        <none>      36m
cluster-1-repl          ClusterIP   10.101.77.119    <none>        5432/TCP    36m
ferretdb-service        ClusterIP   10.102.252.125   <none>        27017/TCP   19m
```

## Access FerretDB instance via `mongosh`

You can connect to the FerretDB instance using `mongosh`.

Start by creating a temp `mongosh` pod:

```sh
kubectl run -it --rm --image=mongo:latest mongo-client -- bash
```

Connect to the FerretDB instance using the Postgres credentials generated earlier:

```sh
mongosh "mongodb://postgres:<password>@<host>:27017/postgres?authMechanism=PLAIN"
```

Ensure to use the password generated from your the user credential in `Secret`.
The host is the cluster IP of the FerretDB service (e.g. `10.102.252.125`).

## Run CRUD operations on FerretDB

Let's run some CRUD commands to see how FerretDB enables you to replace MongoDB and run your familiar queries and operations.

Start by inserting some documents into a `weather` collection:

```js
db.weather.insertOne([
  {
    date: new Date('2024-04-22'),
    location: {
      city: 'New York',
      country: 'USA',
      coordinates: { lat: 40.7128, lon: -74.006 }
    },
    weather: {
      temperature: 18,
      conditions: 'Cloudy',
      wind_speed: 12,
      humidity: 80
    },
    remarks: 'Possible light rain in the evening.'
  }
])
```

Query the collection to see the inserted document:

```js
db.weather.find()
```

Run an update operation to update the humidity of the document:

```js
db.weather.updateMany(
  { 'location.city': 'New York', 'weather.wind_speed': { $gt: 10 } },
  { $set: { 'weather.humidity': 85 } }
)
```

Run `db.weather.find()` on the collection to see the updated document:

```json5
[
  {
    _id: ObjectId('66f008484f7a5c7f5a1681ed'),
    date: ISODate('2024-04-22T00:00:00.000Z'),
    location: {
      city: 'New York',
      country: 'USA',
      coordinates: { lat: 40.7128, lon: -74.006 }
    },
    weather: {
      temperature: 18,
      conditions: 'Cloudy',
      wind_speed: 12,
      humidity: 85
    },
    remarks: 'Possible light rain in the evening.'
  }
]
```

## View data in Postgres via `psql`

FerretDB stores the data in a JSONB format in the Postgres database.

If you want to know how this looks, port-forward the Postgres service to your local machine:

```sh
kubectl port-forward svc/cluster-1 5432:5432 -n cpo
```

Connect to the database via `psql` using the database and user credentials:

```sh
PGPASSWORD=<password> psql -h 127.0.0.1 -p 5432 -U postgres
```

If it's not set, set the `SEARCH_PATH` to `postgres` and list the record in the `weather_36404793` table.

```text
postgres=# \dt
                     List of relations
  Schema  |            Name             | Type  |  Owner
----------+-----------------------------+-------+----------
 postgres | _ferretdb_database_metadata | table | postgres
 postgres | weather_36404793            | table | postgres
(2 rows)

postgres=# SELECT * FROM weather_36404793;
                                                                                                                                                                                                                                                                                                                                                                                                                                                                            _jsonb
--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 {"$s": {"p": {"_id": {"t": "objectId"}, "date": {"t": "date"}, "remarks": {"t": "string"}, "weather": {"t": "object", "$s": {"p": {"humidity": {"t": "int"}, "conditions": {"t": "string"}, "wind_speed": {"t": "int"}, "temperature": {"t": "int"}}, "$k": ["temperature", "conditions", "wind_speed", "humidity"]}}, "location": {"t": "object", "$s": {"p": {"city": {"t": "string"}, "country": {"t": "string"}, "coordinates": {"t": "object", "$s": {"p": {"lat": {"t": "double"}, "lon": {"t": "double"}}, "$k": ["lat", "lon"]}}}, "$k": ["city", "country", "coordinates"]}}}, "$k": ["_id", "date", "location", "weather", "remarks"]}, "_id": "66f008484f7a5c7f5a1681ed", "date": 1713744000000, "remarks": "Possible light rain in the evening.", "weather": {"humidity": 85, "conditions": "Cloudy", "wind_speed": 12, "temperature": 18}, "location": {"city": "New York", "country": "USA", "coordinates": {"lat": 40.7128, "lon": -74.006}}}
(1 row)
```

## Conclusion

Using open-source solutions like FerretDB and CyberTec Postgres, you can migrate from MongoDB to a Kubernetes-based setup without vendor lock-in.
This gives you complete control over your infrastructure, while taking advantage of Postgres advanced features, scalability, and reliability.
Be sure to follow the [CyberTec Postgres documentation](https://cybertec-postgresql.github.io/CYBERTEC-pg-operator/documentation/how-to-use/installation/) for further optimizations and advanced configurations.

This is one of a series of a series of Postgres operator solutions you can use to setup a Postgres cluster on Kubernetes for your FerretDB instance.
Check out some of the others:

- [Run FerretDB and Postgres cluster using CloudNativePG on Kubernetes](https://blog.ferretdb.io/run-ferretdb-cloudnativepg-kubernetes/)
- [Learn to deploy FerretDB with Percona Distribution for PostgreSQL on Kubernetes on Taikun CloudWorks](https://blog.ferretdb.io/deploy-ferretdb-kubernetes-taikun-cloudworks/)
- [How to deploy and run FerretDB with CrunchyData Postgres operator on Leafcloud](https://blog.ferretdb.io/deploy-run-ferretdb-leafcloud/)
- [How to run FerretDB on top of StackGres](https://blog.ferretdb.io/run-ferretdb-on-stackgres/)

To start [migrating from MongoDB to FerretDB, follow the steps in this guide](https://docs.ferretdb.io/migration/migrating-from-mongodb/).
