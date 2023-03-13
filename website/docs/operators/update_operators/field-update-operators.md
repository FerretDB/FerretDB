---
sidebar_position: 1
---

# Field update operators

Field update operators allow you to modify the value of a specified field in a document when certain conditions are met.

| Operator                       | Description                                                                                 |
| ------------------------------ | ------------------------------------------------------------------------------------------- |
| [`$set`](#set)                 | Assigns the value of a given field                                                        |
| [`$unset`](#unset)             | Deletes the records of a field from a document                                              |
| [`$inc`](#inc)                 | Increments a given field's value                                                            |
| [`$mul`](#mul)                 | Multiplies a given field’s value by a specific value                                        |
| [`$rename`](#rename)           | Renames a given field with another name                                                     |
| [`$min`](#min)                 | Updates a particular field only when the specified value is lesser than the specified value |
| [`$max`](#max)                 | Updates a particular field only when the specified value is higher than the specified value |
| [`$currentDate`](#currentdate) | Specifies the current date and time as the value of a given field                           |
| `$setOnInsert`                 | Inserts elements into an array only if they don't already exist                             |

For the examples in this section, insert the following documents into the `employees` collection:

```js
db.employee.insertOne({
    name: "John Doe",
    age: 35,
    email: "johndoe@example.com",
    phone: "123-456-7890",
    address: {
        street: "123 Main St",
        city: "Anytown",
        state: "CA",
    },
    salary: 50000,
    jobTitle: "Manager",
    startDate: new Date("2021-01-01"),
    endDate: null
})
```

## $set

The `$set` operator updates the value of a specified field and if the field does not exist, the `$set` operator creates a new field and adds it to the document.

The below query is an example that updates the value of the `city` field in the `address` embedded document.

```js
db.employee.updateOne(
    { name: "John Doe" },
    { $set: {
        "address.city": "New York",
        "address.zip": "12345",
        } }
)
```

The above query updates the value of the `city` field in the `address` embedded document and adds a new field `zip` to it.

This is the updated document:

```js
[
  {
    _id: ObjectId("640a603558955e0e2b57c00d"),
    name: 'John Doe',
    age: 35,
    email: 'johndoe@example.com',
    phone: '123-456-7890',
    address: {
      street: '123 Main St',
      city: 'New York',
      state: 'CA',
      zip: '12345',
    },
    salary: 50000,
    jobTitle: 'Manager',
    startDate: ISODate("2021-01-01T00:00:00.000Z"),
    endDate: null
  }
]
```

## $unset

The `$unset` operator deletes the specified field from a document and if the field is not present, the `$unset` operator will not do anything.

The below query deletes the `zip` field from the embedded document `address`.

```js
db.employee.updateOne(
    { name: "John Doe" },
    { $unset: { "address.zip": "" } }
)
```

Below is the updated document showing the updated document without the pox :

```js
[
  {
    _id: ObjectId("640a603558955e0e2b57c00d"),
    name: 'John Doe',
    age: 35,
    email: 'johndoe@example.com',
    phone: '123-456-7890',
    address: {
      street: '123 Main St',
      city: 'New York',
      state: 'CA',
    },
    salary: 50000,
    jobTitle: 'Manager',
    startDate: ISODate("2021-01-01T00:00:00.000Z"),
    endDate: null
  }
]
```

## $inc

The `$inc` operator increments the value of a given field by a specified amount.
If the field is non-existent in the document, the `$inc` operator creates a new field and adds it to the document, setting the value to the specified increment amount.

The below query increments the value of the `age` field by `1`.

```js
db.employee.updateOne(
    { name: "John Doe" },
    { $inc: { age: 1 } }
)
```

The updated document looks like this:

```js
[
  {
    _id: ObjectId("640a603558955e0e2b57c00d"),
    name: 'John Doe',
    age: 36,
    email: 'johndoe@example.com',
    phone: '123-456-7890',
    address: {
        street: '123 Main St',
        city: 'New York',
        state: 'CA'
        },
    salary: 50000,
    jobTitle: 'Manager',
    startDate: ISODate("2021-01-01T00:00:00.000Z"),
    endDate: null
  }
]
```

## $mul

The `$mul` operator multiplies the value of a given field by a specified amount.
Similar to all most of the other field update operators, if the field is non-existent in the document, the `$mul` operator creates a new one and sets the value to `0`.

The below query multiplies the value of the `salary` field by `25%`, represented as `1.25`.

```js
db.employee.updateOne(
    { name: "John Doe" },
    { $mul: { salary: 1.25 } }
)
```

The updated record looks like this:

```js
[
  {
    _id: ObjectId("640a603558955e0e2b57c00d"),
    name: 'John Doe',
    age: 36,
    email: 'johndoe@example.com',
    phone: '123-456-7890',
    address: {
        street: '123 Main St',
        city: 'New York',
        state: 'CA'
        },
    salary: 62500,
    jobTitle: 'Manager',
    startDate: ISODate("2021-01-01T00:00:00.000Z"),
    endDate: null
  }
]
```

## $rename

The `$rename` operator renames a given field to another name.

The query below updates the `employee` collection and renames the `jobTitle` field to `title`.

```js
db.employee.updateOne(
    { name: "John Doe" },
    { $rename: {jobTitle: "title" } }
)
```

The updated document looks like this:

```js
[
  {
    _id: ObjectId("640a603558955e0e2b57c00d"),
    name: 'John Doe',
    age: 36,
    email: 'johndoe@example.com',
    phone: '123-456-7890',
    address: {
        street: '123 Main St',
        city: 'New York',
        state: 'CA'
        },
    salary: 62500,
    startDate: ISODate("2021-01-01T00:00:00.000Z"),
    endDate: null,
    title: 'Manager'
  }
]
```

## $min

The `$min` operator compares a specified value with the value of the given field and updates the field to the specified value if the specified value is less than the current value of the field.

The below query updates the value of the `age` field to `30` as long as the current value is less than `30`.

```js
db.employee.updateOne(
    { name: "John Doe" },
    { $min: { age: 30 } }
)
```

Since `30` is less than `36`, the value of the `age` field is updated to `30`.
The updated document now looks like this:

```js
[
  {
    _id: ObjectId("640a603558955e0e2b57c00d"),
    name: 'John Doe',
    age: 30,
    email: 'johndoe@example.com',
    phone: '123-456-7890',
    address: {
        street: '123 Main St',
        city: 'New York',
        state: 'CA'
        },
    salary: 62500,
    startDate: ISODate("2021-01-01T00:00:00.000Z"),
    endDate: null,
    title: 'Manager'
  }
]
```

## $max

The `$max` operator compares a specified value with the value of the given field and updates the field to the specified value if the specified value is greater than the current value of the field.

The below query updates the value of the `age` field to `40` as long as the current value is greater than `40`.

```js
db.employee.updateOne(
    { name: "John Doe" },
    { $max: { age: 40 } }
)
```

This is what the updated document looks like:

```js
[
  {
    _id: ObjectId("640a603558955e0e2b57c00d"),
    name: 'John Doe',
    age: 40,
    email: 'johndoe@example.com',
    phone: '123-456-7890',
    address: { street: '123 Main St', city: 'New York', state: 'CA' },
    salary: 62500,
    startDate: ISODate("2021-01-01T00:00:00.000Z"),
    endDate: null,
    height: 0,
    title: 'Manager'
  }
]
```

## $currentDate

The `$currentDate` operator assigns the current date as the value of a given field.
This can be as a date or timestamp.

To update the `startDate` field with the current date, use the following query:

```js
db.employee.updateOne(
    { name: "John Doe" },
    { $currentDate: { startDate: true } }
)
```

This is the document after the update:

```js
[
  {
    _id: ObjectId("640a603558955e0e2b57c00d"),
    name: 'John Doe',
    age: 40,
    email: 'johndoe@example.com',
    phone: '123-456-7890',
    address: {
        street: '123 Main St',
        city: 'New York',
        state: 'CA' },
    salary: 62500,
    startDate: ISODate("2023-03-10T01:26:35.606Z"),
    endDate: null,
    height: 0,
    title: 'Manager'
  }
]
```
