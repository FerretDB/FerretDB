---
title: FerretDB Blog post
slug:
author:
author_title:
author_url:
author_image_url:
description: This is a description of a FerretDB blog post.
keywords:
  - keyword1
  - keyword2
image:
tags: [tag1, tag2]
---

Leave a space before starting article.
Please write a short summary of the article here.
This can be the same as the `description` above.

<!--truncate-->

![Image alt description](path) <!---Please add the image banner path for the article (i.e. /img/blog/2022-12-29/banner.png).-->

Start body of the article from here.

## The content writing process (Start from Heading 2 since the blog title represents Heading 1)

Please add Markdown files (or folders containing Markdown files) to the `blog` directory.
Files should be in this format `YYYY-MM-DD-shortened-article-name.md` or `YYYY-MM-DD-folder-name/article-name.md`.

All images for this blog post - including the banner image - should be stored in a folder with the blogpost date under this directory `(../../static/img/blog)`, for example, `/img/blog/2022-12-29/banner.png`.

Regular blog authors and engineers can be added to `authors.yml`.
When the author information is present in `authors.yml`, fill the unique author name in the frontmatter as (`authors: [name tag]`).
Ensure to remove all other descriptions in the frontmatter about the author (`author`, `author_url`, `author title`, and `author_image_url`).
Otherwise, please enter all author information in the frontmatter.

### SubHeading
