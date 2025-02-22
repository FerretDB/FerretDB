package main

/*
#include "postgres.h"

extern void BackgroundWorkerMain(Datum args);
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

func main() {
	panic("not reached")
}

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
