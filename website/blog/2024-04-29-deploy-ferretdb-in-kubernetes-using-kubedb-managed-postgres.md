---
slug: deploy-ferretdb-in-kubernetes-using-kubedb-managed-postgres
title: 'Deploy FerretDB in Kubernetes using KubeDB Managed Postgres'
authors: [alex]
description: >
  KubeDB now provides support for FerretDB. Learn how to deploy a FerretDB instance in Kubernetes using KubeDB managed Postgres.
image: /img/blog/ferretdb-kubedb.jpg
tags: [tutorial, community, cloud, open source]
---

![Deploy FerretDB in Kubernetes using KubeDB Managed Postgres](/img/blog/ferretdb-kubedb.jpg)

[KubeDB](https://kubedb.com/) – a popular Kubernetes database management solution – now provides support for [FerretDB](https://www.ferretdb.com/).

<!--truncate-->

FerretDB is an open source document database that adds MongoDB compatibility to PostgreSQL.
It lets you use your familiar MongoDB syntax and commands with your data stored in PostgreSQL backend.

Over the past few years, Kubernetes has become a popular option for deploying production-ready databases.
Tools like KubeDB simplify the management and automation of database tasks in Kubernetes, including provisioning, monitoring, upgrading, automated scaling and backup, and failure detection.

This blog post will demonstrate how to run and deploy FerretDB in Kubernetes using KubeDB.

## Prerequisites

Ensure to have the following set up.

- Kubernetes cluster ([Minikube](https://minikube.sigs.k8s.io/docs/start/) or [Docker Desktop's Kubernetes](https://www.docker.com/products/docker-desktop/), or any cloud-based service ypu prefer)
- [Get AppsCode License (get cluster ID)](https://appscode.com/issue-license/)
- [Helm](https://helm.sh/)
- [`kubectl`](https://kubernetes.io/docs/reference/kubectl/)
- [`mongosh`](https://www.mongodb.com/docs/mongodb-shell/)

## Get cluster ID

To get the AppsCode license, you need the cluster ID.
Run this command to get the cluster ID.

```sh
kubectl get ns kube-system -o jsonpath='{.metadata.uid}'
```

## Install KubeDB

Use Helm to install KubeDB:

```sh
helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
 --version v2024.2.14 \
 --namespace kubedb --create-namespace \
 --set-file global.license=/path/to/the/license.txt \
 --set global.featureGates.FerretDB=true \
 --wait --burst-limit=10000 --debug
```

Be sure to include the AppsCode license path for the KubeDB installation.

Verify the installation:

```text
$ kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"
NAMESPACE   NAME                                            READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-5c97c8c7f9-lw64s       1/1     Running   0          11m
kubedb      kubedb-kubedb-ops-manager-7b8fc4d7bf-28qk4      1/1     Running   0          11m
kubedb      kubedb-kubedb-provisioner-6c89ddd5d8-fw24w      1/1     Running   0          11m
kubedb      kubedb-kubedb-webhook-server-6fc6c8b44f-pwdvr   1/1     Running   0          11m
kubedb      kubedb-sidekick-86c64c8f59-gvzd8                1/1     Running   0          11m
```

KubeDB provides several installed CRD Groups including FerretDB.
Run `kubectl get crd -l app.kubernetes.io/name=kubedb` command to list them.

## Deploy FerretDB with KubeDB Managed PostgreSQL

Create a namespace for all the FerretDB components.

```sh
kubectl create namespace ferretdemo
```

Next, create the FerretDB Custom Resource `YAML`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferret
  namespace: ferretdemo
spec:
  version: '1.18.0'
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  backend:
    externallyManaged: false
  terminationPolicy: WipeOut
```

The `YAML` file will create a `ferret` resource in the `ferretdemo` namespace using KubeDB.
At the moment, KubeDB only supports version FerretDB v1.18.0.
[Check here for recent versions](https://github.com/FerretDB/FerretDB/releases)

Save the config as `ferret.yaml` and apply it.

```sh
kubectl apply -f ferret.yaml
```

Once it is applied, the FerretDB resource is deployed with the all its objects.

Get the objects in the `ferretdemo` namespace:

```text
$ kubectl get all -n ferretdemo
NAME                              READY   STATUS    RESTARTS   AGE
pod/ferret-0                      1/1     Running   0          4m42s
pod/ferret-pg-backend-0           2/2     Running   0          5m15s
pod/ferret-pg-backend-1           2/2     Running   0          5m10s
pod/ferret-pg-backend-arbiter-0   1/1     Running   0          5m1s
NAME                                TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
service/ferret                      ClusterIP   10.111.78.151   <none>        27017/TCP                    5m18s
service/ferret-pg-backend           ClusterIP   10.99.57.62     <none>        5432/TCP,2379/TCP            5m18s
service/ferret-pg-backend-pods      ClusterIP   None            <none>        5432/TCP,2380/TCP,2379/TCP   5m18s
service/ferret-pg-backend-standby   ClusterIP   10.108.28.132   <none>        5432/TCP                     5m18s
NAME                                         READY   AGE
statefulset.apps/ferret                      1/1     4m42s
statefulset.apps/ferret-pg-backend           2/2     5m15s
statefulset.apps/ferret-pg-backend-arbiter   1/1     5m1s
NAME                                                   TYPE                  VERSION   AGE
appbinding.appcatalog.appscode.com/ferret              kubedb.com/ferretdb   1.18.0    4m42s
appbinding.appcatalog.appscode.com/ferret-pg-backend   kubedb.com/postgres   13.13     5m1s
NAME                                    VERSION   STATUS   AGE
postgres.kubedb.com/ferret-pg-backend   13.13     Ready    5m18s
```

To be sure the `ferret` resource is ready, run this command:

```text
$ kubectl get ferretdb -n ferretdemo ferret
NAME     NAMESPACE    VERSION   STATUS   AGE
ferret   ferretdemo   1.18.0    Ready    9m6s
```

### `port-forward` the Service

Before [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/), list the services created by KubeDB.

```text
$ kubectl get service -n ferretdemo | grep ferret
ferret                      ClusterIP   10.111.78.151   <none>        27017/TCP                    11m
ferret-pg-backend           ClusterIP   10.99.57.62     <none>        5432/TCP,2379/TCP            11m
ferret-pg-backend-pods      ClusterIP   None            <none>        5432/TCP,2380/TCP,2379/TCP   11m
ferret-pg-backend-standby   ClusterIP   10.108.28.132   <none>        5432/TCP                     11m
```

Next, forward the `ferret` Service to port `27017` on your local machine.

```sh
kubectl port-forward -n ferretdemo svc/ferret 27017
```

### Get FerretDB credentials in `Secret`

You need the FerretDB credentials before connecting via `mongosh`.
KubeDB creates and stores the `ferret` Service credentials as a `Secret`.

To get the details, run this command:

```sh
kubectl get secret -n ferretdemo | grep ferret
```

Using `ferret-pg-backend-auth`, get the user credentials.

```sh
echo $(kubectl get secret -n ferretdemo ferret-pg-backend-auth -o jsonpath='{.data.username}' | base64 -d)
echo $(kubectl get secret -n ferretdemo ferret-pg-backend-auth -o jsonpath='{.data.password}' | base64 -d)
```

This will print out the `username` and `password` credentials for the instance.

### Connect to FerretDB via `mongosh`

Using the credentials, connect to FerretDB via mongosh using this format:

```sh
mongosh mongodb://<username>:<password>@<host>:27017/ferretdb?authMechanism=PLAIN'
```

Connect via `mongosh`:

```sh
mongosh 'mongodb://postgres:p.i~glw7q9mdbpQ2@localhost:27017/ferretdb?authMechanism=PLAIN'
Current Mongosh Log ID: 662699b8fa65a75337cb3ec7
Connecting to:  mongodb://<credentials>@localhost:27017/ferretdb?authMechanism=PLAIN&directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.2.2
Using MongoDB:    7.0.42
Using Mongosh:    2.2.2
mongosh 2.2.4 is available for download: https://www.mongodb.com/try/download/shell
For mongosh info see: https://docs.mongodb.com/mongodb-shell/
------
   The server generated these startup warnings when booting
   2024-04-22T17:09:13.234Z: Powered by FerretDB v1.18.0 and PostgreSQL 13.13 on aarch64-unknown-linux-musl, compiled by gcc.
   2024-04-22T17:09:13.235Z: Please star us on GitHub: https://github.com/FerretDB/FerretDB.
   2024-04-22T17:09:13.235Z: The telemetry state is undecided.
   2024-04-22T17:09:13.235Z: Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.io.
------
ferretdb>
```

Let's run some commands in the database.
Start by inserting a document record into a `weather` collection as shown below.

```json5
db.weather.insertMany([
    {
        date: new Date("2024-04-22"),
        location: {
            city: "New York",
            country: "USA",
            coordinates: { lat: 40.7128, lon: -74.0060 }
        },
        weather: {
            temperature: 18,
            conditions: "Cloudy",
            wind_speed: 12,
            humidity: 80
        },
        remarks: "Possible light rain in the evening."
    }
]);
```

Suppose you want to update the humidity level in New York where the wind speed was more than 10 km/h:

```json5
ferretdb> db.weather.updateMany(
...     { "location.city": "New York", "weather.wind_speed": { $gt: 10 } },
...     { $set: { "weather.humidity": 85 } }
... );
{
  acknowledged: true,
  insertedId: null,
  matchedCount: 1,
  modifiedCount: 1,
  upsertedCount: 0
}
ferretdb> db.weather.find()
[
  {
    _id: ObjectId('66278976fba61a5fec8bad82'),
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

FerretDB stores the data in the `ferret-pg-backend` PostgreSQL using mongosh.
Let's exec into the Postgres database to view the record.

```sh
% kubectl exec -it -n ferretdemo ferret-pg-backend-0 -- bash -c "psql -d ferretdb"
```

In Postgres, set the `SEARCH_PATH` to `ferretdb` and list the record in the `weather_36404793` table.

```text
ferretdb=# set SEARCH_PATH to ferretdb;
SET
ferretdb=# \dt
                     List of relations
  Schema  |            Name             | Type  |  Owner
----------+-----------------------------+-------+----------
 ferretdb | _ferretdb_database_metadata | table | postgres
 ferretdb | weather_36404793            | table | postgres
(2 rows)
ferretdb=# SELECT * FROM weather_36404793;
                                                                                                                                                                                                                                                                                                                                                                                                                                                                            _jsonb
--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 {"$s": {"p": {"_id": {"t": "objectId"}, "date": {"t": "date"}, "remarks": {"t": "string"}, "weather": {"t": "object", "$s": {"p": {"humidity": {"t": "int"}, "conditions": {"t": "string"}, "wind_speed": {"t": "int"}, "temperature": {"t": "int"}}, "$k": ["temperature", "conditions", "wind_speed", "humidity"]}}, "location": {"t": "object", "$s": {"p": {"city": {"t": "string"}, "country": {"t": "string"}, "coordinates": {"t": "object", "$s": {"p": {"lat": {"t": "double"}, "lon": {"t": "double"}}, "$k": ["lat", "lon"]}}}, "$k": ["city", "country", "coordinates"]}}}, "$k": ["_id", "date", "location", "weather", "remarks"]}, "_id": "66278976fba61a5fec8bad82", "date": 1713744000000, "remarks": "Possible light rain in the evening.", "weather": {"humidity": 85, "conditions": "Cloudy", "wind_speed": 12, "temperature": 18}, "location": {"city": "New York", "country": "USA", "coordinates": {"lat": 40.7128, "lon": -74.006}}}
(1 row)
ferretdb=#
```

## Deploy FerretDB with an externally managed PostgreSQL

So far, we have shown how to set up FerretDB using KubeDB Managed PostgreSQL.
However, if you prefer an external PostgreSQL server as your backend, this is entirely possible.

The YAML configuration provided below outlines how to integrate FerretDB with a PostgreSQL instance managed externally.

```js
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferretdb-external
  namespace: ferretdemo
spec:
  version: "1.18.0"
  authSecret:
    externallyManaged: true
    name: ha-postgres-auth
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  backend:
    externallyManaged: true
    postgres:
      service:
        name: ha-postgres
        namespace: ferretdemo
        pgPort: 5432
  terminationPolicy: WipeOut
```

`spec.postgres.service` contains the service details for the user's external PostgreSQL that exists within the cluster.
`spec.authSecret.name` refers to the name of the authentication secret for accessing the user's external PostgreSQL database.

## Support

This blog post shows you how to deploy a [FerretDB](https://www.ferretdb.com/) instance in Kubernetes using KubeDB.
Ensure to experiment and try it out.
If you have any questions or just want to contact us, please do so via the [FerretDB Slack channel here](https://join.slack.com/t/ferretdb/shared_invite/zt-zqe9hj8g-ZcMG3~5Cs5u9uuOPnZB8~A).
