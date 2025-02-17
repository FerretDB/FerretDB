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
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	_ "golang.org/x/crypto/x509roots/fallback" // register root TLS certificates for production Docker image

	"github.com/FerretDB/FerretDB/v2/ferretdb"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn"
	"github.com/FerretDB/FerretDB/v2/internal/util/telemetry"
)

// The cli struct represents all command-line commands, fields and flags.
// It's used for parsing the user input.
//
// Keep order in sync with documentation.
var cli struct {
	// We hide `run` command to show only `ping` in the help message.
	Run  struct{} `cmd:"" default:"1"                             hidden:""`
	Ping struct{} `cmd:"" help:"Ping existing FerretDB instance."`

	Version bool `default:"false" help:"Print version to stdout and exit." env:"-"`

	PostgreSQLURL string `name:"postgresql-url" default:"postgres://127.0.0.1:5432/postgres" help:"PostgreSQL URL."`

	Listen struct {
		Addr        string `default:"127.0.0.1:27017" help:"Listen TCP address for MongoDB protocol."`
		Unix        string `default:""                help:"Listen Unix domain socket path for MongoDB protocol."`
		TLS         string `default:""                help:"Listen TLS address for MongoDB protocol."`
		TLSCertFile string `default:""                help:"TLS cert file path."`
		TLSKeyFile  string `default:""                help:"TLS key file path."`
		TLSCaFile   string `default:""                help:"TLS CA file path."`
		DataAPIAddr string `default:""                help:"Listen TCP address for HTTP Data API."`
	} `embed:"" prefix:"listen-"`

	Proxy struct {
		Addr        string `default:"" help:"Proxy address."`
		TLSCertFile string `default:"" help:"Proxy TLS cert file path."`
		TLSKeyFile  string `default:"" help:"Proxy TLS key file path."`
		TLSCaFile   string `default:"" help:"Proxy TLS CA file path."`
	} `embed:"" prefix:"proxy-"`

	DebugAddr string `default:"127.0.0.1:8088" help:"Listen address for HTTP handlers for metrics, pprof, etc."`

	Mode     string `default:"${default_mode}" help:"${help_mode}"                           enum:"${enum_mode}"`
	StateDir string `default:"."               help:"Process state directory."`
	Auth     bool   `default:"true"            help:"Enable authentication (on by default)." negatable:""`

	Log struct {
		Level  string `default:"${default_log_level}" help:"${help_log_level}"`
		Format string `default:"console"              help:"${help_log_format}"                     enum:"${enum_log_format}"`
		UUID   bool   `default:"false"                help:"Add instance UUID to all log messages." negatable:""`
	} `embed:"" prefix:"log-"`

	MetricsUUID bool `default:"false" help:"Add instance UUID to all metrics." negatable:""`

	OTel struct {
		Traces struct {
			URL string `default:"" help:"OpenTelemetry OTLP/HTTP traces endpoint URL (e.g. 'http://host:4318/v1/traces')."`
		} `embed:"" prefix:"traces-"`
	} `embed:"" prefix:"otel-"`

	Telemetry telemetry.Flag `default:"undecided" help:"${help_telemetry}"`

	Dev struct {
		ReplSetName string `default:"" help:"Replica set name."`

		RecordsDir string `hidden:""`

		Telemetry struct {
			URL            string        `default:"https://beacon.ferretdb.com/" hidden:""`
			UndecidedDelay time.Duration `default:"1h"                           hidden:""`
			ReportInterval time.Duration `default:"24h"                          hidden:""`
			Package        string        `default:""                             hidden:""`
		} `embed:"" prefix:"telemetry-"`
	} `embed:"" prefix:"dev-"`
}

// Additional variables for [kong.Parse].
var (
	logLevels = []string{
		slog.LevelDebug.String(),
		slog.LevelInfo.String(),
		slog.LevelWarn.String(),
		slog.LevelError.String(),
	}

	logFormats = []string{"console", "text", "json"}

	kongOptions = []kong.Option{
		kong.Vars{
			"default_log_level": ferretdb.DefaultLogLevel().String(),
			"default_mode":      clientconn.AllModes[0],

			"enum_log_format": strings.Join(logFormats, ","),
			"enum_mode":       strings.Join(clientconn.AllModes, ","),

			"help_log_format": fmt.Sprintf("Log format: '%s'.", strings.Join(logFormats, "', '")),
			"help_log_level":  fmt.Sprintf("Log level: '%s'.", strings.Join(logLevels, "', '")),
			"help_mode":       fmt.Sprintf("Operation mode: '%s'.", strings.Join(clientconn.AllModes, "', '")),
			"help_telemetry":  "Enable or disable basic telemetry reporting. See https://beacon.ferretdb.com.",
		},
		kong.DefaultEnvars("FERRETDB"),
	}
)

func main() {
	ctx := kong.Parse(&cli, kongOptions...)

	opts := &ferretdb.RunOpts{
		Version:       cli.Version,
		PostgreSQLURL: cli.PostgreSQLURL,
		Listen: ferretdb.ListenOpts{
			Addr:        cli.Listen.Addr,
			Unix:        cli.Listen.Unix,
			TLS:         cli.Listen.TLS,
			TLSCertFile: cli.Listen.TLSCertFile,
			TLSKeyFile:  cli.Listen.TLSKeyFile,
			TLSCaFile:   cli.Listen.TLSCaFile,
			DataAPIAddr: cli.Listen.DataAPIAddr,
		},
		Proxy: ferretdb.ProxyOpts{
			Addr:        cli.Proxy.Addr,
			TLSCertFile: cli.Proxy.TLSCertFile,
			TLSKeyFile:  cli.Proxy.TLSKeyFile,
			TLSCaFile:   cli.Proxy.TLSCaFile,
		},
		DebugAddr: cli.DebugAddr,
		Mode:      cli.Mode,
		StateDir:  cli.StateDir,
		Auth:      cli.Auth,
		Log: ferretdb.LogOpts{
			Level:  cli.Log.Level,
			Format: cli.Log.Format,
			UUID:   cli.Log.UUID,
		},
		MetricsUUID: cli.MetricsUUID,
		OTel: ferretdb.OTelOpts{
			Traces: ferretdb.OTelTracesOpts{
				URL: cli.OTel.Traces.URL,
			},
		},
		Telemetry: cli.Telemetry,
		Dev: ferretdb.DevOpts{
			ReplSetName: cli.Dev.ReplSetName,
			RecordsDir:  cli.Dev.RecordsDir,
			Telemetry: ferretdb.TelemetryOpts{
				URL:            cli.Dev.Telemetry.URL,
				UndecidedDelay: cli.Dev.Telemetry.UndecidedDelay,
				ReportInterval: cli.Dev.Telemetry.ReportInterval,
				Package:        cli.Dev.Telemetry.Package,
			},
		},
	}

	switch ctx.Command() {
	case "run":
		ferretdb.Run(context.Background(), opts)

	case "ping":
		if !ferretdb.Ping(context.Background(), opts) {
			os.Exit(1)
		}

	default:
		panic("unknown sub-command")
	}
}
