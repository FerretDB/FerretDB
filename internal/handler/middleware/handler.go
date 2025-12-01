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

package middleware

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// Handler is a common interface for [handler.Handler] and [proxy.Handler]
// (but not for [Middleware]).
type Handler interface {
	// Run runs the handler until ctx is canceled and all requests are processed.
	// When this method returns, the handler is fully stopped.
	//
	// It is handler's responsibility to wait for all in-progress requests to be processed,
	// and to return from the future calls to Handle with an error early.
	Run(ctx context.Context)

	// Handle processes a single request.
	//
	// The passed context is canceled when the client disconnects.
	// Canceling ctx should stop the processing of the request.
	// Handle should also exit early when ctx passed to Run is canceled.
	//
	// Response is a normal or error response produced by the handler.
	//
	// Error is returned when the handler cannot process the request;
	// for example, when connection with PostgreSQL or proxy is lost.
	// Returning an error generally means that the listener should
	// close the client connection without responding.
	// Error should not be [*mongoerrors.Error] or have that type in its chain.
	//
	// Exactly one of Response or error should be non-nil.
	Handle(ctx context.Context, req *Request) (resp *Response, err error)

	// Handler should expose its metrics, but not metrics of passed dependencies.
	prometheus.Collector
}
