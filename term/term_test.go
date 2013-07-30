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

package term

import (
	"io"
	"testing"
)

type RW struct {
	*io.PipeReader
	*io.PipeWriter
}

func (rw *RW) Close() error {
	return rw.PipeWriter.Close()
}

func (rw *RW) CloseWithError(err error) error {
	return rw.PipeWriter.CloseWithError(err)
}

type DoublePipe struct {
	Local  *RW
	Remote *RW
}

func NewDoublePipe() *DoublePipe {
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	return &DoublePipe{
		Local:  &RW{inR, outW},
		Remote: &RW{outR, inW},
	}
}

func VerifyReads(t *testing.T, desc, what string, r io.Reader, chunks []string, done chan bool) {
	raw := make([]byte, 4096)
	var idx int
	for idx = 0; idx < 1000; idx++ {
		n, err := r.Read(raw)
		if err == io.EOF {
			break
		} else if err != nil {
			t.Errorf("%s: %s[%d]: %s", desc, what, idx, err)
			continue
		}
		if chunks == nil {
			continue
		}
		if idx >= len(chunks) {
			t.Errorf("%s: extra %s: %q", desc, what, string(raw[:n]))
			continue
		}
		if got, want := string(raw[:n]), chunks[idx]; got != want {
			t.Errorf("%s: %s[%d] = %q, want %q", desc, what, idx, got, want)
		}
	}
	for idx < len(chunks) {
		t.Errorf("%s: missing %s: %q", desc, what, chunks[idx])
		idx++
	}
	done <- true
}
