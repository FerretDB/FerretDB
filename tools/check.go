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

// This file is invoked from tools.go or old.go and should work with any version of Go.
// Keep both old and new styles of build tags.

//go:build ignore
// +build ignore

package main

import (
	"flag"
	"log"
	"regexp"
	"runtime"
	"strconv"
)

func main() {
	log.SetFlags(0)

	oldF := flag.Bool("old", false, "")
	flag.Parse()

	if *oldF {
		log.Fatalf("FerretDB requires Go 1.20 or later.")
	}

	v := runtime.Version()
	re := regexp.MustCompile(`go1\.(\d+)`)
	m := re.FindStringSubmatch(v)
	if len(m) != 2 {
		log.Fatalf("Unexpected version %q.", v)
	}

	minor, err := strconv.Atoi(m[1])
	if err != nil {
		log.Fatalf("Unexpected version %q: %s.", v, err)
	}

	if minor < 18 {
		log.Fatalf("FerretDB requires Go 1.20 or later. The version of `go` binary in $PATH is %q.", v)
	}
}
