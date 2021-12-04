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

package logging

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Setup(level zapcore.Level) {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(level)

	logger, err := config.Build()
	if err != nil {
		log.Fatal(err)
	}

	zap.ReplaceGlobals(logger)

	if _, err = zap.RedirectStdLogAt(logger, zap.InfoLevel); err != nil {
		log.Fatal(err)
	}
}
