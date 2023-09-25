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

Please use markdown file paths for links, not URL paths,
because it works for both editors/IDEs (Ctrl/⌘+click works) and Docusaurus.
Always add `.md` extension to the file paths.
Use relative paths for links to files in the same directory, in a sub-directory, or in a parent directory.

Examples:

To link to a file in the same directory, use the file name.

- `[file in the same directory](writing-guide.md)`

To link to a file in a parent directory, prefix with `../` to go up one directory level.

- `[file in a parent directory](../telemetry.md)`

To link to file in a subdirectory, specify the file path along with its respective directory or directories, such as: `subdirectory/file.md`.

To link to a directory or category, prefix the directory name with `/category/`.

- `[configuration directory](/category/configuration/)`

## Images

Please store all images under `blog` or `docs` in the `static/img` folder.
Also, you can collate images for a specific blog post inside a single folder.
Name the folder appropriately using the `YYYY-MM-DD` format.
For example, a typical path for an image will be `/img/blog/2023-01-01/ferretdb-image.jpg`

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

## Code blocks

Always specify the language in Markdown code blocks.

For MongoDB shell commands, use `js` language.
Our tooling will automatically reformat those blocks.

```js
db.league.find({ club: 'PSG' })
```

For MongoDB shell results, use `json5` language and copy&paste the output as-is,
with unquoted field names, single quotes for strings, without trailing commas, etc.
Our tooling will not reformat those blocks.

```json5
[
  {
    _id: ObjectId("63109e9251bcc5e0155db0c2"),
    club: 'PSG',
    points: 30,
    average_age: 30,
    discipline: { red: 5, yellow: 30 },
    qualified: false
  }
]
```

Use `sql` for SQL queries.
Use `text` for the `psql` output and in other cases.

```sql
SELECT _jsonb FROM "test"."_ferretdb_database_metadata" WHERE ((_jsonb->'_id')::jsonb = '"customers"');
```

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
