// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBlogSlugs(t *testing.T) {
	m := fstest.MapFS{
		"2022-05-16-using-cla-assistant-with-ferretdb.md": {
			Data: []byte(`---
slug: using-cla-assistant-with-ferretdb
title: "Using CLA Assistant with FerretDB"
author: Alexey Palazhchenko
description: Like many other open-source projects FerretDB 
image: /img/blog/cla3.jpg
date: 2022-05-16
---
Finally, we need a web server that would handle HTTPS for us.
For that, we will use
Caddy will listen on both HTTP and
`),
		},
	}

	dirs, err := m.ReadDir(".")
	require.NoError(t, err)

	slugs := getBlogSlugs(dirs)
	tSlug := fileSlug{fileName: "2022-05-16-using-cla-assistant-with-ferretdb.md", slug: "using-cla-assistant-with-ferretdb"}
	assert.Equal(t, slugs[0], tSlug, "should be equal")
}

func TestVerifySlugs(t *testing.T) {
	m := fstest.MapFS{
		"2022-05-16-using-cla-assistant-with-ferretdb.md": {
			Data: []byte(`---
slug: using-cla-assistant-with-ferretdb
title: "Using CLA Assistant with FerretDB"
author: Alexey Palazhchenko
description: Like many other open-source projects
image: /img/blog/cla3.jpg
date: 2022-05-16
---
Finally, we need a web server that would handle HTTPS for us.
For that, we will use [Caddy](https://caddyserver.com):
Caddy will listen on both HTTP and HTTPS ports, 
`),
		},
	}

	dirs, err := m.ReadDir(".")
	require.NoError(t, err)

	slugs := getBlogSlugs(dirs)

	f, err := m.Open("2022-05-16-using-cla-assistant-with-ferretdb.md")
	require.NoError(t, err)

	err = verifySlug(slugs[0], f)
	require.NoError(t, err)
}

func TestVerifyDateNotPresent(t *testing.T) {
	m := fstest.MapFS{
		"2022-05-16-using-cla-assistant-with-ferretdb.md": {
			Data: []byte(`---
slug: using-cla-assistant-with-ferretdb
title: "Using CLA Assistant with FerretDB"
author: Alexey Palazhchenko
description: Like many other open-source projects
image: /img/blog/cla3.jpg
---
Finally, we need a web server that would handle HTTPS for us.
For that, we will use [Caddy](https://caddyserver.com):
Caddy will listen on both HTTP and HTTPS ports, 
`),
		},
	}

	_, err := m.ReadDir(".")
	require.NoError(t, err)

	f, err := m.Open("2022-05-16-using-cla-assistant-with-ferretdb.md")
	require.NoError(t, err)

	err = verifyDateNotPresent(f)
	require.NoError(t, err)
}

func TestVerifyDateNotPresentForError(t *testing.T) {
	m := fstest.MapFS{
		"2022-05-16-using-cla-assistant-with-ferretdb.md": {
			Data: []byte(`---
slug: using-cla-assistant-with-ferretdb
title: "Using CLA Assistant with FerretDB"
author: Alexey Palazhchenko
description: Like many other open-source projects
image: /img/blog/cla3.jpg
date: 2022-05-16
---
Finally, we need a web server that would handle HTTPS for us.
For that, we will use [Caddy](https://caddyserver.com):
Caddy will listen on both HTTP and HTTPS ports, 
`),
		},
	}

	_, err := m.ReadDir(".")
	require.NoError(t, err)

	f, err := m.Open("2022-05-16-using-cla-assistant-with-ferretdb.md")
	require.NoError(t, err)

	err = verifyDateNotPresent(f)
	assert.Equal(t, err.Error(), "date field should not be present in the front matter")
}

func TestVerifyTags(t *testing.T) {
	m := fstest.MapFS{
		"2022-05-16-using-cla-assistant-with-ferretdb.md": {
			Data: []byte(`---
slug: using-cla-assistant-with-ferretdb
title: "Using CLA Assistant with FerretDB"
author: Alexey Palazhchenko
description: Like many other open-source projects
image: /img/blog/cla3.jpg
date: 2022-05-16
tags: [events, document databases, hacktoberfest]
---
Finally, we need a web server that would handle HTTPS for us.
For that, we will use [Caddy](https://caddyserver.com):
Caddy will listen on both HTTP and HTTPS ports, 
`),
		},
	}

	_, err := m.ReadDir(".")
	require.NoError(t, err)

	f, err := m.Open("2022-05-16-using-cla-assistant-with-ferretdb.md")
	require.NoError(t, err)

	err = verifyTags(f)
	require.NoError(t, err)
}

func TestVerifyTagsForError(t *testing.T) {
	m := fstest.MapFS{
		"2022-05-16-using-cla-assistant-with-ferretdb.md": {
			Data: []byte(`---
slug: using-cla-assistant-with-ferretdb
title: "Using CLA Assistant with FerretDB"
author: Alexey Palazhchenko
description: Like many other open-source projects
image: /img/blog/cla3.jpg
date: 2022-05-16
tags: [events, documentss databases, hacktoberfest]
---
Finally, we need a web server that would handle HTTPS for us.
For that, we will use [Caddy](https://caddyserver.com):
Caddy will listen on both HTTP and HTTPS ports, 
`),
		},
	}

	_, err := m.ReadDir(".")
	require.NoError(t, err)

	f, err := m.Open("2022-05-16-using-cla-assistant-with-ferretdb.md")
	require.NoError(t, err)

	err = verifyTags(f)
	assert.Equal(t, err.Error(), "tag 'documentss' is not in the allowed list")
}
