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

//go:build ferretdb_bw

package main

/*
#include "postgres.h"
*/
import "C"

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/ferretdb"
	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
)

//export BackgroundWorkerMain
func BackgroundWorkerMain(args C.Datum) {
	log.SetPrefix("ferretdb: ")
	log.SetFlags(0)

	ctx, stop := ctxutil.SigTerm(context.Background())
	defer stop()

	f, err := ferretdb.New(&ferretdb.Config{
		PostgreSQLURL: "postgres://username:password@127.0.0.1:5432/postgres",
		ListenAddr:    "127.0.0.1:27017",
		StateDir:      ".",
		LogLevel:      slog.LevelDebug,
		LogOutput:     os.Stderr,
	})
	if err != nil {
		log.Fatal(err)
	}

	version.Get().Package = "extension"

	done := make(chan struct{})

	go func() {
		f.Run(ctx)
		close(done)
	}()

	uri := f.MongoDBURI()
	log.Printf("Running at %s", uri)

	<-done

	return
}

func main() {
	panic("not reached")
}
