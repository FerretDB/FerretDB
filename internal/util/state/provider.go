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

package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// Provider provides access to FerretDB process state.
type Provider struct {
	filename string

	m sync.Mutex
	s *State
}

// NewProvider creates a new Provider that stores state in the given file.
//
// If filename is empty, then the state is not persisted.
func NewProvider(filename string) (*Provider, error) {
	p := &Provider{
		filename: filename,
	}

	if _, err := p.Get(); err != nil {
		return nil, err
	}

	return p, nil
}

// If addUUIDToMetric is true, then the UUID is added to the Prometheus metric.
func (p *Provider) MetricsCollector(addUUIDToMetric bool) prometheus.Collector {
	return newMetricsCollector(p, addUUIDToMetric)
}

// Get returns a copy of the current process state.
//
// It is okay to call this function often.
// The caller should not cache result; Provider does everything needed itself.
func (p *Provider) Get() (*State, error) {
	p.m.Lock()
	defer p.m.Unlock()

	if p.s != nil {
		return p.s.deepCopy(), nil
	}

	p.s = new(State)

	// use defaults if state is not persisted
	if p.filename == "" {
		p.s.fill()
		return p.s.deepCopy(), nil
	}

	// Simply read and overwrite state to handle all errors and edge cases
	// like missing directory, corrupted file, invalid UUID, etc.

	b, _ := os.ReadFile(p.filename)
	_ = json.Unmarshal(b, p.s)

	p.s.fill()

	b, err := json.Marshal(p.s)

	if err == nil {
		_ = os.MkdirAll(filepath.Dir(p.filename), 0o777)
		err = os.WriteFile(p.filename, b, 0o666)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to persist state: %w", err)
	}

	return p.s.deepCopy(), nil
}
