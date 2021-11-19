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

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/MangoDB-io/MangoDB/internal/clientconn"
	"github.com/MangoDB-io/MangoDB/internal/pg"
	"github.com/MangoDB-io/MangoDB/internal/util/debug"
	"github.com/MangoDB-io/MangoDB/internal/util/logging"
)

var composeBin string

func runCompose(args []string, stdin io.Reader) {
	cmd := exec.Command(composeBin, args...)
	log.Printf("%#v", cmd.Args)

	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	logger := logging.Setup(zap.InfoLevel).Sugar()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go debug.RunHandler(ctx, "127.0.0.1:8089", logger.Named("debug").Desugar())

	var err error
	if composeBin, err = exec.LookPath("docker-compose"); err != nil {
		logger.Fatal(err)
	}

	args := strings.Split(`exec -T postgres psql -U postgres -d mangodb`, " ")
	stdin := strings.NewReader(strings.Join([]string{
		`ALTER SCHEMA public RENAME TO pagila;`,
		`CREATE SCHEMA monila;`,
		`CREATE SCHEMA test;`,
	}, "\n"))
	runCompose(args, stdin)

	pgPool, err := pg.NewPool("postgres://postgres@127.0.0.1:5432/mangodb", logger.Desugar(), true)
	if err != nil {
		logger.Fatal(err)
	}

	// listen on all interfaces to make mongoimport below work from inside Docker
	addr := ":27018"
	if runtime.GOOS == "darwin" {
		// do not trigger macOS firewall; it works with Docker Desktop
		addr = "127.0.0.1:27018"
	}

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		ListenAddr: addr,
		Mode:       "normal",
		PgPool:     pgPool,
		Logger:     logger.Named("listener").Desugar(),
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		l.Run(ctx)
	}()

	var wg sync.WaitGroup
	for _, c := range []string{
		"actor",
		"address",
		"category",
		"city",
		"country",
		"customer",
		"film_actor",
		"film_category",
		"film",
		"inventory",
		"language",
		"rental",
		"staff",
		"store",
	} {
		l := fmt.Sprintf(
			`exec -T mongodb mongoimport --uri mongodb://host.docker.internal:27018/monila `+
				`--drop --maintainInsertionOrder --collection %[1]s /docker-entrypoint-initdb.d/%[1]s.json`,
			c,
		)

		wg.Add(1)
		// go func() {
		runCompose(strings.Split(l, " "), nil)
		wg.Done()
		// }()
	}

	wg.Wait()

	cancel()
	<-done
}
