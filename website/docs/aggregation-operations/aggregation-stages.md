---
sidebar_position: 2
---

# Aggregation stages

Aggregation stages are a series of one or more processes in a pipeline that acts upon the returned result of the previous stage, starting with the input documents.

| Supported aggregation stages | Description                                                                                           |
| ---------------------------- | ----------------------------------------------------------------------------------------------------- |
| `$count`                     | Returns the count of all matched documents in a specified query                                       |
| `$group`                     | Groups documents based on specific value or expression and returns a single document for each group   |
| `$limit`                     | Limits specific documents and passes the rest to the next stage                                       |
| `$match`                     | Acts as a `find` operation by only returning documents that match a specified query to the next stage |
| `$project`                   | Specifies `n` number of fields in a document to pass to the next stage in the pipeline                |
| `$skip`                      | Skips a specified `n` number of documents and passes the rest to the next stage                       |
| `$sort`                      | Sorts and returns all the documents based on a specified order                                        |
| `$unset`                     | Specifies `n` number of fields to be removed/excluded from a document                                 |
| `$unwind`                    | Deconstructs and returns a document for every element in an array field                               |
