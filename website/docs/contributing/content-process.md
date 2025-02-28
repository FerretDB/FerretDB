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

### File name and URL

All blog posts should be written in Markdown format.
Start by creating a new file in the `website/blog` directory.

The file name should be in the format `YYYY-MM-DD-shortened-article-url.md` or `YYYY-MM-DD-folder-name/article-url.md`,
where `YYYY-MM-DD` is the date of the post, and `shortened-article-url` is a shortened descriptive title of the post.

### Article template

Please find the blog writing template below and feel free to start writing the draft with it.
You may copy and paste the template into your new blog post file.

```markdown
---
slug: url-slug-of-post # the URL slug of the blog post with dashes instead of spaces
title: Blog post template
authors: [name] # the author name should be stored in the `authors.yml` file and referenced here
description: >
  This is a short description of a FerretDB blog post.
image: /img/blog/postgresql.png
tags: [tag1, tag2]
---

![Image alt description](/img/blog/postgresql.png) <!--Please add the path for the image banner (i.e. /img/blog/banner-image.png).-->

Leave a space before starting article.
Please write a short summary of the article here.
This can be the same as the `description` above.

<!--truncate-->

Start body of the article from here.
This section should contain the rest of the article introduction.

## The content writing process

Each section of the article should be a heading, starting from Heading 2.
```

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

### Formatting

We have a writing guide that provides guidelines for creating clear, concise, and engaging content.
Please see our [writing guide](writing-guide.md) for help formatting your blog post.

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
