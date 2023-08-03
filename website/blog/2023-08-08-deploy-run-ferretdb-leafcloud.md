---
slug: deploy-run-ferretdb-leafcloud
title: 'How to Deploy and Run FerretDB on Leafcloud'
authors: [alex]
description: >
  In this how-to guide, we’ll be showing you how to configure, deploy, and run FerretDB on [Leafcloud](https://www.leaf.cloud/), using the power of Kubernetes and the PostgreSQL Operator.
image: /img/blog/ferretdb-leafcloud.jpg
tags: [compatible applications, tutorial, cloud]
---

![How to Deploy and Run FerretDB on leafcloud](/img/blog/ferretdb-leafcloud.jpg)

In this how-to guide, we'll be showing you how to configure, deploy, and run FerretDB on [Leafcloud](https://www.leaf.cloud/), using the power of Kubernetes and the PostgreSQL Operator.

Building resilient, scalable, and performant applications today can be quite a complex process.
As such, [FerretDB](https://www.ferretdb.io), an open-source database alternative to MongoDB that leverages PostgreSQL as the backend, can be a strategic choice for users looking to avoid vendor lock-in and build a totally flexible and reliable backend for many applications.

On the other hand, Leafcloud is an eco-friendly cloud infrastructure that provides a distributed architecture for you to run your FerretDB application.

This guide will offer you a comprehensive, step-by-step guide to navigating the intricacies of setting up a Kubernetes containerization environment, persistent storage provisioning, and FerretDB deployment on Leafcloud.

Let's get started!

## Setting up FerretDB on top of Leafcloud

We're going to set up FerretDB on a Kubernetes cluster and deploy it to Leafcloud.

### Prerequisites

- Kubectl
- Leafcloud account (ensure you have enough volumes to run the cluster)

## Creating a Kubernetes Cluster in Leafcloud

We will start by creating a Kubernetes cluster, using the OpenStack CLI.
Leafcloud manages and deploys container clusters using the OpenStack Magnum project.

Start by installing the OpenStack CLI using the following command:

```sh
sudo apt update -y
sudo apt install -y python3-pip python3-dev -y
sudo apt install virtualenv -y
virtualenv -p python3 openstack_venv
source openstack_venv/bin/activate
pip install --upgrade pip
pip install python-openstackclient
pip install python-magnumclient
```

Once it's installed, you need to log in to your Leafcloud account to download the OpenStack RC file that contains the environment variables for your command-line client.

[image of Leafcloud download](.../)

After downloading the file, copy and paste the file's contents into a new document as:

```text
~/leafcloudopenrc.sh
```

To set up the configuration for your Leafcloud account, run the following command:

```sh
source ~/leafcloudopenrc.sh
```

It will prompt you to enter your account password.
When you do, you will have access to the OpenStack CLI for your Leafcloud account.

To check, run the command below to see if you can access the server list using the OpenStack CLI:

```sh
openstack server list
```

If you get an authentication error, try to run `source ~/leafcloudopenrc.sh` again and enter the right password for your account.

Leafcloud provides a set of cluster templates that you can take advantage of right away.
See the templates by running:

```sh
openstack coe cluster template list
```

Select the best template for your project.

We'll be using the `K8s-ha-v1.21.2-template-v2.0-rc3` template for the guide since it comes with Kubernetes OpenStack autoscaling, encrypted cinder volumes for containers and hosts (persistent volume claims), and high-availability load balancers.

Using the selected template, we can now create a new cluster.
The keypair, which we associate using the `--keypair` parameter, will be integrated into the hosts, granting us root SSH access (with 'core' as the default user).
The keypairs are essential for encrypting SSH traffic for your instances and enabling secure communications.

To create your cluster, use this command:

```sh
openstack coe cluster create my-k8s-cluster --cluster-template k8s-ha-v1.21.2-template-v2.0-rc3 --keypair <keypair>
```

You may need to wait a few minutes for this process to be complete.
To check up on the installation enter the following:

```sh
openstack coe cluster list
```

The installation is complete once the status of the cluster changes from CREATE_IN_PROGRESS to CREATE_COMPLETE.

Now we need to fetch the configuration file for the cluster.
Do this by running the following:

```sh
openstack coe cluster config my-k8s-cluster
```

A file named 'config' will be downloaded to your home directory.

```sh
export KUBECONFIG=/home/<username>/config
```

Your cluster should now be reachable using Kubectl.
Enter the command to confirm:

```sh
kubectl get nodes -o wide
```

Lastly, we need to set up the storage class to use persistent volume claims (PVCs).
Do this by creating the storageclass.yaml file:

```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: cinder-csi
  annotations:
    storageclass.kubernetes.io/is-default-class: 'true'
provisioner: cinder.csi.openstack.org
```

You can apply it by running:

```sh
kubectl apply -f storageclass.yaml
```

## Creating the Postgres Cluster

Use kustomize postgres-operator 5.4.0 to install the postgres operator.

## Installing the Postgres Operator

In this section, we'll be creating the PostgreSQL cluster.
We'll start by installing the latest version of the PostgreSQL operator, which is essential when setting up extremely reliable PostgreSQL clusters on Kubernetes.

Here we will use kustomize to install the operator, but first, we need to fork the GitHub PostgreSQL Operator website.

```text
YOUR_GITHUB_UN="<your GitHub username>"
git clone --depth 1 "git@github.com:${YOUR_GITHUB_UN}/postgres-operator-examples.git"
cd postgres-operator-examples
```

Using the same CLI, we can install the PostgreSQL Operator from Crunchy Data with the following command:

```sh
kubectl apply -k kustomize/install/namespace
kubectl apply --server-side -k kustomize/install/default
```

This process generates a namespace known as postgres-operator, establishing all necessary objects for PGO deployment.

To monitor the progress of your installation, execute the command below:

```text
kubectl -n postgres-operator get pods \
  --selector=postgres-operator.crunchydata.com/control-plane=postgres-operator \
  --field-selector=status.phase=Running
```

When the PGO Pod is healthy, you will see output that looks like this:

```text
NAME                   READY   STATUS    RESTARTS   AGE
pgo-6f664c9f44-mmptx   1/1     Running   0          10s
```

Before creating the PostgreSQL cluster, we need to modify the ~/kustomize/postgres/postgres.yaml file from the cloned folder, and enable the cluster to use a data volume claim that specifies the 'cinder-csi' storage class and a capacity request of 1Gi.
It should look like this:

```yaml
apiVersion: postgres-operator.crunchydata.com/v1beta1
kind: PostgresCluster
metadata:
  name: hippo
spec:
  image: registry.developers.crunchydata.com/crunchydata/crunchy-postgres:ubi8-15.3-2
  postgresVersion: 15
  instances:
    - name: instance1
      dataVolumeClaimSpec:
        accessModes:
          - 'ReadWriteOnce'
        resources:
          requests:
            storage: 1Gi
        storageClassName: cinder-csi
  backups:
    pgbackrest:
      image: registry.developers.crunchydata.com/crunchydata/crunchy-pgbackrest:ubi8-2.45-2
      repos:
        - name: repo1
          volume:
            volumeClaimSpec:
              accessModes:
                - 'ReadWriteOnce'
              resources:
                requests:
                  storage: 1Gi
              storageClassName: cinder-csi
```

Now let's create the PostgreSQL cluster by running the following command:

```sh
kubectl apply -k kustomize/postgres
```

That command will create the Postgres cluster with the name `hippo` in the `postgres-operator` namespace.
The following command can help you track the cluster's progress:

```sh
kubectl -n postgres-operator describe postgresclusters.postgres-operator.crunchydata.com hippo
```

In a new terminal, run the following command to create a port-forward (if you're getting a connection error, try running `export KUBECONFIG=/home/<username>/config` to configure the environment):

```text
PG_CLUSTER_PRIMARY_POD=$(kubectl get pod -n postgres-operator -o name \
  -l postgres-operator.crunchydata.com/cluster=hippo,postgres-operator.crunchydata.com/role=master)
kubectl -n postgres-operator port-forward "${PG_CLUSTER_PRIMARY_POD}" 5432:5432
```

Go back to the main terminal and establish a connection to your PostgreSQL cluster.

```sh
kubectl exec -it hippo-instance1-mrpt-0 -n postgres-operator -- psql -U postgres
```

We need to configure the PostgreSQL according to FerretDB requirements.
We're going to create a new user and password credential, and then create a database assigned with all privileges to that user.

```sql
CREATE USER <username> WITH PASSWORD <password>;
```

Then create a database named `ferretdb`

```sql
postgres=# CREATE DATABASE ferretdb OWNER ferretdb;
```

Next, grant all privileges to on the new database to the user:

```sql
GRANT ALL PRIVILEGES ON DATABASE ferretdb TO ferretdb;
```

Let's use the postgres database context:

```text
postgres=# \c ferretdb
You are now connected to database "ferretdb" as user "postgres".
```

Finally we're going to set the search_path to ferretdb:

```sql
set search_path to ferretdb;
```

## Deploying FerretDB

Now that we have verified that PostgreSQL is working properly, it's time to set up and install FerretDB to communicate with our PostgreSQL cluster.
To do this, we'll need to create and apply a deployment and service manifest.
This YAML file will define the container specifications for FerretDB, any associated MongoDB components, and the necessary service configuration to establish a connection to PostgreSQL.

The deployment YAML used for this project looks like this:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ferretdb
  namespace: postgres-operator
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
            - name: FERRETDB_POSTGRESQL_URL
              value: postgres://username:password@hippo-ha:5432/ferretdb
      imagePullSecrets:
        - name: ghcr-ferretdb-secret
---
apiVersion: v1
kind: Service
metadata:
  name: ferretdb-service
  namespace: postgres-operator
spec:
  selector:
    app: ferretdb
  ports:
    - name: mongo
      protocol: TCP
      port: 27017
      targetPort: 27017
```

Apply the deployment YAML:

```sh
kubectl apply -f deployment.yaml
```

Ensure that you're on the right path where the deployment YAML file resides.

To check if the pods are running correctly without any errors, run the following command:

```text
kubectl get pods -n postgres-operator
NAME                       READY   STATUS      RESTARTS   AGE
ferretdb-86c45849d-bd6tq   1/1     Running     0          6h
hippo-backup-6shl-qf9mf    0/1     Completed   0          6h
hippo-instance1-mrpt-0     4/4     Running     0          7h
hippo-repo-host-0          2/2     Running     0          7h
pgo-6f664c9f44-mmptx       1/1     Running     0          7h
```

Next you need to connect using your FerretDB URI, where username and password should correspond with the PosgreSQL credentials set earlier, and you can get the FERRETDB SVC by running `kubectl -n postgres-operator get pods`.

```text
kubectl get svc -n postgres-operator
NAME               TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
ferretdb-service   ClusterIP   10.254.17.193    <none>        27017/TCP   4d8h
hippo-ha           ClusterIP   10.254.153.218   <none>        5432/TCP    4d8h
hippo-ha-config    ClusterIP   None             <none>        <none>      4d8h
hippo-pods         ClusterIP   None             <none>        <none>      4d8h
hippo-primary      ClusterIP   None             <none>        5432/TCP    4d8h
hippo-replicas     ClusterIP   10.254.33.136    <none>        5432/TCP    4d8h
```

Great!
Let's use the following command to open up a mongosh shell:

```sh
kubectl -n postgres-operator run mongosh --image=rtsp/mongosh --rm -it -- bash
```

Once the mongosh shell is open, connect to your FerretDB instance using the command:

```sh
mongosh "mongodb://<username>:<password>@{FERRETDB SVC}/ferretdb?authMechanism=PLAIN"
```

And that's it!.
You're connected to FerretDB.

## Basic examples on FerretDB

Let's run a few basic examples using FerretDB:

Insert documents into the database:

```js
db.testing.insertMany([
  { a: 23, b: 'b', c: [1, 5], d: { a: 1 } },
  { a: 1, b: 34, c: '1', d: [3, 5] }
])
```

Now let's read all these documents and see what we get.

```js
db.testing.find()[
  ({
    _id: ObjectId('64ca02e119e6b74d10806107'),
    a: 23,
    b: 'b',
    c: [1, 5],
    d: { a: 1 }
  },
  {
    _id: ObjectId('64ca02e119e6b74d10806108'),
    a: 1,
    b: 34,
    c: '1',
    d: [3, 5]
  })
]
```

We can take a look our data in PostgreSQL to see how the FerretDB conversion works out.

```text
ferretdb=# \dt
                     List of relations
  Schema  |            Name             | Type  |  Owner
----------+-----------------------------+-------+----------
 ferretdb | _ferretdb_database_metadata | table | ferretdb
 ferretdb | testing_eb5f499b            | table | ferretdb
(2 rows)

ferretdb=# table testing_eb5f499b;
                                                                                                                                                                _jsonb
--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 {"a": 23, "b": "b", "c": [1, 5], "d": {"a": 1}, "$s": {"p": {"a": {"t": "int"}, "b": {"t": "string"}, "c": {"i": [{"t": "int"}, {"t": "int"}], "t": "array"}, "d": {"t": "object", "$s": {"p": {"a": {"t": "int"}}, "$k": ["a"]}}, "_id": {"t": "objectId"}}, "$k": ["_id", "a", "b", "c", "d"]}, "_id": "64ca02e119e6b74d10806107"}
 {"a": 1, "b": 34, "c": "1", "d": [3, 5], "$s": {"p": {"a": {"t": "int"}, "b": {"t": "int"}, "c": {"t": "string"}, "d": {"i": [{"t": "int"}, {"t": "int"}], "t": "array"}, "_id": {"t": "objectId"}}, "$k": ["_id", "a", "b", "c", "d"]}, "_id": "64ca02e119e6b74d10806108"}
(2 rows)
~

```

Conclusion

FerretDB offers you the chance to leverage the truly open-source replacement for MongoDB.
And with Leafcloud, you can deploy and run your FerretDB applications using Postgres clusters suitable for production, including persistent storage volumes, automated backups, autoscaling, connection pooling, monitoring, and more.

To learn more about FerretDB, please visit our [GitHub page](ferretdb.io).
