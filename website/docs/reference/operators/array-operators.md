---
sidebar_position: 3
---

# Array query operators

Array query operators allow you to search for specific elements within an array field in a document.

| Operator                   | Description                                                                                                             |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| [`$all`](#all)             | Matches an array that contains all the elements in the specified query                                                  |
| [`$elemMatch`](#elemmatch) | Matches a document that contains an array field with at least one element that matches all the specified query criteria |
| [`$size`](#size)           | Matches an array with a specified number of elements                                                                    |

For the examples in this section, insert the following documents into the `team` collection:

```js
db.team.insertMany([
   {
      id: 1,
      name: "Jack Smith",
      position: "Manager",
      skills: ["leadership", "communication", "project management"],
      contact: {
         email: "john@example.com",
         phone: "123-456-7890"
      },
      active: true
   },
   {
      id: 2,
      name: "Jane Mark",
      position: "Software Developer",
      skills: ["Java", "Python", "C++"],
      contact: {
         email: "jane@example.com",
         phone: "123-456-7891"
      },
      active: false
   },
   {
      id: 3,
      name: "Bob Johnson",
      position: "Graphic Designer",
      skills: ["Adobe Photoshop", "Illustrator", "InDesign"],
      contact: {
         email: "bob@example.com",
         phone: "123-456-7892"
      },
      active: true
   },
   {
      id: 4,
      name: "Alice Williams",
      position: "Marketing Coordinator",
      skills: ["communication", "content creation", "event planning"],
      contact: {
         email: "alice@example.com",
         phone: "123-456-7893"
      },
      active: true
   }
])
```

## $all

*Syntax*: `{ <field>: { $all: [ <value1>, <value2>, ... ] } }`

The `$all` operator matches documents where the field value is an array that contains all the specified elements.

For example, find all documents in the `team` collection where the `skills` field contains both `communication` and `content creation` as elements using the following query operation:

```js
db.team.find({
   "skills": {
      $all: ["communication", "content creation"]
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("63a5bb4acf72d6203bb45bb5"),
    id: 4,
    name: 'Alice Williams',
    position: 'Marketing Coordinator',
    skills: [ 'communication', 'content creation', 'event planning' ],
    contact: { email: 'alice@example.com', phone: '123-456-7893' },
    active: true
  }
]
```

## $elemMatch

*Syntax*: `{ <field>: { $elemMatch: { <query1>, <query2>, ... } } }`

The `$elemMatch` operator matches documents where the field value is an array that contains at least one element that matches the given condition.

For example, to find documents in the `team` collection where the `skills` field is an array that contains the element "Java", use the following query with the `$elemMatch` operator:

```js
db.team.find({
   skills: {
      $elemMatch: {
         $eq: "Java",
         $nin: [ "communication" ]
      }
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("63aa247e69c82de72bd40b93"),
    id: 2,
    name: 'Jane Mark',
    position: 'Software Developer',
    skills: [ 'Java', 'Python', 'C++' ],
    contact: { email: 'jane@example.com', phone: '123-456-7891' },
    active: false
  }
]
```

## $size

*Syntax*: `{ <field>: { $size: <number> } }`

The `$size` operator matches arrays with a specified number of elements.

Select the documents in the `team` collection where the `skills` array contains only three elements.

```js
db.team.find({
   skills: {
      $size: 3
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("63aa247e69c82de72bd40b92"),
    id: 1,
    name: 'Jack Smith',
    position: 'Manager',
    skills: [ 'leadership', 'communication', 'project management' ],
    contact: { email: 'john@example.com', phone: '123-456-7890' },
    active: true
  },
  {
    _id: ObjectId("63aa247e69c82de72bd40b93"),
    id: 2,
    name: 'Jane Mark',
    position: 'Software Developer',
    skills: [ 'Java', 'Python', 'C++' ],
    contact: { email: 'jane@example.com', phone: '123-456-7891' },
    active: false
  },
  {
    _id: ObjectId("63aa247e69c82de72bd40b94"),
    id: 3,
    name: 'Bob Johnson',
    position: 'Graphic Designer',
    skills: [ 'Adobe Photoshop', 'Illustrator', 'InDesign' ],
    contact: { email: 'bob@example.com', phone: '123-456-7892' },
    active: true
  },
  {
    _id: ObjectId("63aa247e69c82de72bd40b95"),
    id: 4,
    name: 'Alice Williams',
    position: 'Marketing Coordinator',
    skills: [ 'communication', 'content creation', 'event planning' ],
    contact: { email: 'alice@example.com', phone: '123-456-7893' },
    active: true
  }
]
```
