---
sidebar_position: 99
unlisted: true
---

# Writing guide

## File names

Use `kebab-case-with-dashes` instead of `snake_case_with_underscores` for file names because URL paths typically use dashes.

## Links

Please use markdown file paths for links, not URL paths,
because it works for both editors/IDEs (Ctrl/⌘+click works) and Docusaurus.
Always add `.md` extension to the file paths.
Examples:

* [file in the same directory](contributing.md)
* [file in a parent directory](../telemetry.md)

## Images

Please store all images under `blog` or `docs` in the `static/img` folder.
Also, you can collate images for a specific blog post inside a single folder.
Name the folder appropriately using the `YYYY-MM-DD` format.
For example, a typical path for an image will be `/img/blog/2023-01-01/ferretdb-image.jpg`

### Alt Text

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

`![FerretDB logo](/img/logo_dark.png)`.

## Terminologies

To be sure that you're using the right descriptive term, please check our [glossary page](../reference/glossary.md) for relevant terms and terminologies about FerretDB.
If the word is not present in the glossary page, please feel free to ask on Slack or in the blog post issue.
