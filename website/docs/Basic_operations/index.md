---
sidebar_position: 1
---



# Performing CRUD operations

CRUD (Create, Read, Update, and Delete) operations in FerretDB uses the same protocols and drivers as MongoDB. 


## Create operations in FerretDB

The create operation adds a new document to a collection. If the collection does not exist, this operation will create it. The following methods are available for adding documents to a collection.

```sh
db.collection.insertOne()
db.collection.insertMany()
```


## Read operations in FerretDB

The read operation retrieves document records in a collection. You can also filter the documents by targeting specific criteria for retrieval. The following commands are used to retrieve documents from a collection:

```sh
db.collection.find()
db.collection.findOne()
```

The read operation can also retrieve subdocuments that are nested within a document.


## Update operations in FerretDB

The update operation modifies document records in a collection. It changes existing documents in a collection according to the query criteria. The following update operations are supported:

```sh
db.collection.updateOne()
db.collection.updateMany()
db.collection.replaceOne()
```


## Delete operations in FerretDB

The delete operation removes document records from a collection. The following delete operations are supported:

```sh
db.collection.deleteOne()
db.collection.deleteMany()
```

Similar to the update operation, this operation retrieves documents matching specific criteria in a collection and deletes them.
