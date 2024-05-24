---
slug: deploy-ferretdb-kubernetes-taikun-cloudworks
title: 'Learn to Deploy FerretDB on Kubernetes on Taikun CloudWorks'
authors: [alex]
description: >
  Learn how to deploy FerretDB with a Percona Distribution for PostgreSQL as the backend on a Kubernetes cluster on Taikun CloudWorks
image: /img/blog/ferretdb-taikun.jpg
tags: [compatible applications, tutorial, cloud, postgresql tools, open source]
---

![Learn to Deploy FerretDB on Kubernetes on Taikun CloudWorks](/img/blog/ferretdb-taikun.jpg)

Are you looking to set up a highly scalable and reliable enterprise Kubernetes cluster for your [FerretDB](https://www.ferretdb.com/) databases?

<!--truncate-->

[Taikun CloudWorks](https://taikun.cloud/taikun-cloudworks/) offers an open-source solution that simplifies the deployment and management of Kubernetes clusters.
This setup is ideal for developers seeking robust infrastructure for their FerretDB deployments, which is critical for modern enterprise level applications.

You can package and deploy FerretDB on Taikun CloudWorks.
Taikun's simplified Kubernetes platform offers a unified and user-friendly interface to deploy, manage, and monitor Kubernetes clusters across multiple cloud environments, including AWS, Azure, OpenStack, GCP.

By the end of this blog post, you will have a comprehensive understanding of how to deploy FerretDB with a [Percona Distribution for PostgreSQL](https://www.percona.com/postgresql/software/postgresql-distribution) as the backend on a Kubernetes cluster on Taikun CloudWorks.

## Prerequisites

- Kubernetes cluster configured and running on Taikun CloudWorks
- [kubectl](https://kubernetes.io/docs/reference/kubectl/)

## Set up a Kubernetes cluster in Taikun cloud

Start by creating a Taikun project.
It will act as a central management place for the Kubernetes cluster.
Read the following documentation to learn how to create a Kubernetes cluster in Taikun:

- [Create a project](https://taikun.cloud/docs/taikun-project-creation/)
- [Create Kubernetes Cluster in Taikun](https://taikun.cloud/docs/creating-kubernetes-cluster/)
- [Access Cluster Kubeconfig](https://taikun.cloud/docs/accessing-cluster-kubeconfig/)

Export your Kubeconfig file to access the Kubernetes cluster:

```sh
export KUBECONFIG=/Users/<path>/<to>/kubeconfig.yaml
```

Next, create a namespace isolate the workloads and resources in custom namespace.

```sh
kubectl create namespace ferretdb
```

### Install Percona PostgreSQL Operator

In this blog post, we will use [Percona Distribution for PostgreSQL](https://www.percona.com/postgresql/software/postgresql-distribution) as the FerretDB PostgreSQL backend.
Percona Distribution for PostgreSQL offers consists of several open source components that enable high availability, robust performance, backup, and scalability for enterprise level deployments.

The Percona PostgreSQL Operator enables simplified configuration and management of a Percona Distribution for PostgreSQL cluster in a Kubernetes-based environment on-premises or in the cloud.

You can install the Operator with `kubectl` on the Kubernetes environment setup.

Apply the Percona PostgreSQL Operator within the `ferretdb` namepace:

```sh
kubectl apply --server-side -f https://raw.githubusercontent.com/percona/percona-postgresql-operator/v2.3.1/deploy/bundle.yaml -n ferretdb
```

You should see output indicating the resources have been server-side applied:

```sh
customresourcedefinition.apiextensions.k8s.io/perconapgbackups.pgv2.percona.com serverside-applied
customresourcedefinition.apiextensions.k8s.io/perconapgclusters.pgv2.percona.com serverside-applied
customresourcedefinition.apiextensions.k8s.io/perconapgrestores.pgv2.percona.com serverside-applied
customresourcedefinition.apiextensions.k8s.io/postgresclusters.postgres-operator.crunchydata.com serverside-applied
serviceaccount/percona-postgresql-operator serverside-applied
role.rbac.authorization.k8s.io/percona-postgresql-operator serverside-applied
rolebinding.rbac.authorization.k8s.io/service-account-percona-postgresql-operator serverside-applied
deployment.apps/percona-postgresql-operator serverside-applied
alexanderfashakin@Mac-mini percona-postgresql-operator-2.3.1 % kubectl apply -f deploy/cr.yaml -n ferretdb
perconapgcluster.pgv2.percona.com/cluster1 created
```

### Deploy the PostgreSQL cluster

You need to configure the PostgreSQL database according to FerretDB requirements.
That means setting up a `ferretdb` database and user credentials with the necessary privileges to the database.

With that in mind, let's adjust the `cr.yaml` file to reflect that so it creates the database, user, password and other credentials in `Secret`.

```yaml
apiVersion: pgv2.percona.com/v2
kind: PerconaPGCluster
metadata:
  name: cluster1
spec:
  crVersion: 2.3.1
  users:
    - name: ferretuser
      databases:
        - ferretdb
  image: percona/percona-postgresql-operator:2.3.1-ppg16-postgres
  imagePullPolicy: Always
  postgresVersion: 16
  instances:
    - name: instance1
      replicas: 3
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 1
              podAffinityTerm:
                labelSelector:
                  matchLabels:
                    postgres-operator.crunchydata.com/data: postgres
                topologyKey: kubernetes.io/hostname
      dataVolumeClaimSpec:
        accessModes:
          - ReadWriteOnce
        storageClassName: cinder-csi
        resources:
          requests:
            storage: 1Gi
  proxy:
    pgBouncer:
      replicas: 3
      image: percona/percona-postgresql-operator:2.3.1-ppg16-pgbouncer
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 1
              podAffinityTerm:
                labelSelector:
                  matchLabels:
                    postgres-operator.crunchydata.com/role: pgbouncer
                topologyKey: kubernetes.io/hostname
  backups:
    pgbackrest:
      image: percona/percona-postgresql-operator:2.3.1-ppg16-pgbackrest
      repoHost:
        affinity:
          podAntiAffinity:
            preferredDuringSchedulingIgnoredDuringExecution:
              - weight: 1
                podAffinityTerm:
                  labelSelector:
                    matchLabels:
                      postgres-operator.crunchydata.com/data: pgbackrest
                  topologyKey: kubernetes.io/hostname
      manual:
        repoName: repo1
        options:
          - --type=full
      repos:
        - name: repo1
          schedules:
            full: '0 0 * * 6'
          volume:
            volumeClaimSpec:
              accessModes:
                - ReadWriteOnce
              storageClassName: cinder-csi
              resources:
                requests:
                  storage: 1Gi
  pmm:
    enabled: false
    image: percona/pmm-client:2.41.0
    secret: cluster1-pmm-secret
    serverHost: monitoring-service
```

Apply the custom resource (CR) YAML to create the PostgreSQL cluster:

```sh
kubectl apply -f deploy/cr.yaml -n ferretdb
```

Check the status of the cluster:

```sh
kubectl get pg -n ferretdb
```

Initially, the status will be `initializing`:

```sh
NAME       ENDPOINT                          STATUS         POSTGRES   PGBOUNCER   AGE
cluster1   cluster1-pgbouncer.ferretdb.svc   initializing                          13s
```

After some time, the status should change to `ready`:

```sh
kubectl get pg -n ferretdb
NAME       ENDPOINT                          STATUS   POSTGRES   PGBOUNCER   AGE
cluster1   cluster1-pgbouncer.ferretdb.svc   ready    3          3           5m10s
```

Run `kubectl get pods -n ferretdb` to ensure all pods are running – you should see output indicating that the pods are in the `Running` state:

```text
NAME                                           READY   STATUS    RESTARTS   AGE
cluster1-backup-p25l-xh6dm                     1/1     Running   0          52s
cluster1-instance1-dtj6-0                      3/4     Running   0          97s
cluster1-instance1-j9jv-0                      4/4     Running   0          97s
cluster1-instance1-k8tj-0                      4/4     Running   0          97s
cluster1-pgbouncer-87c4d584d-dj52c             2/2     Running   0          97s
cluster1-pgbouncer-87c4d584d-ks87p             2/2     Running   0          97s
cluster1-pgbouncer-87c4d584d-pxlg8             2/2     Running   0          97s
cluster1-repo-host-0                           2/2     Running   0          97s
percona-postgresql-operator-55fff7dd8b-x7kjx   1/1     Running   0          147m
```

### Apply FerretDB deployment

Deploy FerretDB using the following configuration:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ferretdb
  namespace: ferretdb
  labels:
    app: ferretdb
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
          image: ghcr.io/ferretdb/ferretdb
          ports:
            - containerPort: 27017
          env:
            - name: POSTGRES_HOST
              valueFrom:
                secretKeyRef:
                  name: cluster1-pguser-ferretuser
                  key: host
            - name: FERRETDB_POSTGRESQL_URL
              value: postgres://postgres@$(POSTGRES_HOST):5432/ferretdb
---
apiVersion: v1
kind: Service
metadata:
  name: ferretdb-service
  namespace: ferretdb
spec:
  selector:
    app: ferretdb
  ports:
    - name: mongo
      protocol: TCP
      port: 27017
      targetPort: 27017
```

Apply the yaml file:

```sh
kubectl apply -f ferretdeploy.yaml -n ferretdb
```

Check the status of the deployment:

```sh
kubectl get pods -n ferretdb
```

You should see the `ferretdb` pod in the `Running` state:

```sh
NAME                                           READY   STATUS    RESTARTS   AGE
ferretdb-5f6d9dfd59-szl78                      1/1     Running   0          9s
```

### Connect to FerretDB

Before connecting to FerretDB, let's get the MongoDB URI user credentials needed to set up connection.

Start by retrieving the password for the `ferretuser` stored in `secret` and the `ferretdb-service` address and port:

```sh
kubectl get secret cluster1-pguser-ferretuser -n ferretdb -o jsonpath="{.data.password}" | base64 --decode
```

Check the services created in the namespace:

```sh
kubectl get svc -n ferretdb
```

Example output:

```sh
NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
kubectl get svc -n ferretdb
NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
cluster1-ha          ClusterIP   10.233.60.147   <none>        5432/TCP    9m49s
cluster1-ha-config   ClusterIP   None            <none>        <none>      9m49s
cluster1-pgbouncer   ClusterIP   10.233.36.128   <none>        5432/TCP    9m48s
cluster1-pods        ClusterIP   None            <none>        <none>      9m49s
cluster1-primary     ClusterIP   None            <none>        5432/TCP    9m49s
cluster1-replicas    ClusterIP   10.233.48.145   <none>        5432/TCP    9m49s
ferretdb-service     ClusterIP   10.233.22.188   <none>        27017/TCP   3m36s
```

The ferretdb-service acts as the ClusterIP service, and allows you to connect to the FerretDB instance within the Kubernetes cluster (`10.233.22.188:27017`).

Run a MongoDB shell to connect to FerretDB:

```sh
kubectl -n ferretdb run mongosh --image=rtsp/mongosh --rm -it -- bash
```

Once inside the container, connect to FerretDB using the URI:

```sh
mongosh 'mongodb://ferretuser:<password>@10.233.22.188:27017/ferretdb?authMechanism=PLAIN'
```

You should see a successful connection message.
To verify, you can insert and query a document:

```json5
ferretdb> db.test.insert({a:34})
{
  acknowledged: true,
  insertedIds: { '0': ObjectId('664ff91b7207189218a26a13') }
}

ferretdb> db.test.find()
[ { _id: ObjectId('664ff91b7207189218a26a13'), a: 34 } ]
```

### Check data in PostgreSQL

FerretDB stores all data on PostgreSQL.
So if you find yourself wondering how that looks in PostgreSQL, you can check it out:

```sh
FERRETUSER_URI=kubectl get secret cluster1-pguser-ferretuser --namespace ferretdb -o jsonpath='{.data.uri}' | base64 --decode
kubectl run -i --rm --tty pg-client --image=perconalab/percona-distribution-postgresql:16 --restart=Never -- psql $FERRETUSER_URI
```

```text
ferretdb=> set search_path to ferretdb;
SET
ferretdb=> \dt
                      List of relations
  Schema  |            Name             | Type  |   Owner
----------+-----------------------------+-------+------------
 ferretdb | _ferretdb_database_metadata | table | ferretuser
 ferretdb | test_afd071e5               | table | ferretuser
(2 rows)

ferretdb=> table test_afd071e5;
                                                            _jsonb
------------------------------------------------------------------------------------------------------------------------------
 {"a": 34, "$s": {"p": {"a": {"t": "int"}, "_id": {"t": "objectId"}}, "$k": ["_id", "a"]}, "_id": "664ff91b7207189218a26a13"}
(1 row)
```

## Clean up resources

You can clean up all used resources by just deleting the namespace:

```sh
kubectl delete namespace ferretdb
```

## Conclusion

Deploying FerretDB and Percona Distribution for PostgreSQL on a Kubernetes cluster using Taikun CloudWorks, offers a highly scalable, reliable, and open-source solution for managing enterprise databases.

By following the steps outlined in this guide, you can ensure your databases are resilient and performant on a simplified Kubernetes platform.

To learn read more about FerretDB and other solutions in this guide, explore:

- [FerretDB GitHub repository](https://github.com/FerretDB/FerretDB)
- [Taikun CloudWorks documentation](https://www.taikun.cloud/documentation)
- [Percona Distribution for PostgreSQL documentation](https://www.percona.com/software/postgresql-distribution)
