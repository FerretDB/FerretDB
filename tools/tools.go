// Copyright 2021 Baltoro OÜ.
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

//go:build tools
// +build tools

package tools // import "github.com/MangoDB-io/MangoDB/tools"

import (
	_ "github.com/BurntSushi/go-sumtype"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/quasilyte/go-consistent"
	_ "github.com/reviewdog/reviewdog/cmd/reviewdog"
	_ "golang.org/x/perf/cmd/benchstat"
	_ "golang.org/x/tools/cmd/stringer"
	_ "mvdan.cc/gofumpt"
)

//go:generate go build -v -o ../bin/go-sumtype github.com/BurntSushi/go-sumtype
//go:generate go build -v -o ../bin/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint
//go:generate go build -v -o ../bin/go-consistent github.com/quasilyte/go-consistent
//go:generate go build -v -o ../bin/reviewdog github.com/reviewdog/reviewdog/cmd/reviewdog
//go:generate go build -v -o ../bin/benchstat golang.org/x/perf/cmd/benchstat
//go:generate go build -v -o ../bin/stringer golang.org/x/tools/cmd/stringer
//go:generate go build -v -o ../bin/gofumpt mvdan.cc/gofumpt
