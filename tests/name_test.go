// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package tests

import (
	keepgo "github.com/faelmori/keepgo/internal"
	"runtime"
	"strings"
	"testing"
)

func TestPlatformName(t *testing.T) {
	got := keepgo.Platform()
	t.Logf("Platform is %v", got)
	wantPrefix := runtime.GOOS + "-"
	if !strings.HasPrefix(got, wantPrefix) {
		t.Errorf("Platform() want: /^%s.*$/, got: %s", wantPrefix, got)
	}
}
