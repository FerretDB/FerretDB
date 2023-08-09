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

// Package common provides common code for all handlers.
package common

import (
	"errors"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

const (
	// MinWireVersion is the minimal supported wire protocol version.
	MinWireVersion = int32(0) // needed for some apps and drivers

	// MaxWireVersion is the maximal supported wire protocol version.
	MaxWireVersion = int32(17)
)

// IsNotClientMetadata checks if the message does not contain client metadata.
func IsNotClientMetadata(msg *wire.OpMsg) error {
	document, err := msg.Document()
	if err != nil {
		return lazyerrors.Error(err)
	}

	if client, _ := document.Get("client"); client != nil {
		return errors.New("The client metadata document may only be sent in the first hello")
	}

	return nil
}
