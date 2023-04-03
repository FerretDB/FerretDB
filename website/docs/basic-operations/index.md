---
sidebar_position: 1
---

# Performing CRUD operations

CRUD (Create, Read, Update, and Delete) operations in FerretDB uses the same protocols and drivers as MongoDB.

## Create operations in FerretDB

The create operation adds a new document to a collection.
If the collection does not exist, this operation will create it.
The following methods are available for adding documents to a collection:

[`db.collection.insertOne()`](create.md#insert-a-single-document),
[`db.collection.insertMany()`](create.md#insert-multiple-documents-at-once)

## Read operations in FerretDB

The read operation retrieves document records in a collection.
You can also filter the documents by targeting specific criteria for retrieval.
The following commands are used to retrieve documents from a collection:

[`db.collection.find()`](read.md#retrieve-all-documents-in-a-collection), [`db.collection.findOne()`](read.md#retrieve-a-single-document)

The read operation can also retrieve subdocuments that are nested within a document.

## Update operations in FerretDB

The update operation modifies document records in a collection.
It changes existing documents in a collection according to the query criteria.
The following update operations are supported:

[`db.collection.updateOne()`](update.md#update-a-single-document), [`db.collection.updateMany()`](update.md#update-many-documents), [`db.collection.replaceOne()`](update#replace-a-document)

## Delete operations in FerretDB

The delete operation removes document records from a collection.
The following delete operations are supported:

[`db.collection.deleteOne()`](delete.md#delete-a-single-document), [`db.collection.deleteMany()`](delete.md#deletes-multiple-documents)

Similar to the update operation, this operation retrieves documents matching specific criteria in a collection and deletes them.
