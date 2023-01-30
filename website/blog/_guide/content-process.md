---
sidebar_position: 2
unlisted: true
---

# Content Creation Guide for FerretDB

## Introduction

This guide will walk you through the process of creating and publishing content on the FerretDB blog, which is powered by Docusaurus.

## Suggesting a blog post

Everybody is welcome to contribute and write a blog post for FerretDB.
The first step to creating a blog post, pitching a draft, or idea is to [create an issue here](https://github.com/FerretDB/engineering/issues/new/choose).
Please note that the FerretDB blog represents the official voice of the company and product, and as such, all published content will be carefully vetted and reviewed.

## Creating a New Blog Post

Once the blog post pitch is approved, you can start creating the draft.

### Article Template

We have provided a template for writing blog posts.
Please find the [template](YYYY-MM-DD.blog-template.md) here and feel free to start writing the draft with the provided template.

### File name

The file name should be in the format `YYYY-MM-DD-title.md`, where `YYYY-MM-DD` is the date of the post or issue, and `title` is a short, descriptive title of the post or issue.

### Writing Guide

We have a writing guide that provides guidelines for creating clear, concise, and engaging content.
Please see our [writing guide](writing-guide.md) for help formatting your blog post.

## Setting Front Matter and Publishing Options

Front matter is the metadata that appears at the top of the markdown file and provides information about the post or issue, such as the title, author, and date.

In the front matter, ensure to set the `unlisted: true` in the front matter until it's ready to publish.
Make sure to include all necessary information in the front matter, such as the title, author, and date.

## Reviewing and Editing Content

To publish a blog post, you will need to create a Pull Request with a file of your blog post content formatted in Markdown.
However, before publishing any content, it must be reviewed and edited by the FerretDB Team to ensure that it meets our standards for quality and accuracy.
The review process will vary depending on the size and complexity of the content.
It may involve one or more rounds of editing.

Before opening a PR, be sure to double-check the content for any errors or inconsistencies, such as spelling mistakes or broken links.
Please preview the blog post to make sure that it is properly formatted and looks as you expect it to.

Once the content is ready for review, please open a PR and assign it to @Ferretdb/core.

## Final Approval and Publishing

The final approval for publishing content is given once it has passed through all reviews and approved by the team.
To publish the content, change the date in the front matter to the proposed published date, and then remove `unlisted: true` from the front matter.

## Post Publishing

Once the article is published, we encourage you to share the blog post across your social media pages.
The article may be updated from time-to-time to ensure we are putting the most accurate information out, and to improve search engine rankings.
