---
slug: run-ferretdb-on-stackgres
title: 'How to Run FerretDB on Top of StackGres'
authors:
  - alex
  - name: Álvaro Hernández
    title: Founder and CEO @ OnGres
    url: https://www.linkedin.com/in/ahachete
    image_url: https://stackgres.io/img/team/alvaro.jpg
image: /img/blog/stackgres-ferretdb.png
description: >
  Learn to set up and run FerretDB – MongoDB open-source alternative on your Kubernetes cluster – and easily deploy and manage PostgreSQL instances using StackgGres operator.
tags: [compatible applications, open source, tutorial]
---

![How to Run FerretDB on Top of StackGres](/img/blog/stackgres-ferretdb.png)

In this how-to guide, we'll walk you through the whole process of setting up [FerretDB](https://www.ferretdb.io/) on a Kubernetes cluster using [StackGres](https://stackgres.io/).

<!--truncate-->

As an open-source MongoDB alternative, FerretDB translates MongoDB wire protocols to SQL using a [PostgreSQL](https://www.postgresql.org/) backend.
And by using the StackGres operator, you can easily deploy and manage PostgreSQL instances, as well as access all the management features you need.

So if you're exploring a scalable and open-source MongoDB alternative on your Kubernetes cluster, then this article on FerretDB and StackGres should pique your interest.

## OnGres (StackGres)

[StackGres](https://stackgres.io) is the full-stack Postgres Platform.
A fully open source software to run your own Postgres-as-a-Service on any cloud or on-prem.
StackGres is a project from [OnGres](https://ongres.com), the Postgres laser-focused startup ("OnGres" means "ON postGRES").

StackGres is a Kubernetes Operator for Postgres.
It allows you to create production-ready Postgres clusters in seconds.
No advanced Postgres expertise required.
You can use the built-in Web Console or the high level Kubernetes CRDs for CLI and GitOps.

With StackGres, writing a simple YAML manifest (or point-and-click on the Web Console) is all that is needed to create production-ready Postgres clusters including high availability with Patroni, replication, connection pooling, automated backups, monitoring and centralized logs.

With support for [more than 150 Postgres extensions](https://stackgres.io/extensions), StackGres is the most extensible Postgres platform available.
It also provides support for [Babelfish for Postgres](https://babelfishpg.org/) (which brings SQL Server compatibility, at the wire protocol level, SQL and T-SQL) and integrations with Citus for sharding Postgres, Timescale for time-series, Supabase and now, FerretDB.

## FerretDB

FerretDB is the defacto open-source replacement for MongoDB with the popular and reliable PostgreSQL as the database backend.
What FerretDB does is to [convert MongoDB wire protocols in BSON format into JSONB in PostgreSQL](https://blog.ferretdb.io/pjson-how-to-store-bson-in-jsonb/).

Despite MongoDB's popularity as an open-source database and its popularity among developers, after the switch from open source license to SSPL ([get the full story on that here!](https://blog.ferretdb.io/open-source-is-in-danger/)), it was important to restore MongoDB workloads back to open-source so users can have complete control of their data without vendor lock-in.

Built on an ever-reliable PostgreSQL database backend, you can host FerretDB anywhere or run it locally on your own machine.
Another significant advantage of FerretDB is that you get to use the same syntax and commands as you're used to in MongoDB.
Besides, you can also query it with SQL in PostgreSQL (in some cases, you may also need intricate knowledge of JSONB for more advanced queries).

At present, FerretDB is compatible with the most common MongoDB use cases and plans on improving and adding more features as needs arise.
Plus, we've just recently released [FerretDB version 1.5.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.5.0), which includes beta support for SQLite backend.

## Setting up FerretDB on top of StackGres

Before installing StackGres, you will need a running Kubernetes cluster and the usual command line tools [`kubectl`](https://kubernetes.io/docs/tasks/tools/) and [`Helm`](https://helm.sh/docs/intro/install/).
Please refer to the respective installation pages if you don't have these tools.
As for Kubernetes, if you don't have one you can try easily with [K3s](https://k3s.io/).
It can be installed with a single command line as in:

```sh
curl -sfL https://get.k3s.io | sh -
```

This should give you a running single-node cluster in seconds (depending on your Internet connection speed).

Keep in mind that K3s is not available natively on macOS.
If you want to run it on macOS, you'll have to use a virtual machine or a Docker container running Linux.

First, you need to install Docker Desktop for Mac from the Docker website.
After installing Docker Desktop, go to preferences and navigate to the Kubernetes tab to "Enable Kubernetes", then click "Apply & Restart".

Fortunately, K3D is a tool that makes it easy to run K3s inside Docker, which works on macOS.
You can install K3D using Homebrew and if successful, create a Kubernetes cluster:

```sh
brew install k3d && k3d cluster create <cluster-name> --image rancher/k3s:v1.25.9-k3s1
```

Then you can get the configuration so that you can connect and operate with it via both `kubectl` and `Helm`:

```sh
mkdir -p ~/.kube/; sudo k3s kubectl config view --raw >> ~/.kube/config
```

For macOS users, using the command above might create some issues so it's better to use this one below:

```sh
k3d kubeconfig get <cluster-name> > ~/.kube/config
```

Once you are done you can uninstall k3s if you wish with `sudo
/usr/local/bin/k3s-uninstall.sh`.

### Installing StackGres

The best way to install StackGres is through the official Helm chart.
Here's the installation guide in the official docs.

For our particular setup, we use the following Helm commands:

```sh
helm repo add stackgres-charts https://stackgres.io/downloads/stackgres-k8s/stackgres/helm/
helm install --create-namespace --namespace stackgres stackgres-operator stackgres-charts/stackgres-operator
```

To confirm that the operator is running while also waiting for setup to complete, run the following commands:

```sh
kubectl wait -n stackgres deployment -l group=stackgres.io --for=condition=Available
kubectl get pods -n stackgres -l group=stackgres.io
```

As you run the first `kubectl` command, it should wait for the successful deployment, and the second command will list the pods running in the `stackgres` namespace.

```text
NAME                                 READY   STATUS    RESTARTS   AGE
stackgres-operator-c4c6b4bcd-trsgp   1/1     Running   0          4m50s
stackgres-restapi-6986cc8997-lfwql   2/2     Running   0          4m49s
```

### Creating a StackGres Cluster

Here, we'll create an SGCluster configured to fit FerretDB's requirements.
Resources for the example are available in the [apps-on-stackgres GitHub repository](https://github.com/ongres/apps-on-stackgres/tree/main/examples/ferretdb).

Please clone the repository and navigate to the examples/ferretdb directory, which contains all the files referenced in this guide.

To properly group all related resources together, let's first create a namespace:

```yaml
kind: Namespace
apiVersion: v1
metadata:
  name: ferretdb
```

From the apps-on-stackgres GitHub repository on your local machine, navigate to the `examples/ferretdb` folder within your terminal and apply the configurations specified in the `01-namespace.yaml` file to your Kubernetes cluster using this command:

```sh
kubectl apply -f 01-namespace.yaml
```

During startup, FerretDB will try to configure the `search_path` parameter in PostgreSQL.
However, PgBouncer, the Postgres sidecar deployed by StackGres by default, does not support this.
You can either choose to disable PgBouncer's sidecar (not recommended) or customize PgBouncer's connection pooling configuration so it ignores this parameter:

```yaml
apiVersion: stackgres.io/v1
kind: SGPoolingConfig
metadata:
  name: sgpoolingconfig1
  namespace: ferretdb
spec:
  pgBouncer:
    pgbouncer.ini:
      pgbouncer:
        ignore_startup_parameters: extra_float_digits,search_path
```

Create with:

```sh
kubectl apply -f 02-sgpoolingconfig.yaml
```

FerretDB uses one or more Postgres databases, and requires them to be created and owned by a given user.
In this guide, we will create one database, with one user and a unique password, but we will not be using a Potgres superuser for this.

We plan to use the StackGres' [SGScript](https://stackgres.io/doc/latest/reference/crd/sgscript/) facility to create, manage and apply SQL scripts in the database automatically.

First, we'll start by creating a `Secret` containing the SQL command that will generate a random password for the user.

```sh
#!/bin/sh

PASSWORD="$(dd if=/dev/urandom bs=1 count=8 status=none | base64 | tr / 0)"

kubectl -n ferretdb create secret generic createuser \
  --from-literal=sql="create user ferretdb with password '"${PASSWORD}"'"
```

```sh
./03-createuser_secret.sh
```

The next step is to create SGScript, which contains two scripts: a script for creating a password-protected user by obtaining the SQL literal from this `Secret`, and another one to create the database, owned by this user, and configured to FerretDB's requirements:

```yaml
apiVersion: stackgres.io/v1
kind: SGScript
metadata:
  name: createuserdb
  namespace: ferretdb
spec:
  scripts:
    - name: create-user
      scriptFrom:
        secretKeyRef:
          name: createuser
          key: sql
    - name: create-database
      script: |
        create database ferretdb owner ferretdb encoding 'UTF8' locale 'en_US.UTF-8' template template0;
```

```sh
kubectl apply -f 04-sgscript.yaml
```

We are now ready to create the Postgres cluster:

```yaml
apiVersion: stackgres.io/v1
kind: SGCluster
metadata:
  namespace: ferretdb
  name: postgres
spec:
  postgres:
    version: '15'
  instances: 1
  pods:
    persistentVolume:
      size: '5Gi'
  configurations:
    sgPoolingConfig: sgpoolingconfig1
  managedSql:
    scripts:
      - sgScript: createuserdb
```

```sh
kubectl apply -f 05-sgcluster.yaml
```

It should take a few seconds to a few minutes for the cluster to be up and running:

```text
kubectl -n ferretdb get pods
NAME                           READY   STATUS    RESTARTS   AGE
postgres-0                     6/6     Running   0          16m
```

Likewise, a database named `ferretdb` must exist and be owned by the same user:

```text
kubectl -n ferretdb exec -it postgres-0 -c postgres-util -- psql -l ferretdb
                                                 List of databases
   Name    |  Owner   | Encoding |   Collate   |    Ctype    | ICU Locale | Locale Provider |   Access privileges
-----------+----------+----------+-------------+-------------+------------+-----------------+-----------------------
 ferretdb  | ferretdb | UTF8     | en_US.UTF-8 | en_US.UTF-8 |            | libc            |
  ...
```

### Deploying FerretDB

FerretDB itself is a stateless application, and as such, so we can just use the standard `Deployment` pattern (for easy scaling) with a `Service` to deploy it:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ferretdb-dep
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
            - name: FERRETDB_POSTGRESQL_URL
              value: postgres://postgres/ferretdb

---
apiVersion: v1
kind: Service
metadata:
  name: ferretdb
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

```sh
kubectl apply -f 06-ferretdb.yaml
```

Note the line where we pass the `FERRETDB_POSTGRESQL_URL` environment variable to FerretDB's container, set with the value `postgres://postgres/ferretdb`: the second `postgres` on the string is the `Service` name that StackGres exposes pointing to the primary instance of the created cluster, which is named after the `SGCluster`'s name; and `ferretdb` is the name of the database.

If all goes well, you should see the pod up & running.
To test it, we need to run a MongoDB client.
For example, we can use `kubectl run` to run a `mongosh` image:

```sh
kubectl -n ferretdb run mongosh --image=rtsp/mongosh --rm -it -- bash
```

FerretDB exposes the same Postgres database, usernames, and passwords over the MongoDB wire protocol.
In that case, we have access to the user, password, and database that were created previously using `SGScript`.

When the terminal prompt appears, type the command:

```sh
mongosh mongodb://ferretdb:${PASSWORD}@${FERRETDB_SVC}/ferretdb?authMechanism=PLAIN
```

where `${PASSWORD}` represents the randomly generated password for the `ferretdb` user in Postgres:

```sh
kubectl -n ferretdb get secret createuser --template '{{ printf "%s\n" (.data.sql | base64decode) }}'
```

and `${FERRETDB_SVC}` is the address and port exposed by the FerretDB `Service` (`10.43.94.52:27017` in the example below):

```text
kubectl -n ferretdb get svc ferretdb
NAME       TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)     AGE
ferretdb   ClusterIP   10.43.94.52   <none>        27017/TCP   10m
```

### Quickstart Example

Once the mongosh command is executed, you can try inserting and querying data:

```text
ferretdb> db.test.insertOne({a:1})
{
  acknowledged: true,
  insertedId: ObjectId("646dca174663396264d4bfeb")
}
ferretdb> db.test.find()
[ { _id: ObjectId("646dca174663396264d4bfeb"), a: 1 } ]
```

If you are curious, you can see how data was materialized on the Postgres database:

```text
kubectl -n ferretdb exec -it postgres-0 -c postgres-util -- psql ferretdb
psql (15.1 (OnGres 15.1-build-6.18))
Type "help" for help.

ferretdb=# set search_path to ferretdb;
SET
ferretdb=# \dt
                     List of relations
  Schema  |            Name             | Type  |  Owner
----------+-----------------------------+-------+----------
 ferretdb | _ferretdb_database_metadata | table | ferretdb
 ferretdb | test_afd071e5               | table | ferretdb
(2 rows)

ferretdb=# table test_afd071e5;
                                                           _jsonb
-----------------------------------------------------------------------------------------------------------------------------
 {"a": 1, "$s": {"p": {"a": {"t": "int"}, "_id": {"t": "objectId"}}, "$k": ["_id", "a"]}, "_id": "646dca174663396264d4bfeb"}
(1 row)
```

## Creating a Meteor App with FerretDB on top of Stackgres

To test and run our entire setup locally, we can use port forwarding with `kubectl` command, which should forward all connections from our local machine port to the `ferretdb` service's port `27017` in the cluster.

```sh
kubectl port-forward svc/ferretdb 27017:27017 -n ferretdb
```

In a new terminal, run this command to be sure your connection is working (note that this requires that you have mongosh installed.)

```sh
mongosh 'mongodb://ferretdb:${PASSWORD}@localhost:27017/ferretdb?authMechanism=PLAIN'
```

To test our database locally, we'll be executing commands from a Meteor application.
If you don't have Meteor installed, you can install it by following the steps in their documentation.
Then create a project using:

```sh
meteor create <app-name>
cd <app-name>
meteor npm install
```

Be sure to replace `<app-name>` with that of your project.

Without making any changes, on startup, the function will insert documents with the fields `title`, `url`, `createdAt` as long as there are none present.

```js
async function insertLink({ title, url }) {
  await LinksCollection.insertAsync({ title, url, createdAt: new Date() })
}

Meteor.startup(async () => {
  // If the Links collection is empty, add some data.
  if ((await LinksCollection.find().countAsync()) === 0) {
    await insertLink({
      title: 'Do the Tutorial',
      url: 'https://www.meteor.com/tutorials/react/creating-an-app'
    })

    await insertLink({
      title: 'Follow the Guide',
      url: 'https://guide.meteor.com'
    })

    await insertLink({
      title: 'Read the Docs',
      url: 'https://docs.meteor.com'
    })

    await insertLink({
      title: 'Discussions',
      url: 'https://forums.meteor.com'
    })
  }

  // We publish the entire Links collection to all clients.
  // In order to be fetched in real-time to the clients
  Meteor.publish('links', function () {
    return LinksCollection.find()
  })
})
```

Once installed, you can run it locally by connecting our MongoDB URI with Meteor, which then sets the environment variable for the MongoDB connection string.
This way, Meteor will use an external MongoDB database instead of starting its own.

```sh
MONGO_URL='mongodb://username:password@localhost:27017/mydatabase' meteor
```

You can check out these documents in the Postgres database.

```text
ferretdb-# \dt
                     List of relations
  Schema  |            Name             | Type  |  Owner
----------+-----------------------------+-------+----------
 ferretdb | _ferretdb_database_metadata | table | ferretdb
 ferretdb | links_e9ca9aee              | table | ferretdb
 ferretdb | test_afd071e5               | table | ferretdb
(3 rows)
ferretdb=# table links_e9ca9aee;
                                                                                                                                                         _jsonb
------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 {"$s": {"p": {"_id": {"t": "string"}, "url": {"t": "string"}, "title": {"t": "string"}, "createdAt": {"t": "date"}}, "$k": ["_id", "title", "url", "createdAt"]}, "_id": "7NmpEw6k7z7HNTJpf", "url": "https://www.meteor.com/tutorials/react/creating-an-app", "title": "Do the Tutorial", "createdAt": 1687291371079}
 {"$s": {"p": {"_id": {"t": "string"}, "url": {"t": "string"}, "title": {"t": "string"}, "createdAt": {"t": "date"}}, "$k": ["_id", "title", "url", "createdAt"]}, "_id": "f3D3wB25KGduKzbsc", "url": "https://guide.meteor.com", "title": "Follow the Guide", "createdAt": 1687291371148}
 {"$s": {"p": {"_id": {"t": "string"}, "url": {"t": "string"}, "title": {"t": "string"}, "createdAt": {"t": "date"}}, "$k": ["_id", "title", "url", "createdAt"]}, "_id": "rdJH59NGgnonSSe5o", "url": "https://docs.meteor.com", "title": "Read the Docs", "createdAt": 1687291371159}
 {"$s": {"p": {"_id": {"t": "string"}, "url": {"t": "string"}, "title": {"t": "string"}, "createdAt": {"t": "date"}}, "$k": ["_id", "title", "url", "createdAt"]}, "_id": "zoGgwqbofdGDvPyj4", "url": "https://forums.meteor.com", "title": "Discussions", "createdAt": 1687291371164}
(4 rows)
```

## Conclusion

With FerretDB, you can seamlessly store and access your data, enabling the flexibility and freedom of using MongoDB's document model and syntax without the vendor lock-in.

Throughout this article, we have explored the process of setting up a Kubernetes cluster, deploying the StackGres operator, and running FerretDB on top of it.
We have covered steps such as creating an SGCluster, and performing basic database operations in FerretDB.

If you're eager to try FerretDB and experience its capabilities firsthand, we encourage you to try it out in your own Kubernetes environment, through the StackGres operator.

Here's the [Stackgres runbook](https://stackgres.io/doc/latest/runbooks/ferretdb-stackgres/) and [FerretDB installation guide](https://docs.ferretdb.io/quickstart-guide/) to get you started.

If you encounter any issue or just wish to say "Hi!" and tell us how FerretDB works on StackGres, you may drop a line at the:

- [FerretDB Community channels](https://docs.ferretdb.io/#community)
- [StackGres Community Slack](https://slack.stackgres.io)
