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

package trace

import (
	"context"
)

// executableTraceContextKey is a value key used to retrieve the ExecutableTrace
// from Context.
type executableTraceContextKey struct{}

// ExecutableTrace is a set of hooks used to trace the execution of binary
// executables. Any particular hook may be nil.
type ExecutableTrace struct {
	// ExecuteStart is called before the execution of the executable. The
	// executableName parameter is the name of the credential helper executable
	// used with NativeStore. The action parameter is one of "store", "get" and
	// "erase".
	//
	// Reference:
	//   - https://docs.docker.com/engine/reference/commandline/login#credentials-store
	ExecuteStart func(executableName string, action string)

	// ExecuteDone is called after the execution of an executable completes.
	// The executableName parameter is the name of the credential helper executable
	// used with NativeStore. The action parameter is one of "store", "get" and
	// "erase". The err parameter is the error (if any) returned from the execution.
	//
	// Reference:
	//   - https://docs.docker.com/engine/reference/commandline/login#credentials-store
	ExecuteDone func(executableName string, action string, err error)
}

// ContextExecutableTrace returns the ExecutableTrace associated with the context.
// If none, it returns nil.
func ContextExecutableTrace(ctx context.Context) *ExecutableTrace {
	trace, _ := ctx.Value(executableTraceContextKey{}).(*ExecutableTrace)
	return trace
}

// WithExecutableTrace takes a Context and an ExecutableTrace, and returns a Context with
// the ExecutableTrace added as a Value.
func WithExecutableTrace(ctx context.Context, trace *ExecutableTrace) context.Context {
	if trace == nil {
		return ctx
	}
	oldTrace, _ := ctx.Value(executableTraceContextKey{}).(*ExecutableTrace)
	if oldTrace != nil {
		trace.compose(oldTrace)
	}
	return context.WithValue(ctx, executableTraceContextKey{}, trace)
}

// compose takes an oldTrace and modifies the existing trace to include
// the hooks defined in the oldTrace.
func (trace *ExecutableTrace) compose(oldTrace *ExecutableTrace) {
	oldStart := oldTrace.ExecuteStart
	if oldStart != nil {
		start := trace.ExecuteStart
		trace.ExecuteStart = func(executableName, action string) {
			oldStart(executableName, action)
			start(executableName, action)
		}
	}
	oldDone := oldTrace.ExecuteDone
	if oldDone != nil {
		done := trace.ExecuteDone
		trace.ExecuteDone = func(executableName, action string, err error) {
			oldDone(executableName, action, err)
			done(executableName, action, err)
		}
	}
}
