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

* [file in the same directory](writing-guide.md)
* [file in a parent directory](../telemetry.md)

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
For example, *ferretdb-queries.jpg*.

### Image links

Use Markdown syntax for images with descriptive alt texts and the path.
All assets (images, gifs, videos, etc.) relating to FerretDB documentation and blog are in the `static/img/` folder.
Rather than use relative paths, we strongly suggest the following approach, since our content engine renders all images directly from the `img` folder.

`![FerretDB logo](/img/logo-dark.png)`.

## Terminologies

To be sure that you're using the right descriptive term, please check our [glossary page](../reference/glossary.md) for relevant terms and terminologies about FerretDB.
If the word is not present in the glossary page, please feel free to ask on Slack or in the blog post issue.
