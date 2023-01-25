---
title: "This is the title of the FerretDB Blog post"
slug: url-slug-of-post #the URL slug of the blog post
author: Firstname Lastname
author_url: #link to website or GitHub account
author_image_url: #link to author profile image # optional – please remove this line if you don't want to use an author image. See below for more info
description: This is a short description of a FerretDB blog post.
keywords:
  - keyword1
  - keyword2
image: /img/blog/post-cover-image.jpg
tags: [tag1, tag2]
unlisted: true
---

Leave a space before starting article.
Please write a short summary of the article here.
This can be the same as the `description` above.
Then add the feature image of the article

![Image alt description](/img/blog/banner-image.jpg) <!---Please add the path for the image banner (i.e. /img/blog/banner-image.png).-->

<!--truncate-->

Start body of the article from here.
This section should contain the rest of the article introduction.

## The content writing process (Start from Heading 2 since the blog title represents Heading 1)

All blog posts are written in Markdown.
Please add Markdown files to the `blog` directory.
Files should be in this format `YYYY-MM-DD-shortened-article-name.md` or `YYYY-MM-DD-folder-name/article-name.md`.

All images for this blog post - including the banner image - should be stored in this folder `(../../static/img/blog)`.
You can also store the images in a folder with the blog post date under this directory for example, `/img/blog/2022-12-29/banner.png`.

Regular blog authors and engineers can be added to `authors.yml`.
When the author information is present in `authors.yml`, fill the unique author name in the frontmatter as (`authors: [name tag]`).
Ensure to remove all other descriptions in the frontmatter about the author (`author`, `author_url`, `author title`, and `author_image_url`).

## formatting guidelines

See [content writing guidelines](../../docs/contributing/writing-guide.md) for more information.
