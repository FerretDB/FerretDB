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

// Package conninfo provides access to connection-specific information.
package conninfo

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sync"

	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// contextKey is a named unexported type for the safe use of context.WithValue.
type contextKey struct{}

// Context key for WithConnInfo/Get.
var connInfoKey = contextKey{}

// ConnInfo represents connection info.
type ConnInfo struct {
	PeerAddr string

	token *resource.Token

	rw       sync.RWMutex
	username string
	password string

	clientMetadata ClientMetadata
}

type ClientMetadata struct {
	K           []string    `json:"$k"`
	Application Application `json:"application"`
	Driver      Driver      `json:"driver"`
	Platform    string      `json:"platform"`
	Os          Os          `json:"os"`
}

type Application struct {
	K    []string `json:"$k"`
	Name string   `json:"name"`
}

type Driver struct {
	K       []string `json:"$k"`
	Name    string   `json:"name"`
	Version string   `json:"version"`
}

type Os struct {
	K            []string `json:"$k"`
	Name         string   `json:"name"`
	Architecture string   `json:"architecture"`
	Version      string   `json:"version"`
	Type         string   `json:"type"`
}

// NewConnInfo return a new ConnInfo.
func NewConnInfo() *ConnInfo {
	connInfo := &ConnInfo{
		token: resource.NewToken(),
	}
	resource.Track(connInfo, connInfo.token)

	return connInfo
}

// Close frees resources.
func (connInfo *ConnInfo) Close() {
	resource.Untrack(connInfo, connInfo.token)
}

// Auth returns stored username and password.
func (connInfo *ConnInfo) Auth() (username, password string) {
	connInfo.rw.RLock()
	defer connInfo.rw.RUnlock()

	return connInfo.username, connInfo.password
}

// SetAuth stores username and password.
func (connInfo *ConnInfo) SetAuth(username, password string) {
	connInfo.rw.Lock()
	defer connInfo.rw.Unlock()

	connInfo.username = username
	connInfo.password = password
}

// SetClientMetadata sets the client metadata.
func (connInfo *ConnInfo) SetClientMetadata(metadata any) error {
	connInfo.rw.Lock()
	defer connInfo.rw.Unlock()

	metadataBytes, ok := metadata.(string)
	if !ok {
		return errors.New("failed converting the client's metadata")
	}

	if err := json.Unmarshal([]byte(metadataBytes), &connInfo.clientMetadata); err != nil {
		return err
	}
	return nil
}

// IsClientMetadataSet checks if the client metadata is set.
func (connInfo *ConnInfo) IsClientMetadataSet() bool {
	return !reflect.DeepEqual(connInfo.clientMetadata, ClientMetadata{})
}

// WithConnInfo returns a new context with the given ConnInfo.
func WithConnInfo(ctx context.Context, connInfo *ConnInfo) context.Context {
	return context.WithValue(ctx, connInfoKey, connInfo)
}

// Get returns the ConnInfo value stored in ctx.
func Get(ctx context.Context) *ConnInfo {
	value := ctx.Value(connInfoKey)
	if value == nil {
		panic("connInfo is not set in context")
	}

	connInfo, ok := value.(*ConnInfo)
	if !ok {
		panic("connInfo is set in context with wrong value type")
	}

	if connInfo == nil {
		panic("connInfo is set in context with nil value")
	}

	return connInfo
}
