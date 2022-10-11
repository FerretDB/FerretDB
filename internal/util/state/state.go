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

// Package state stores FerretDB process state.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// State represents FerretDB process state.
type State struct {
	UUID string `json:"uuid"`
}

// Provider provides access to FerretDB process state.
type Provider struct {
	filename string

	rw sync.RWMutex
	s  State
}

// NewProvider creates a new Provider that stores state in the given file.
func NewProvider(filename string) (*Provider, error) {
	p := &Provider{
		filename: filename,
	}

	if _, err := p.Get(); err != nil {
		return nil, err
	}

	return p, nil
}

// Get returns the current process state.
//
// It is okay to call this function often.
// The caller should not cache result; Provider does everything needed itself.
func (p *Provider) Get() (*State, error) {
	// return different copies to each caller
	p.rw.RLock()
	s := p.s
	p.rw.RUnlock()

	if s.UUID != "" {
		return &s, nil
	}

	b, _ := os.ReadFile(p.filename)
	_ = json.Unmarshal(b, &s)
	_, err := uuid.Parse(s.UUID)

	if err == nil {
		// store a copy
		p.rw.Lock()
		p.s = s
		p.rw.Unlock()

		return &s, nil
	}

	// all errors (missing file, invalid file permission, invalid JSON, etc)
	// are handled in the same way - by regenerating state

	s.UUID = must.NotFail(uuid.NewRandom()).String()
	b = must.NotFail(json.Marshal(s))

	if err := os.MkdirAll(filepath.Dir(p.filename), 0o777); err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	if err := os.WriteFile(p.filename, b, 0o666); err != nil {
		return nil, fmt.Errorf("failed to write state file: %w", err)
	}

	// store a copy
	p.rw.Lock()
	p.s = s
	p.rw.Unlock()

	return &s, nil
}
