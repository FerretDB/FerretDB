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

//go:build unix

package ctxutil

import (
	"context"
	"os/signal"

	"golang.org/x/sys/unix"
)

// SigTerm returns a copy of the parent context that is marked done
// (its Done channel is closed) when termination signal arrives,
// when the returned stop function is called, or when the parent context's
// Done channel is closed, whichever happens first.
func SigTerm(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, unix.SIGTERM, unix.SIGINT)
}
