// Copyright 2018 Palantir Technologies, Inc.
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

package baseapp

import (
	"context"
	"net/http"

	"github.com/palantir/go-baseapp/pkg/errfmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// httpError represents any error that presents itself as an HTTP error with a
// status code.
type httpError interface {
	StatusCode() int
}

// RichErrorMarshalFunc is a zerolog error marshaller that formats the error as
// a string that includes a stack trace, if one is available.
func RichErrorMarshalFunc(err error) interface{} {
	return errfmt.Print(err)
}

// HandleRouteError is a hatpear error handler that logs the error and sends
// an error response to the client. If the error has a `StatusCode` function
// this will be called and converted to an appropriate HTTP status code error.
func HandleRouteError(w http.ResponseWriter, r *http.Request, err error) {
	var log *zerolog.Event
	// Either the deadline has passed or the request was canceled
	// 499 is an NGINX style response code for 'Client Closed Connection'
	// and is a non-standard, but widely used, HTTP status code
	if cerr := r.Context().Err(); cerr == context.Canceled {
		log = hlog.FromRequest(r).Debug()
		WriteJSON(w, 499, map[string]string{
			"error": "Client Closed Connection",
		})
	} else {
		log = hlog.FromRequest(r).Error().Err(err)

		cause := errors.Cause(err)
		statusCode := http.StatusInternalServerError
		if aerr, ok := cause.(httpError); ok {
			statusCode = aerr.StatusCode()
		}

		rid, _ := hlog.IDFromRequest(r)
		WriteJSON(w, statusCode, map[string]string{
			"error":      http.StatusText(statusCode),
			"request_id": rid.String(),
		})
	}

	log.Str("method", r.Method).
		Str("path", r.URL.String()).
		Msg("Unhandled error while serving route")
}
