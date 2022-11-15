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

package telemetry

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewReporterLock(t *testing.T) {
	t.Parallel()

	provider, err := state.NewProvider("")
	require.NoError(t, err)

	opts := NewReporterOpts{
		F: &Flag{
			v: pointer.ToBool(true),
		},
		ConnMetrics: connmetrics.NewListenerMetrics().ConnMetrics,
		P:           provider,
		L:           zap.L(),
	}

	_, err = NewReporter(&opts)
	assert.NoError(t, err)

	assert.True(t, provider.Get().TelemetryLocked)
}
