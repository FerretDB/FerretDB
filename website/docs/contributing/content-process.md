---
sidebar_position: 3
unlisted: true
---

# Blog creation guide

## Introduction

This guide will walk you through the process of creating and publishing content on the FerretDB blog, which is powered by Docusaurus.

## Suggesting a blog post

Everybody is welcome to contribute and write a blog post for FerretDB.
The first step to creating a blog post, pitching a draft, or idea is to [create an issue here](https://github.com/FerretDB/engineering/issues/new/choose).
Please note that the FerretDB blog represents the official voice of the company and product, and as such, all published content will be carefully vetted and reviewed.

## Creating a new blog post

Once the blog post pitch is approved, you can start creating the draft.

### Article template

We have provided a template for writing blog posts.
Please find the [template](YYYY-MM-DD.blog-template.md) here and feel free to start writing the draft with the provided template.

### Names and URLs

The file name should be in the format `YYYY-MM-DD-shortened-article-url.md` or `YYYY-MM-DD-folder-name/article-url.md`,
where `YYYY-MM-DD` is the date of the post, and `shortened-article-url` is a shortened descriptive title of the post.

### Tags

Tags are an important part of every blog post, and they appear at the top of the front matter.
They make it easy for readers to search and identify related blog posts based on their categories or subject matter.
You can view all [currently listed tags here](https://blog.ferretdb.io/tags/).

Please note that all tags must be in small-case, such that `FerretDB` should be written as `ferretdb`.

Hyphens should be disregarded when writing tags, e.g. `mongodb-compatible database` should be written as `mongodb compatible database`.

A blog post can have as many tags as possible, as long as it is relevant to the post.
Please only include the following tags (and keep them in sync with the `checkdocs` linter):

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
- observability
- open source
- postgresql tools
- product
- release
- sspl
- tutorial

This is not an exhaustive list, and the direction of our blog posts can surely expand.
If a blog post calls for a new tag or you would like to suggest more tags, please ensure to add it to this list.
This helps to maintain consistency across all blog posts.

### Writing guide

We have a writing guide that provides guidelines for creating clear, concise, and engaging content.
Please see our [writing guide](writing-guide.md) for help formatting your blog post.

## Setting front matter and publishing options

Front matter is the metadata that appears at the top of the markdown file and provides information about the post, such as the title and author.

In the front matter, `draft: true` keeps the page hidden from the visitors of the site.
Use this option if you plan to merge the content into `main` branch while keeping the page hidden.
Other cases, remove `draft: true` to enable CI to build and render the new content during the reviewing.
Make sure to include all necessary information in the front matter, such as the title and author.

## Reviewing and editing content

To publish a blog post, you will need to create a Pull Request with a file of your blog post content formatted in Markdown.
However, before publishing any content, it must be reviewed and edited by the FerretDB Team to ensure that it meets our standards for quality and accuracy.
The review process will vary depending on the size and complexity of the content.
It may involve one or more rounds of editing.

Before opening a PR, be sure to double-check the content for any errors or inconsistencies, such as spelling mistakes or broken links.
Please preview the blog post to make sure that it is properly formatted and looks as you expect it to.

Once the content is ready for review, please open a PR and assign it to @Ferretdb/docs.

## Final approval and publishing

The final approval for publishing content is given once it has passed through all reviews and approved by the team.
To publish the content, change the date in file name to the proposed published date, and then remove `draft: true` if set from the front matter.

## Post publishing

Once the article is published, we encourage you to share the blog post across your social media pages.
The article may be updated from time-to-time to ensure we are putting the most accurate information out, and to improve search engine rankings.
