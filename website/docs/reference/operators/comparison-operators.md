---
sidebar_position: 1
---

# Comparison query operators

Comparison query operators allow you to compare the elements in a document to a given query value.

Go to the comparison query operators:

| Operator       | Description                                                               |
| -------------- | ------------------------------------------------------------------------- |
| [`$eq`](#eq)   | Selects documents with elements that are equal to a given query value     |
| [`$gt`](#gt)   | Selects documents with elements that are greater than a given query value |
| [`$gte`](#gte) | Selects documents greater than or equal to specified query                |
| [`$lt`](#lt)   | Selects documents with elements that are less than a given query value    |
| [`$lte`](#lte) | Selects documents with elements less than or equal to given query value   |
| [`$in`](#in)   | Selects documents that contain the elements in a given array query        |
| [`$ne`](#ne)   | Selects documents with elements that are not equal to given query value   |
| [`$nin`](#nin) | Selects documents that do not contain the elements in a given array query |

For the examples in this section, insert the following documents into the `employees` collection:

```js
db.employees.insertMany([
  {
     name: {
        first: "Earl",
        last: "Thomas"
     },
     employeeID: 1234,
     age: 23,
     role: "salesperson",
     catalog: [
        "printer",
        "cardboard",
        "crayons",
        "books"
     ]
  },
  {
     name: {
        first: "Sam",
        last: "Johnson"
     },
     employeeID: 2234,
     age: 35,
     role: "salesperson",
     catalog: [
        "cabinet",
        "fridge",
        "blender",
        "utensils"
     ]
  },
  {
     name: {
        first: "Clarke",
        last: "Dane"
     },
     employeeID: 3234,
     age: 21,
     role: "salesperson",
     catalog: [
        "printer",
        "pencils",
        "crayons",
        "toys"
     ]
  }
])
```

## $eq

*Syntax*: `{ field: { $eq: <query-value> } }`

The `$eq` operator selects the documents containing the element that is equal to the given query value.
This operator can be used to match values of different types, including documents, array, embedded documents, etc.

**Example:** The following operation queries the `employees` collection for all documents where the field `age` equals `21`.

```js
db.employees.find({
   age: {
      $eq: 21
   }
})
```

The response returns a single document that matches the query:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0e"),
    name: { first: 'Clarke', last: 'Dane' },
    employeeID: 3234,
    age: 21,
    role: 'salesperson',
    catalog: [ 'printer', 'pencils', 'crayons', 'toys' ]
  }
]
```

**Example:** To query values in an embedded document, use [dot notation](../../understanding_ferretdb.md#dot-notation).
The following operation queries the `employees` collection for documents that match the field `first` in the embedded document `name`.

```js
db.employees.find({
   "name.first":{
      $eq: "Earl"
   }
})
```

The response returns a single document that matches the query:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0c"),
    name: { first: 'Earl', last: 'Thomas' },
    employeeID: 1234,
    age: 23,
    role: 'salesperson',
    catalog: [ 'printer', 'cardboard', 'crayons', 'books' ]
  }
]
```

## $gt

*Syntax*: `{ field: { $gt: <query-value> } }`

The greater than operator `$gt` selects documents containing elements that are greater than the given query value.

**Example:** Use the following operation to query for all the documents in the `employees` collection where the field `age` is greater than `21`.

```js
db.employees.find({
   age: {
      $gt: 21
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0c"),
    name: { first: 'Earl', last: 'Thomas' },
    employeeID: 1234,
    age: 23,
    role: 'salesperson',
    catalog: [ 'printer', 'cardboard', 'crayons', 'books' ]
  },
  {
    _id: ObjectId("639a3cce071b6bed396a8f0d"),
    name: { first: 'Sam', last: 'Johnson' },
    employeeID: 2234,
    age: 35,
    role: 'salesperson',
    catalog: [ 'cabinet', 'fridge', 'blender', 'utensils' ]
  }
]
```

## $gte

*Syntax*: `{ field: { $gte: <query-value> } }`

The greater than or equal operator `$gte` selects document with values greater than or equal to the given query value.

**Example:** The following operation selects documents based on the specified query, where the field `age` is greater than or equal to `21`.

```js
db.employees.find({
   age: {
      $gte: 21
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0c"),
    name: { first: 'Earl', last: 'Thomas' },
    employeeID: 1234,
    age: 23,
    role: 'salesperson',
    catalog: [ 'printer', 'cardboard', 'crayons', 'books' ]
  },
  {
    _id: ObjectId("639a3cce071b6bed396a8f0d"),
    name: { first: 'Sam', last: 'Johnson' },
    employeeID: 2234,
    age: 35,
    role: 'salesperson',
    catalog: [ 'cabinet', 'fridge', 'blender', 'utensils' ]
  },
  {
    _id: ObjectId("639a3cce071b6bed396a8f0e"),
    name: { first: 'Clarke', last: 'Dane' },
    employeeID: 3234,
    age: 21,
    role: 'salesperson',
    catalog: [ 'printer', 'pencils', 'crayons', 'toys' ]
  }
]
```

## $lt

*Syntax*: `{ field: { $lt: <query-value> } }`

The less than operator `$lt` selects document with elements that are less than the given query value.

**Example:** The following operation queries for documents where the field `age` is less than `25`.

```js
db.employees.find({
   age: {
      $lt: 25
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0c"),
    name: { first: 'Earl', last: 'Thomas' },
    employeeID: 1234,
    age: 23,
    role: 'salesperson',
    catalog: [ 'printer', 'cardboard', 'crayons', 'books' ]
  },
  {
    _id: ObjectId("639a3cce071b6bed396a8f0e"),
    name: { first: 'Clarke', last: 'Dane' },
    employeeID: 3234,
    age: 21,
    role: 'salesperson',
    catalog: [ 'printer', 'pencils', 'crayons', 'toys' ]
  }
]
```

## $lte

*Syntax*: `{ field: { $lte: <query-value> } }`

The less than or equal operator `$lte` selects documents with elements that are less than or equal to the specified query value.
The following operation queries for documents where the field `age` is less than or equal to `21`.

```js
db.employees.find({
   age: {
      $lte: 21
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0e"),
    name: { first: 'Clarke', last: 'Dane' },
    employeeID: 3234,
    age: 21,
    role: 'salesperson',
    catalog: [ 'printer', 'pencils', 'crayons', 'toys' ]
  }
]
```

## $in

*Syntax*: `{ field: { $in: [<array-value1>, <array-value2>, ... <array-valueN> ] } }`

The `$in` operator selects documents that contain any of the values in a given query array.
To match a document, the field must contain at least one of the elements in the specified array query.

**Example:** The following operation queries the `employees` collection for documents where the value of the field `age` is either `21` or `35`.

```js
db.employees.find({
   age: {
      $in: [ 21, 35 ]
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0d"),
    name: { first: 'Sam', last: 'Johnson' },
    employeeID: 2234,
    age: 35,
    role: 'salesperson',
    catalog: [ 'cabinet', 'fridge', 'blender', 'utensils' ]
  },
  {
    _id: ObjectId("639a3cce071b6bed396a8f0e"),
    name: { first: 'Clarke', last: 'Dane' },
    employeeID: 3234,
    age: 21,
    role: 'salesperson',
    catalog: [ 'printer', 'pencils', 'crayons', 'toys' ]
  }
]
```

## $ne

*Syntax*: `{ field: { $ne: <query-value> } }`

The `$ne` operator selects all the documents with elements that are not equal to a given query.

**Example:** The following operation queries the `employees` collection for documents where the field `age` is not equal to `21`.

```js
db.employees.find({
   age: {
      $ne: 21
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0c"),
    name: { first: 'Earl', last: 'Thomas' },
    employeeID: 1234,
    age: 23,
    role: 'salesperson',
    catalog: [ 'printer', 'cardboard', 'crayons', 'books' ]
  },
  {
    _id: ObjectId("639a3cce071b6bed396a8f0d"),
    name: { first: 'Sam', last: 'Johnson' },
    employeeID: 2234,
    age: 35,
    role: 'salesperson',
    catalog: [ 'cabinet', 'fridge', 'blender', 'utensils' ]
  }
]
```

## $nin

*Syntax*: `{ field: { $nin: [ <array-value1>, <array-value2> ... <array-valueN> ] } }`

The `$nin` operator selects documents that does not contain any of the values in a given array.

**Example:** The following operation queries the `employees` collection for documents where the value of the field `age` is not `21` or `35`.

```js
db.employees.find({
   age: {
      $nin: [ 21, 35 ]
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("639a3cce071b6bed396a8f0c"),
    name: { first: 'Earl', last: 'Thomas' },
    employeeID: 1234,
    age: 23,
    role: 'salesperson',
    catalog: [ 'printer', 'cardboard', 'crayons', 'books' ]
  }
]
