/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Deprecated: This package is deprecated.
// The same functionality is now provided by [oras-go].
//
// [oras-go]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials/trace
package trace

import (
	"context"

	orastrace "oras.land/oras-go/v2/registry/remote/credentials/trace"
)

// executableTraceContextKey is a value key used to retrieve the ExecutableTrace
// from Context.
type executableTraceContextKey struct{}

// ExecutableTrace is a set of hooks used to trace the execution of binary
// executables. Any particular hook may be nil.
//
// Deprecated: This type is deprecated.
// The same functionality is now provided by oras-go.
type ExecutableTrace = orastrace.ExecutableTrace

// ContextExecutableTrace returns the ExecutableTrace associated with the
// context. If none, it returns nil.
//
// Deprecated: This type is deprecated.
// The same functionality is now provided by oras-go.
func ContextExecutableTrace(ctx context.Context) *ExecutableTrace {
	return orastrace.ContextExecutableTrace(ctx)
}

// WithExecutableTrace takes a Context and an ExecutableTrace, and returns a
// Context with the ExecutableTrace added as a Value. If the Context has a
// previously added trace, the hooks defined in the new trace will be added
// in addition to the previous ones. The recent hooks will be called first.
//
// Deprecated: This type is deprecated.
// The same functionality is now provided by oras-go.
func WithExecutableTrace(ctx context.Context, trace *ExecutableTrace) context.Context {
	return orastrace.WithExecutableTrace(ctx, trace)
}
