package main

/*
#include "postgres.h"

extern void BackgroundWorkerMain(Datum args);
*/
import "C"

import (
	"context"
	"log"

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/ferretdb"
	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
)

func main() {
	panic("not reached")
}

//export BackgroundWorkerMain
func BackgroundWorkerMain(args C.Datum) {
	ctx, stop := ctxutil.SigTerm(context.Background())
	defer stop()

	log.Printf("FerretDB is starting")

	f, err := ferretdb.New(&ferretdb.Config{
		PostgreSQLURL: "postgres://username:password@127.0.0.1:5432/postgres",
		ListenAddr:    "127.0.0.1:27027",
		StateDir:      ".",
		// LogLevel:      slog.LevelDebug,
		// LogOut:        os.Stderr,
	})
	if err != nil {
		log.Print(err)
		return
	}

	version.Get().Package = "extension"

	done := make(chan struct{})

	go func() {
		f.Run(ctx)
		close(done)
	}()

	uri := f.MongoDBURI()
	log.Printf("FerretDB is running at %s", uri)

	<-done

	return
}
