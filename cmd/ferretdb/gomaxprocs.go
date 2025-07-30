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

// We should remove that code (and this dependency) with 1.25:
// https://tip.golang.org/doc/go1.25#container-aware-gomaxprocs
//
//go:build !go1.25

package main

import (
	"fmt"
	"log/slog"
	"math"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

func setGOMAXPROCS(logger *slog.Logger) {
	maxprocsOpts := []maxprocs.Option{
		maxprocs.Min(2),
		maxprocs.RoundQuotaFunc(func(v float64) int {
			return int(math.Ceil(v))
		}),
		maxprocs.Logger(func(format string, a ...any) {
			logger.Info(fmt.Sprintf(format, a...))
		}),
	}
	if _, err := maxprocs.Set(maxprocsOpts...); err != nil {
		logger.Warn("Failed to set GOMAXPROCS", logging.Error(err))
	}
}
