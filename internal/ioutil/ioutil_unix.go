//go:build !windows
// +build !windows

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

package ioutil

import (
	"fmt"
	"os"
	"syscall"
)

// CopyFilePermissions copies file ownership and permissions from "src" to "dst".
// Reference: https://github.com/docker/cli/blob/v24.0.0-beta.1/cli/config/configfile/file_unix.go#L11-L36
func CopyFilePermissions(src, dst string) error {
	mode := os.FileMode(0600)

	fi, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			// no ops
			return nil
		}
		return fmt.Errorf("failed to stat %s: %w", src, err)
	}
	if fi.Mode().IsRegular() {
		mode = fi.Mode()
	}
	if err := os.Chmod(dst, mode); err != nil {
		return fmt.Errorf("failed to chmod for %s: %w", dst, err)
	}

	uid := int(fi.Sys().(*syscall.Stat_t).Uid)
	gid := int(fi.Sys().(*syscall.Stat_t).Gid)
	if uid > 0 && gid > 0 {
		if err := os.Chown(dst, uid, gid); err != nil {
			return fmt.Errorf("failed to chown for %s: %w", dst, err)
		}
	}
	return nil
}
