---
sidebar_position: 99
unlisted: true # linked from CONTRIBUTING.md
---

# Writing guide

## Front matter

The front matter represents the metadata for each page.
It is written at the top of each page and must be enclosed by `---` on both sides.

Example:

```yaml
---
sidebar_position: 1
description: How to write documentation
---
```

Learn more about [front matter in Docusaurus](https://docusaurus.io/docs/api/plugins/@docusaurus/plugin-content-docs#markdown-front-matter).

## Names and URLs

Use `kebab-case-with-dashes` instead of `snake_case_with_underscores` or spaces
for file names, directory names, and slugs because URL paths typically use dashes.

Ensure that the file name/URL path matches the title of the page.
For example, if the title of your page is "Getting Started", then the file name/URL path should also be "getting-started" to maintain consistency.
The `slug` field should be the same as the file name.
Only use a different `slug` field in some special cases, such as for backward compatibility with existing links.

## Sidebar position

Use the `sidebar_position` in the front matter to set the order of the pages in the sidebar.
Please ensure that the `sidebar_position` is unique for each page in that directory.
For example, if there are several pages in the folder "Getting Started", let `sidebar_position` equal "1", "2", "3", "4", and so on to avoid duplication.

## Headers

Use sentence case for headers: `### Some header with URL`, not `### Some Header With URL`.

## Links

Please use relative `.md` file paths for links.
It is required for [documentation versioning](https://docusaurus.io/docs/versioning#link-docs-by-file-paths).

Examples:

To link to a file in the same directory, use the file name.

```text
[file in the same directory](writing-guide.md)
```

To link to file in a different directory, specify the relative file path.

```text
[file in a different directory](../basic-operations/read.md)
```

When referencing files such as configuration files, specs, or internal definitions on GitHub, ensure that the link references a specific release tag, not the `main` branch.
This is important because the `main` branch may change frequently, and links to it may break.

For example, if you want to link to the FerretDB Data API OpenAPI 3.0 specification, use the following link:

```text
[FerretDB Data API OpenAPI 3.0 specification](https://raw.githubusercontent.com/FerretDB/FerretDB/refs/tags/v2.7.0/internal/dataapi/api/openapi.json)
```

## Images

Please store all images under `blog` or `docs` in the `static/img` folder.

Also, you can collate images for a specific blog post inside a single folder.
For partner blog posts, store related images in the same folder, as `/img/blog/partner-name/image.png`.

Otherwise, name the folder appropriately using the `YYYY-MM-DD` format, for example, a typical path for an image will be `/img/blog/2023-01-01/ferretdb-image.jpg`.

### Alt text

Please remember to add an alternate text for images.
The alt text should provide a description of the image for the user.
When you add a banner image, please use the title of the article as the alt text.

### Image names

Use of two or three descriptive words written in `kebab-case-with-dashes` as the names for the images.
For example, _ferretdb-queries.jpg_.

### Image links

Use Markdown syntax for images with descriptive alt texts and the path.
All assets (images, gifs, videos, etc.) relating to FerretDB documentation and blog are in the `static/img/` folder.
Rather than use relative paths, we strongly suggest the following approach, since our content engine renders all images directly from the `img` folder.

`![FerretDB logo](/img/logo-dark.png)`.

## Lists

Lists should describe a sequence of items, such as a series of steps, features, or group related items.
They should not be used for highlighting or emphasizing a single item; use code blocks or bold text instead.

Our formatting tool will automatically reformat lists.

## Code blocks

Code blocks should be used for code snippets, including shell commands, SQL queries, and JSON documents.
It can also be used to highlight specific texts, including URLs, file names, and other important information.

Always specify the language in Markdown code blocks.

### MongoDB shell commands and results

#### Documentation

For our documentation, we use the CTS tool to test and validate the code snippets for the MongoDB shell commands and responses.
Related MongoDB shell commands and responses should reside in the same directory as the documentation file in extended JSON format.
See this [TTL indexes example](../guides/ttl-indexes.json) for reference.
The code snippet prefix `1-` (found in `1-<file-name>.json` file) in ascending order is used to enforce the order in the documentation and their execution within the CTS tool.

The CTS tool will be responsible for generating the formatted code snippets which can be imported into MDX files.
Run `task docs-gen` to generate the formatted code snippets.
The generated code snippets will be stored in `.js` files under `website/docs/guides/<extended-json-file-name>` directory.

#### Blog posts

For blog posts, please use `js` language for MongoDB shell commands.

Our tooling will automatically reformat those blocks.

```js
db.league.find({ club: 'PSG' })
```

For MongoDB shell results, use `js` language, assign the `mongosh` output to `response` and copy&paste it as-is,
with unquoted field names, single quotes for strings, without trailing commas, etc.
Our tooling will not reformat those blocks.

```js
//Assign the output to response
response = [
  {
    _id: ObjectId('63109e9251bcc5e0155db0c2'),
    club: 'PSG',
    points: 30,
    average_age: 30,
    discipline: { red: 5, yellow: 30 },
    qualified: false
  }
]
```

### Other code blocks

The following formatting instructions apply for both documentation and blog posts.

Use `sql` for SQL queries.

```sql
SELECT _jsonb FROM "test"."_ferretdb_database_metadata" WHERE ((_jsonb->'_id')::jsonb = '"customers"');
```

For `psql` output, environment variables, and in all other cases, use `text`.

```text
 _jsonb ----------------------------------------------------------------------------------------------------------------------------------------------
 {"$s": {"p": {"_id": {"t": "string"}, "table": {"t": "string"}}, "$k": ["_id", "table"]}, "_id": "customers", "table": "customers_c09344de"}
```

```text
ferretdb=# \d test._ferretdb_settings
          Table "test._ferretdb_settings"
  Column  | Type  | Collation | Nullable | Default
----------+-------+-----------+----------+---------
 settings | jsonb |           |          |

ferretdb=# SELECT settings FROM test._ferretdb_settings;
                                             settings
--------------------------------------------------------------------------------------------------
 {"$k": ["collections"], "collections": {"$k": ["groceries"], "groceries": "groceries_6a5f9564"}}
(1 row)
```

## Terminologies

To be sure that you're using the right descriptive term, please check our [glossary page](../reference/glossary.md) for relevant terms and terminologies about FerretDB.
If the word is not present in the glossary page, please feel free to ask on Slack or in the blog post issue.
