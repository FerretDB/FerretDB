---
sidebar_position: 1
---

# Bitwise query operators

Bitwise query operators help to select documents by evaluating query conditions according to the location of bits.

| Operator                         | Description                                                |
| -------------------------------- | ---------------------------------------------------------- |
| [`$bitsAllClear`](#bitsallclear) | Selects documents with clear bit locations (0)             |
| [`$bitsAllSet`](#bitsallset)     | Selects documents with set bit locations (1)               |
| [`$bitsAnyClear`](#bitsanyclear) | Selects documents with at least one clear bit location (0) |
| [`$bitsAnySet`](#bitsanyset)     | Selects documents with at least one set bit location (1)   |

For the examples in this section, insert the following documents into the `numbers` collection:

```js
db.numbers.insertMany([
  { _id: 1, value: 23, binaryValue: '10111' },
  { _id: 2, value: 56, binaryValue: '111000' },
  { _id: 3, value: 67, binaryValue: '1000011' },
  { _id: 4, value: 102, binaryValue: '1100110' },
  { _id: 5, value: 5, binaryValue: '101' },
])
```

## $bitsAllClear

*Syntax*: `{ <field>: { $bitsAllClear: <bitmask> } }`

Use the `$bitsAllClear` operator to select documents where the specified bitmask locations in the query are clear (0).

:::tip
The bitmask can either be a numeric or BinData value.
A BinData value is a BSON type that represents a binary value.
The position of the bits is read from right to left with the rightmost position being `0`.
:::

**Example:** The following query returns documents in which the `value` field has the second and third bit (position `1` and position `2`) from the right as clear (0).

```js
db.numbers.find({
  value: {
    $bitsAllClear: 6,
  },
})
```

The binary representation for `6` in this query is `110`.
The query can also be written to show the positions of the bits to be checked:

```js
db.numbers.find({
  value: {
    $bitsAllClear: [1, 2]
  },
})
```

The output:

```js
[
  { _id: 2, value: 56, binaryValue: '111000' },
]
```

For the same query above, the bitmask can also be written as a BinData value:

```js
db.numbers.find({
  value: {
    $bitsAllClear: BinData(0, 'Bg=='),
  },
})
```

## $bitsAllSet

*Syntax*: `{ <field>: { $bitsAllSet: <bitmask> } }`

To select documents where the bitmask locations in a query are set (1), use the `$bitsAllSet` operator.

**Example:** The following query returns all the documents with positions `1` and positions `2` as set (1):

```js
db.numbers.find({
  value: {
    $bitsAllSet: [1, 2],
  },
})
```

The output:

```js
[
  { _id: 1, value: 23, binaryValue: '10111' },
  { _id: 4, value: 102, binaryValue: '1100110' }
]
```

See the [$BitsallClear query operators](#bitsallclear) section for more usage examples.

## $bitsAnyClear

*Syntax*: `{ <field>: { $bitsAnyClear: <bitmask> } }`

Use the `$bitsAnyClear` operator to select documents where at least one of the bitmask locations in the query is clear (0).

**Example:** The following query returns all the documents with positions `0` and positions `2` as clear (0):

```js
db.numbers.find({
  value: {
    $bitsAnyClear: [0, 2],
  },
})
```

The output:

```js
[
  { _id: 2, value: 56, binaryValue: '111000' },
  { _id: 3, value: 67, binaryValue: '1000011' },
  { _id: 4, value: 102, binaryValue: '1100110' }
]
```

See the [$BitsallClear query operators](#bitsallclear) section for more usage examples.

## $bitsAnySet

*Syntax*: `{ <field>: { $bitsAnySet: <bitmask> } }`

The `$bitsAnySet` operator selects documents where at least one of the bitmask locations in the query is set (1).

**Example:** The following query returns all the documents with positions `0` and positions `2` as set (1):

```js
db.numbers.find({
  value: {
    $bitsAnySet: [0, 2],
  },
})
```

The output:

```js
[
  { _id: 1, value: 23, binaryValue: '10111' },
  { _id: 3, value: 67, binaryValue: '1000011' },
  { _id: 4, value: 102, binaryValue: '1100110' },
  { _id: 5, value: 5, binaryValue: '101' }
]
```

See the [$BitsallClear query operators](#bitsallclear) section for more usage examples.
