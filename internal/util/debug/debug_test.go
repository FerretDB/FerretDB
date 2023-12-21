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

package debug

import (
	"context"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/state"
)

type MockCollector struct {
	mock.Mock
}

func (mc *MockCollector) Register(prometheus.Collector) error {
	return nil
}

func (mc *MockCollector) MustRegister(...prometheus.Collector) {
}

func (mc *MockCollector) Unregister(prometheus.Collector) bool {
	return true
}

type MockZapOption struct {
	mock.Mock
}

func (mc *MockCollector) apply(l zap.Logger) {
}

func TestRunHandler(t *testing.T) {
	t.Parallel()

	host := "127.0.0.1:5454"

	ctx := context.Background()

	filename := filepath.Join(t.TempDir(), "state.json")
	stateProvider, err := state.NewProvider(filename)
	require.NoError(t, err)

	metricsRegisterer := prometheus.DefaultRegisterer
	metricsProvider := stateProvider.MetricsCollector(true)
	metricsRegisterer.MustRegister(metricsProvider)

	l := zap.S()

	RunHandler(ctx, host, metricsRegisterer, l.Named("debug").Desugar())

	var u url.URL
	u.Path = "debug/archive"
	u.Host = host
	u.Scheme = "http"

	stream, err := performRequest(u)
	require.NoError(t, err)

	t.Log("printing bytes", stream)
	return
}
