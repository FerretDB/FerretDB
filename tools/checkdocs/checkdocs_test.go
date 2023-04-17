package main

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestGetBlogSlugs(t *testing.T) {
	m := fstest.MapFS{
		"2022-05-16-using-cla-assistant-with-ferretdb.md": {
			Data: []byte(`---
slug: using-cla-assistant-with-ferretdb
title: "Using CLA Assistant with FerretDB"
author: Alexey Palazhchenko
description: Like many other open-source projects, FerretDB requires all contributors to sign [our Contributor License Agreement (CLA)](https://gist.github.com/ferretdb-bot/554e6a30bfcc1d954f3853b4aad95281) to protect them from liability.
image: /img/blog/cla3.jpg
date: 2022-05-16
---
Finally, we need a web server that would handle HTTPS for us.
For that, we will use [Caddy](https://caddyserver.com):
Caddy will listen on both HTTP and HTTPS ports, and retrieve the TLS certificate from Let’s Encrypt that will be stored in “./data/caddy” on the host.
`),
		},
	}
	dirs, _ := m.ReadDir(".")
	slugs := GetBlogSlugs(dirs)
	assert.Equal(t, slugs[0], FileSlug{fileName: "2022-05-16-using-cla-assistant-with-ferretdb.md", slug: "using-cla-assistant-with-ferretdb"}, "should be equal")
}

func TestVerifySlugs(t *testing.T) {
	m := fstest.MapFS{
		"2022-05-16-using-cla-assistant-with-ferretdb.md": {
			Data: []byte(`---
slug: using-cla-assistant-with-ferretdb
title: "Using CLA Assistant with FerretDB"
author: Alexey Palazhchenko
description: Like many other open-source projects, FerretDB requires all contributors to sign [our Contributor License Agreement (CLA)](https://gist.github.com/ferretdb-bot/554e6a30bfcc1d954f3853b4aad95281) to protect them from liability.
image: /img/blog/cla3.jpg
date: 2022-05-16
---
Finally, we need a web server that would handle HTTPS for us.
For that, we will use [Caddy](https://caddyserver.com):
Caddy will listen on both HTTP and HTTPS ports, and retrieve the TLS certificate from Let’s Encrypt that will be stored in “./data/caddy” on the host.
`),
		},
	}
	dirs, _ := m.ReadDir(".")
	slugs := GetBlogSlugs(dirs)
	f, _ := m.Open("2022-05-16-using-cla-assistant-with-ferretdb.md")
	VerifySlug(slugs[0], f)
}
