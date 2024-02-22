---
slug: run-deploy-ferretdb-kubernetes-using-pulumi
title: 'How to run and deploy FerretDB in Kubernetes using Pulumi'
authors: [alex]
description: >
  Here, you’ll learn to set up a fully managed, scalable FerretDB on top of StackGres using using Pulumi.
image: /img/blog/ferretdb-pulumi.jpg
tags: [tutorial, community, postgresql tools, open source, cloud]
---

![How to run and deploy FerretDB in Kubernetes using Pulumi](/img/blog/ferretdb-pulumi.jpg)

Setting up a scalable, reliable, and highly performant database application is essential for all production environments.
But before then, you still need robust testing environments for your operations and teams.

<!--truncate-->

Running your databases on Kubernetes can be a way to quickly provision, test, and drop databases using a CI/CD pipeline.

[Pulumi](https://www.pulumi.com/), an Infrastructure as Code (IaC) platform, can be quite useful in building, managing, and deploying infrastructure in your favorite programming language and on numerous public clouds or cloud-native platforms, including Kubernetes.
That also means you can use Pulumi to deploy and manage your FerretDB databases.

[FerretDB](https://www.ferretdb.com/) is an open-source [document database](https://blog.ferretdb.io/5-database-alternatives-mongodb-2023/) that adds MongoDB compatibility to relational databases like [Postgres](https://www.postgresql.org/) and [SQLite](https://www.sqlite.org/).
With Postgres as your backend, you can use Pulumi to [set up FerretDB on top of StackGres](https://blog.ferretdb.io/run-ferretdb-on-stackgres/) — a fully-featured platform for running Postgres on Kubernetes.

In this guide, you'll learn to set up FerretDB with [StackGres](https://stackgres.io/) on Kubernetes using Pulumi.

## Pulumi as your IaC tool

Pulumi empowers DevOps professionals and developers to manage infrastructure using different programming languages.
Let's take a look at some of the key benefits of Pulumi.

- **Language support:** Perhaps, this is the best advantage of Pulumi.
  Rather than learn a new vendor-specific language with its own strict schema and syntax, Pulumi lets you use manage infrastructure with programming languages like Python, JavaScript, Go, Java, and C#, among others.
- **Multi-cloud support:** Pulumi supports multiple cloud providers and SaaS offerings, including AWS, Google Cloud, Microsoft Azure, DigitalOcean, Kubernetes, Docker, ConfluentCloud, and DataDog.
- **Open source:** Pulumi is fully open source under the Apache 2.0 license.
  Being open-source means the community is a huge part of Pulumi.
  They help in building and supporting Pulumi providers, components, and configurations while also creating educational resources about Pulumi.

## Prerequisites

- [minikube](https://minikube.sigs.k8s.io/docs/start/)
- [Pulumi](https://www.pulumi.com/docs/install/)
- [Stackgres](https://stackgres.io/install)
- `mongosh`
- [kubectl](https://kubernetes.io/docs/reference/kubectl/)

## Set up a Kubernetes Cluster

Before deploying FerretDB using Pulumi, you need to have a Kubernetes cluster running, along with command line tools `kubectl`.
If these tools are not installed, please see their respective documentation.

This guide uses minikube to create a cluster.
minikube is a local Kubernetes setup that makes it easy to learn, experiment, and develop for Kubernetes.
Check out the minikube [setup on how to get started and create a cluster](https://minikube.sigs.k8s.io/docs/start/).

Please also ensure that your cluster is set as the current context.

To list all available contexts available to `kubectl`:

```sh
kubectl config get-contexts
```

The current context is marked with an asterisk (\*) in the output.
Use the `kubectl config use-context` command with the name of your desired context to switch the current context to the specified cluster.

```sh
kubectl config use-context <your-desired-context>
```

Start the setup by creating a `ferretdb` namespace within the cluster.
That way, you can isolate, group, and manage your resources, access controls, and other configurations.

```sh
kubectl create namespace ferretdb
```

## Installing the StackGres Operator

Here, you will need to install the StackGres operator on a Kubernetes cluster.
To do that,

Install the operator with the following command:

```sh
kubectl create -f https://stackgres.io/downloads/stackgres-k8s/stackgres/1.7.0/stackgres-operator-demo.yml
```

This will install all the necessary resources, and also add the operator to a new namespace `stackgres`.

It may take some time for the operator to be ready.
If you want to wait until the operator is ready, run the following command:

```sh
kubectl wait -n stackgres deployment -l group=stackgres.io --for=condition=Available
```

The pod status should be set to running once they are ready:

```text
% kubectl get pods -n stackgres
NAME                                  READY   STATUS    RESTARTS        AGE
stackgres-operator-6f7c75bff4-mwchl   1/1     Running   1 (6m38s ago)   8m28s
stackgres-restapi-77c978b5dc-2lzm6    2/2     Running   0               6m17s
```

The next to do is to create a `Secret` in the `ferretdb` namespace.
The `Secret` will contain a random password generated using the SQL command in the following script.

```sh
#!/bin/bash

NAMESPACE="ferretdb"

# Secret name
SECRET_NAME="createuser"

PASSWORD=$(openssl rand -base64 12)

kubectl -n $NAMESPACE create secret generic $SECRET_NAME --from-literal=sql="CREATE USER ferretdb WITH PASSWORD '${PASSWORD}';" --dry-run=client -o yaml | kubectl apply -f -

```

Save the script as `create_secret.sh` or whatever you prefer.
Make this executable by running this in the directory terminal:

```sh
chmod +x ./create_secret.sh
```

Then execute the script and create a unique password.

```sh
./create_secret.sh
```

## Create Pulumi project

Pulumi will be used to orchestrate our entire setup.
With our cluster now available, let's go ahead to set up the Pulumi project.
But before setting up the project, ensure to install Pulumi ([find the installation guide here](https://www.pulumi.com/docs/clouds/kubernetes/get-started/begin/)).

Create a directory for the project.
This will hold the configuration and particular project details for Pulumi.
Since you already have `kubectl` configured, Pulumi will use the same configuration settings.

```sh
mkdir ferretdb-pulumi && cd ferretdb-pulumi
pulumi new kubernetes-python
```

You will be prompted to enter a project name, description, and stack.
Go ahead and press ENTER to confirm the default values, or you can specify your own.

Once this is complete, the project will initialize a new Pulumi project using the Kubernetes Python template and this will generate a couple of files:

- Pulumi.yaml contains the project definition.
- Pulumi.dev.yaml contains the initialized stack configuration values.
- `__main__.py` is the main Pulumi program that defines the resources in your stack.

## Deploy StackGres Cluster with FerretDB on Kubernetes using Pulumi

Here, you'll use Pulumi to provision and orchestrate all the resources.
This will include setting up the connection pooling, creating the Postgres database, deploying the Stackgres, and finally, setting up FerretDB.

Start by deleting the existing content in the `__main__.py` file.
Then add the following to the `__main__.py` file.

```py
import pulumi
from pulumi_kubernetes.core.v1 import Namespace, Service
from pulumi_kubernetes.apps.v1 import Deployment
from pulumi_kubernetes.apiextensions import CustomResource
from pulumi import ResourceOptions


# SGPoolingConfig Creation
sg_pooling_config = CustomResource(
   "sg-pooling-config",
   api_version="stackgres.io/v1",
   kind="SGPoolingConfig",
   metadata={
       "name": "sgpoolingconfig1",
       "namespace": "ferretdb"
   },
   spec={
       "pgBouncer": {
           "pgbouncer.ini": {
               "pgbouncer": {
                   "ignore_startup_parameters": "extra_float_digits,search_path"
               }
           }
       }
   },
)

# SGScript Creation
sg_script = CustomResource(
   "createuserdb",
   api_version="stackgres.io/v1",
   kind="SGScript",
   metadata={
       "name": "createuserdb",
       "namespace": "ferretdb"
   },
   spec={
       "scripts": [
           {
               "name": "create-user",
               "scriptFrom": {
                   "secretKeyRef": {
                       "name": "createuser",
                       "key": "sql"
                   }
               }
           },
           {
               "name": "create-database",
               "script": "create database ferretdb owner ferretdb encoding 'UTF8' locale 'en_US.UTF-8' template template0;"
           }
       ]
   },
   opts=ResourceOptions(depends_on=[sg_pooling_config])
)

# Stackgres Cluster Creation
stackgres_cluster = CustomResource(
   "stackgres-cluster",
   api_version="stackgres.io/v1",
   kind="SGCluster",
   metadata={
       "namespace": "ferretdb",
       "name": "postgres"
   },
   spec={
       "postgres": {
           "version": "15"
       },
       "instances": 1,
       "pods": {
           "persistentVolume": {
               "size": "1Gi"
           }
       },
       "configurations": {
           "sgPoolingConfig": "sgpoolingconfig1"
       },
       "managedSql": {
           "scripts": [
               {"sgScript": "createuserdb"}
           ]
       }
   },
   opts=ResourceOptions(depends_on=[sg_script])
)

# FerretDB Deployment
ferretdb_deployment = Deployment(
   "ferretdb-deployment",
   metadata={
       "name": "ferretdb-deployment",
       "namespace": "ferretdb"
   },
   spec={
       "replicas": 1,
       "selector": {"matchLabels": {"app": "ferretdb"}},
       "template": {
           "metadata": {"labels": {"app": "ferretdb"}},
           "spec": {
               "containers": [{
                   "name": "ferretdb",
                   "image": "ferretdb/ferretdb:latest",
                   # Update this URL with the correct connection string
                   "env": [{
                       "name": "FERRETDB_POSTGRESQL_URL",
                       "value": "postgres://ferretdb:PASSWORD@postgres.ferretdb.svc.cluster.local:5432/ferretdb"
                   }],
                   "ports": [{"containerPort": 27017}]
               }]
           }
       }
   },
   opts=ResourceOptions(depends_on=[stackgres_cluster])
)

# FerretDB Service
ferretdb_service = Service(
   "ferretdb-service",
   metadata={
       "name": "ferretdb-service",
       "namespace": "ferretdb"
   },
   spec={
       "ports": [{"port": 27017, "targetPort": 27017}],
       "selector": {"app": "ferretdb"},
       "type": "ClusterIP"
   },
   opts=ResourceOptions(depends_on=[ferretdb_deployment])
)

pulumi.export('ferretdb_service_ip', ferretdb_service.metadata.apply(lambda meta: meta.name))
```

The resources:

- **SGPoolingConfig:** Defines a CustomResource config for connection pooling, particularly to customize PgBouncer settings.
  Normally, FerretDB will attempt to configure the `search_path` parameter in PostgreSQL during startup.
  However, PgBouncer does not support this.
  So this will customize PgBouncer behavior to ignore this parameter.
- **SGScript:** Creates a StackGres script resource for initializing the user and the `ferretdb` database.
  The user password is sourced from the `Secret` created earlier.
- **Stackgres Cluster:** Deploys a PostgreSQL database cluster managed by StackGres.
  This will help to deploy, scale, and manage Postgres clusters and also apply the custom resource for SGPoolingConfig and SGScript.
- **FerretDB deployment and service:** There are two additional resources.
  One contains the deployment specifications for FerretDB and configures it to connect to a `ferretdb` database created with StackGres.
  It specifies the `FERRETDB_POSTGRESQL_URL` environment variable, which points to the Postgres database defined by the StackGres cluster.
- **FerretDB Service:** This exposes the FerretDB deployment within the cluster on port 27017, making it accessible to other services within the same cluster.

This setup should provide us with a fully managed, scalable FerretDB deployment in Kubernetes environments.

Run `pulumi up` to tie and deploy all the resources.

```text
% pulumi up
Previewing update (dev)

View in Browser (Ctrl+O): https://app.pulumi.com/Fashander/ferretdb-pulumi/dev/previews/4f3bac19-0fa6-4bc6-810e-525883129e3c

     Type                                           Name                 Plan
 +   pulumi:pulumi:Stack                            ferretdb-pulumi-dev  create
 +   ├─ kubernetes:stackgres.io/v1:SGPoolingConfig  sg-pooling-config    create
 +   ├─ kubernetes:stackgres.io/v1:SGScript         createuserdb         create
 +   ├─ kubernetes:stackgres.io/v1:SGCluster        stackgres-cluster    create
 +   ├─ kubernetes:apps/v1:Deployment               ferretdb-deployment  create
 +   └─ kubernetes:core/v1:Service                  ferretdb-service     create

Outputs:
    ferretdb_service_ip: "ferretdb-service"

Resources:
    + 6 to create

Do you want to perform this update? yes
Updating (dev)

View in Browser (Ctrl+O): https://app.pulumi.com/Fashander/ferretdb-pulumi/dev/updates/23

     Type                                           Name                 Status
 +   pulumi:pulumi:Stack                            ferretdb-pulumi-dev  created (42s)
 +   ├─ kubernetes:stackgres.io/v1:SGPoolingConfig  sg-pooling-config    created (0.92s)
 +   ├─ kubernetes:stackgres.io/v1:SGScript         createuserdb         created (0.36s)
 +   ├─ kubernetes:stackgres.io/v1:SGCluster        stackgres-cluster    created (0.86s)
 +   ├─ kubernetes:apps/v1:Deployment               ferretdb-deployment  created (22s)
 +   └─ kubernetes:core/v1:Service                  ferretdb-service     created (10s)

Outputs:
    ferretdb_service_ip: "ferretdb-service"

Resources:
    + 6 created

Duration: 44s
```

The postgres pods might take a few minutes to be ready.
To be sure the pods are running, run `kubectl get pods -n ferretdb`.

```text
% kubectl get pods -n ferretdb
NAME                                   READY   STATUS    RESTARTS   AGE
ferretdb-deployment-69bc74d967-2xss7   1/1     Running   0          2m6s
postgres-0                             6/6     Running   0          2m5s
```

### Connect via mongosh

Let's launch a temporary `mongosh` pod in the `ferretdb` namespace.

```sh
kubectl -n ferretdb run mongosh --image=rtsp/mongosh --rm -it -- bash
```

This should open up a `mongosh` shell in your terminal, and you can use it to connect to FerretDB.

FerretDB exposes the Postgres database created previously with SGScript and allows you to interact with it as if you were using MongoDB.

Connect using the following format:

```sh
mongosh 'mongodb://ferretdb:<password>@<host-address>:27017/ferretdb?authMechanism=PLAIN'
```

You will need the host address for the FerretDB Service and password credential for the Postgres database.

Run `kubectl -n ferretdb get svc` to get the address and port exposed by the FerretDB Service (10.106.153.95:27017 in the example below):

```text
% kubectl -n ferretdb get svc
NAME                TYPE           CLUSTER-IP       EXTERNAL-IP                           PORT(S)             AGE
ferretdb-service    ClusterIP      10.106.153.95    <none>                                27017/TCP           11m
```

You can also access the password by running:

```sh
kubectl -n ferretdb get secret createuser --template '{{ printf "%s\n" (.data.sql | base64decode) }}'
```

Connect to FerretDB:

```text
# mongosh 'mongodb://ferretdb:<password>@<host-address>:27017/ferretdb?authMechanism=PLAIN'
Current Mongosh Log ID: 65d02c968660cd7b3c7ad89e
Connecting to:    mongodb://<credentials>@10.106.153.95:27017/ferretdb?authMechanism=PLAIN&directConnection=true&appName=mongosh+2.1.4
Using MongoDB:    7.0.42
Using Mongosh:    2.1.4
For mongosh info see: https://docs.mongodb.com/mongodb-shell/
------
   The server generated these startup warnings when booting
   2024-02-17T03:48:39.213Z: Powered by FerretDB v1.19.0 and PostgreSQL 15.5.
   2024-02-17T03:48:39.213Z: Please star us on GitHub: https://github.com/FerretDB/FerretDB.
   2024-02-17T03:48:39.213Z: The telemetry state is undecided.
   2024-02-17T03:48:39.213Z: Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.io.
------
ferretdb>
```

Awesome!
With Pulumi, you've been able to run and deploy FerretDB in a Kubernetes cluster.
So you can just go right ahead to run a couple of MongoDB operations.

#### Insert documents

Let's insert documents showing single-day stock data for a fictional company.

```js
db.stocks.insertMany([
  {
    symbol: 'ZTI',
    date: new Date('2024-02-17'),
    tradingData: {
      open: 250.75,
      high: 255.5,
      low: 248.25,
      close: 254.1,
      volume: 1200000
    },
    metadata: {
      analystRating: 'Buy',
      sector: 'Technology'
    }
  },
  {
    symbol: 'ZTI',
    date: new Date('2024-02-18'),
    tradingData: {
      open: 254.1,
      high: 260.0,
      low: 253.0,
      close: 258.45,
      volume: 1500000
    },
    metadata: {
      analystRating: 'Strong Buy',
      sector: 'Technology'
    }
  }
])
```

Run `db.stocks.find()` to see the documents.

#### Query Document

Find stock data indicating where the volume was greater than 1,200,000.

```js
db.stocks.find({ symbol: 'ZTI', 'tradingData.volume': { $gt: 1200000 } })
```

Result:

```json5
[
  {
    _id: ObjectId('65d030d38660cd7b3c7ad8a0'),
    symbol: 'ZTI',
    date: ISODate('2024-02-18T00:00:00.000Z'),
    tradingData: {
      open: 254.1,
      high: 260,
      low: 253,
      close: 258.45,
      volume: 1500000
    },
    metadata: { analystRating: 'Strong Buy', sector: 'Technology' }
  }
]
```

You can go ahead and try more MongoDB commands.

### View data in Postgres

Since FerretDB adds MongoDB compatibility to your Postgres database, you can view the stored data in your Postgres instance.

Connect to Postgres by running this command:

```sh
kubectl -n ferretdb exec -it postgres-0 -c postgres-util -- psql ferretdb
```

Once you're in, set the `search_path` to the `ferretdb` database and then check out the tables in the database.
FerretDB stores the data as JSONB.

```psql
ferretdb=# SET SEARCH_PATH TO ferretdb;
SET
ferretdb=# \dt
                     List of relations
  Schema  |            Name             | Type  |  Owner
----------+-----------------------------+-------+----------
 ferretdb | _ferretdb_database_metadata | table | ferretdb
 ferretdb | stocks_5fb3a312             | table | ferretdb
(2 rows)

ferretdb=# SELECT * from stocks_5fb3a312;
                                                                                                                                                                                                                                                                                                                                                                                                                                                                   _jsonb
---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 {"$s": {"p": {"_id": {"t": "objectId"}, "date": {"t": "date"}, "symbol": {"t": "string"}, "metadata": {"t": "object", "$s": {"p": {"sector": {"t": "string"}, "analystRating": {"t": "string"}}, "$k": ["analystRating", "sector"]}}, "tradingData": {"t": "object", "$s": {"p": {"low": {"t": "int"}, "high": {"t": "int"}, "open": {"t": "double"}, "close": {"t": "double"}, "volume": {"t": "int"}}, "$k": ["open", "high", "low", "close", "volume"]}}}, "$k": ["_id", "symbol", "date", "tradingData", "metadata"]}, "_id": "65d030d38660cd7b3c7ad8a0", "date": 1708214400000, "symbol": "ZTI", "metadata": {"sector": "Technology", "analystRating": "Strong Buy"}, "tradingData": {"low": 253, "high": 260, "open": 254.1, "close": 258.45, "volume": 1500000}}
 {"$s": {"p": {"_id": {"t": "objectId"}, "date": {"t": "date"}, "symbol": {"t": "string"}, "metadata": {"t": "object", "$s": {"p": {"sector": {"t": "string"}, "analystRating": {"t": "string"}}, "$k": ["analystRating", "sector"]}}, "financials": {"t": "object", "$s": {"p": {"dividendPayout": {"t": "double"}}, "$k": ["dividendPayout"]}}, "tradingData": {"t": "object", "$s": {"p": {"low": {"t": "double"}, "high": {"t": "double"}, "open": {"t": "double"}, "close": {"t": "double"}, "volume": {"t": "int"}}, "$k": ["open", "high", "low", "close", "volume"]}}}, "$k": ["_id", "symbol", "date", "tradingData", "metadata", "financials"]}, "_id": "65d030d38660cd7b3c7ad89f", "date": 1708128000000, "symbol": "ZTI", "metadata": {"sector": "Technology", "analystRating": "Buy"}, "financials": {"dividendPayout": 0.5}, "tradingData": {"low": 248.25, "high": 255.5, "open": 250.75, "close": 254.1, "volume": 1200000}}
(2 rows)
```

## Clean up

Pulumi provides a way to clean up and de-provision all the resources created with Pulumi from your project's directory.
Then delete the cluster along with other resources created and managed outside of Pulumi.

```sh
pulumi destroy
minikube delete
```

## Conclusion

So far, you've been able to leverage the power of Pulumi for infrastructure as code, to run, deploy, and manage a scalable FerretDB database in Kubernetes.
Pulumi orchestrates all the resources using Python, deploys a Postgres database using StackGres, and adds MongoDB compatibility to it with FerretDB.
But you can still choose to use any other language of your choice.

FerretDB, due to its open-source nature, offers complete control over your data without any fear of vendor lock-in.

To get started with FerretDB, [check out our quickstart guide](https://docs.ferretdb.io/quickstart-guide/).
