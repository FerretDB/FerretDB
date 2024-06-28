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

// Package debug provides debug facilities.
package debug

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRunHandler(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})

	a, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoError(t, err)

	l, err := net.ListenTCP("tcp", a)
	require.NoError(t, err)

	addr := l.Addr().(*net.TCPAddr)

	require.NoError(t, l.Close())

	go RunHandler(context.TODO(), addr.String(), prometheus.NewRegistry(), zap.L(), started)

	time.Sleep(5 * time.Second)

	res, err := http.Get("http://" + addr.String() + "/debug/started")
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	close(started)

	res, err = http.Get("http://" + addr.String() + "/debug/started")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	res, err = http.Get("http://" + addr.String() + "/debug/started")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}
