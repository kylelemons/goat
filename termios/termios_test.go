// Copyright 2013 Google, Inc.  All rights reserved.
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

// Run this test with:
//   go test -c && ./term.test

package termios

import (
	"testing"
)

func TestTermSettings(t *testing.T) {
	tio, err := NewTermSettings(0)
	if err != nil {
		t.Fatalf("NewTermSettings: %s", err)
	}
	if err := tio.Apply(); err != nil {
		t.Errorf("Apply: %s", err)
	}
	if err := tio.Reset(); err != nil {
		t.Errorf("Reset: %s", err)
	}
	t.Log(tio)
}

func TestTermSize(t *testing.T) {
	tio, err := NewTermSettings(0)
	if err != nil {
		t.Fatalf("NewTermSettings: %s", err)
	}
	w, h, err := tio.GetSize()
	if err != nil {
		t.Fatalf("GetSize: %s", err)
	}
	t.Logf("Size: %d cols, %d rows", w, h)
}
