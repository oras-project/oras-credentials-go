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
// The same functionality is now provided by [oras.land/oras-go/v2/registry/remote/credentials/trace].
//
// [oras.land/oras-go/v2/registry/remote/credentials/trace]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials/trace
package trace

import (
	"context"

	trace "oras.land/oras-go/v2/registry/remote/credentials/trace"
)

// ExecutableTrace is a set of hooks used to trace the execution of binary
// executables. Any particular hook may be nil.
//
// Deprecated: This type is now simply [trace.ExecutableTrace] of oras-go.
//
// [trace.ExecutableTrace]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials/trace#ExecutableTrace
type ExecutableTrace = trace.ExecutableTrace

// ContextExecutableTrace returns the ExecutableTrace associated with the
// context. If none, it returns nil.
//
// Deprecated: This function now simply calls [trace.ContextExecutableTrace] of oras-go.
//
// [trace.ContextExecutableTrace]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials/trace#ContextExecutableTrace
func ContextExecutableTrace(ctx context.Context) *ExecutableTrace {
	return trace.ContextExecutableTrace(ctx)
}

// WithExecutableTrace takes a Context and an ExecutableTrace, and returns a
// Context with the ExecutableTrace added as a Value. If the Context has a
// previously added trace, the hooks defined in the new trace will be added
// in addition to the previous ones. The recent hooks will be called first.
//
// Deprecated: This function now simply calls [trace.WithExecutableTrace] of oras-go.
//
// [trace.WithExecutableTrace]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials/trace#WithExecutableTrace
func WithExecutableTrace(ctx context.Context, et *ExecutableTrace) context.Context {
	return trace.WithExecutableTrace(ctx, et)
}
