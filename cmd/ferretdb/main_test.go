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

////go:build ferretdb_testcover

package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/alecthomas/kong"
	"go.uber.org/zap/zapcore"
)

// TestCover allows us to run FerretDB with coverage enabled.
func TestCover(t *testing.T) {
	main()
}

// TestMain ensures that command-line flags are initialized correctly when FerretDB is run with coverage enabled.
func TestMain(m *testing.M) {
	levels := []string{
		zapcore.DebugLevel.String(),
		zapcore.InfoLevel.String(),
		zapcore.WarnLevel.String(),
		zapcore.ErrorLevel.String(),
	}

	kong.Parse(&cli,
		kong.Vars{
			"default_logLevel": zapcore.DebugLevel.String(),
			"default_mode":     string(clientconn.AllModes[0]),
			"help_handler":     "Backend handler: " + strings.Join(registry.Handlers(), ", "),
			"help_logLevel":    "Log level: " + strings.Join(levels, ", "),
			"help_mode":        fmt.Sprintf("Operation mode: %v", clientconn.AllModes),
		})

	os.Exit(m.Run())
}
