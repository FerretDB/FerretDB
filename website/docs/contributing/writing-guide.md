---
sidebar_position: 99
draft: true
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
It is recommended to omit the `slug` field from the front matter, since the file name is used by default as the URL path.
Only use the `slug` field in some special cases, such as when creating custom or external links, or for backward compatibility with existing links.

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
Examples:

- [file in the same directory](writing-guide.md)
- [file in a parent directory](../telemetry.md)

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

## Tags

Tags are an important part of every blog post, and they appear at the top of the front matter.
They make it easy for readers to search and identify related blog posts based on their categories or subject matter.
You can view all [currently listed tags here](https://blog.ferretdb.io/tags/).

Please note that tags are case-sensitive, such that `Release` and `release` are two separate tags.
Unless distinctly written (as in the case with CI/CD, DevOps), all tags should be in small-case.

A blog post can have as many tags as possible, as long as it is relevant to the post.
Please only include the following tags:

- cloud
- community
- compatible applications
- demo
- document databases
- events
- hacktoberfest
- javascript frameworks
- mongodb compatible
- mongodb gui
- open source
- postgresql tools
- product
- release
- sspl
- tutorial

This is not an exhaustive list, and the direction of our blog posts can surely expand.
If a blog post calls for a new tag or you would like to suggest more tags, please ensure to add it to this list.
This helps to maintain consistency across all blog posts.

## Keywords (optional)

Keywords in the front matter are displayed as meta keywords tag in HTML.
Meta keywords tag are not so important for SEO anymore, but they can help with focusing the content on specific keywords that should be used (appear at least once) in the blog content, meta description, title, or alt images.

The use of meta keywords is not mandatory, but if you want to add them, please ensure that they are relevant to the blog post.

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
