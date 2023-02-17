---
sidebar_position: 5
---

# Evaluation query operators

Evaluation query operators return data based on the evaluation of a specified expression.

| Operator           | Description                                                                                    |
| ------------------ | ---------------------------------------------------------------------------------------------- |
| [`$mod`](#mod)     | Matches documents where the field element modulo a given value equals the specified remainder. |
| [`$regex`](#regex) | Matches documents where a field matches a specified regular expression query                   |

For the examples in this section, insert the following documents into the `catalog` collection:

```js
db.catalog.insertMany([
   {
      product: "bottle",
      price: 15,
      stock: 1,
   },
   {
      product: "spoon",
      price: 500,
      stock: 0,
   },
   {
      product: "cup",
      price: 100,
      stock: 14,
   },
   {
      product: "BoWL",
      price: 56,
      stock: 5,
   },
   {
      product: "boTtLe",
      price: 20,
      stock: 3,
   }
])
```

## $mod

*Syntax*: `{ <field>: { $mod: [ <divisor-value>, <modulus> ] } }`

The `$mod` operator matches documents where the field element modulo (`%`) a specified divisor returns a given modulus.

**Example:** The following query returns all the documents where the value of the "stock" field is evenly divisible by 2:

```js
db.catalog.find({
   stock:{
      $mod: [ 2, 0 ]
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("63e3ac0184f488929a3f737a"),
    product: 'spoon',
    price: 500,
    stock: 0
  },
  {
    _id: ObjectId("63e3ac0184f488929a3f737b"),
    product: 'cup',
    price: 100,
    stock: 14
  }
]
```

:::caution
Note that the `$mod` expression returns an error if you only have a single element in the array, more than two elements in the array, or if the array is empty.
It also rounds down decimal input down to zero (e.g. `$mod: [ 3.5 , 2 ]` is executed as `$mod: [ 3 , 2 ]`).
:::

## $regex

*Syntax*: `{ <field>: { $regex: '<expression-string>', $options: '<flag>' } }`

Other syntaxes: `{ <field>: { $regex: /<expression-string>/, $options: '<flag>' } }` and `{ <field>: /<expression-string>/<flag> }`.

The `$regex` operator matches documents where the value of a field matches a specified regular expression pattern.

**Example:** The following query returns all the documents where the value of the "product" field starts with the letter "b":

```js
db.catalog.find({
   product:{
      $regex: /^b/
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("63e4ce469695494b86bf2b2d"),
    product: 'bottle',
    price: 15,
    stock: 1
  },
  {
    _id: ObjectId("63e4ce469695494b86bf2b31"),
    product: 'boTtLe',
    price: 20,
    stock: 3
  }
]
```

`$options` is an optional parameter that specifies the regular expression flags to use, such as:

* Case-insensitivity (`i`)
* Multi-line matching (`m`)
* Dot character matching (`s`)

:::note
The regex flag for ignoring white spaces (`x`) is not currently supported.
Follow [here](https://github.com/FerretDB/FerretDB/issues/592) for more updates.
:::

To perform case-insensitive matching, use the `i` flag in the `regex` expression.

**Example:** The following query returns all the documents where the value of the "product" field is equal to "bottle" (case-insensitive):

```js
db.catalog.find({
   product: {
      $regex: /bottle/i
   }
})
```

The output:

```js
[
  {
    _id: ObjectId("63e3ac0184f488929a3f7379"),
    product: 'bottle',
    price: 15,
    stock: 1
  },
  {
    _id: ObjectId("63e3ac0184f488929a3f737d"),
    product: 'boTtLe',
    price: 20,
    stock: 3
  }
]
```
