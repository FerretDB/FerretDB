---
sidebar_position: 1
---

# Concepts

Documents are at the heart of every record in FerretDB.
By understanding the structure of documents, you can better store and retrieve data.

In the following sections, we will discuss documents, collections, and dot notation.

## Documents

Documents are self-describing records containing both data types and a description of the data being stored.
They are similar to rows in relational databases.
Here is an example of a single document:

```js
{
    first: "Thomas",
    last: "Edison",
    invention: "Lightbulb",
    birth: 1847
}
```

The above data is stored in a single document.

:::note
FerretDB follows almost the same naming conventions as MongoDB.
However, there are a few restrictions, which you can find [here](migration/diff.md).
:::

For complex documents, you can nest documents inside other documents:

```js
{
  name: {
    first: "Thomas",
    last: "Edison"
  },
  invention: "Lightbulb",
  birth: 1847
}
```

In the example above, the `name` field is a subdocument embedded into a document.

## Dot notation

Dot notations `(.)` are used to reference a field in an embedded document or its index position in an array.

### Arrays

Dot notations can be used to specify or query an array by concatenating a dot `(.)` with the index position of the field.

```js
'array_name.index'
```

:::note
When using dot notations, the field name of the array and the specified value must be enclosed in quotation marks.
:::

For example, let's take the following array field in a document:

```js
animals: ['dog', 'cat', 'fish', 'fox']
```

To reference the fourth field in the array, use the dot notation `"animals.3"`.

Here are more examples of dot notations on arrays:

- [Query an array](read.md#retrieve-documents-containing-a-specific-value-in-an-array)
- [Update an array](update.md#update-an-array-element)

### Embedded documents

To reference or query a field in an embedded document, concatenate the name of the embedded document and the field name using the dot notation.

```js
'embedded_document_name.field'
```

Take the following document, for example:

```js
{
   name:{
      first: "Tom",
      last: "Barry"
   },
   contact:{
      address:{
         city: "Kent",
         state: "Ohio"
      },
      phone: "432-124-1234"
   }
}
```

To reference the `city` field in the embedded document, use the dot notation `"contact.address.city"`.

For dot notation examples on embedded documents, see here:

- [Query an embedded document](read.md#query-on-an-embedded-or-nested-document)
- [Update an embedded document](update.md#update-an-embedded-document)

## Collections

Collections are a repository for documents.
To some extent, they are similar to tables in a relational database.
If a collection does not exist, FerretDB creates a new one when you insert documents for the first time.
A collection may contain one or more documents.
For example, the following collection contains three documents.

```js
{
  Scientists: [
    {
      first: 'Alan',
      last: 'Turing',
      born: 1912
    },
    {
      first: 'Thomas',
      last: 'Edison',
      birth: 1847
    },
    {
      first: 'Nikola',
      last: 'Tesla',
      birth: 1856
    }
  ]
}
```
