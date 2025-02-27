---
slug: url-slug-of-post # the URL slug of the blog post with dashes instead of spaces
title: Blog post template
author: Firstname Lastname
author_url: # link to website or GitHub account
author_image_url: # link to author profile image (optional – please remove this line if you don't want to use an author image. See below for more info)
description: >
  This is a short description of a FerretDB blog post.
keywords:
  - keyword1
  - keyword2
image: /img/blog/postgresql.png
tags: [tag1, tag2]
unlisted: true
sidebar_position: 99
---

![Image alt description](/img/blog/postgresql.png) <!--Please add the path for the image banner (i.e. /img/blog/banner-image.png).-->

Leave a space before starting article.
Please write a short summary of the article here.
This can be the same as the `description` above.

<!--truncate-->

Start body of the article from here.
This section should contain the rest of the article introduction.

## The content writing process (Start from Heading 2 since the blog title represents Heading 1)

Blog authors should be added to `authors.yml`.
When the author information is present in `authors.yml`, fill the unique author name in the frontmatter as (`authors: [name1 name2]`).
Ensure to remove all other descriptions in the frontmatter about the author (`author`, `author_url`, `author title`, and `author_image_url`).

## formatting guidelines

See [content writing guidelines](writing-guide.md) for more information.
