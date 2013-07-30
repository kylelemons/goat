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
	"reflect"
	"testing"
)

var growTests = []struct {
	Start  rect
	Dw, Dh int
	Expect rect
}{
	{
		rect{1, 2, 10, 5},
		1, 1,
		rect{0, 1, 12, 7},
	},
	{
		rect{1, 2, 10, 5},
		-1, -1,
		rect{2, 3, 8, 3},
	},
}

func TestGrow(t *testing.T) {
	for _, test := range growTests {
		grown := test.Start.grow(test.Dw, test.Dh)
		if got, want := grown, test.Expect; !reflect.DeepEqual(got, want) {
			t.Errorf("%v.grow(%d,%d) = %v, want %v",
				test.Start, test.Dw, test.Dh, got, want)
		}
	}
}

var frameTests = []struct {
	Desc   string
	Func   func(*Region)
	Input  []string
	Output []string
}{
	{
		"Empty region",
		func(r *Region) {
			r.SetSize(4, 3)
		},
		[]string{},
		[]string{
			"\x1b[1;1H", "    ",
			"\x1b[2;1H", "    ",
			"\x1b[3;1H", "    ",
			"\x1b[1;1H",
		},
	},
	{
		"Empty region, with border",
		func(r *Region) {
			r.SetSize(4, 3)
			r.SetBorder(SimpleBorder)
		},
		[]string{},
		[]string{
			"\x1b[1;1H", ",--.",
			"\x1b[2;1H", "|  |",
			"\x1b[3;1H", "`--'",
			"\x1b[2;2H",
		},
	},
}

func TestFrame(t *testing.T) {
	for _, test := range frameTests {
		desc := test.Desc
		done := make(chan bool)
		pipe := NewDoublePipe()
		tty, region := NewFrameTTY(pipe.Remote)

		go VerifyReads(t, desc, "read", tty, nil, done)
		go VerifyReads(t, desc, "echo", pipe.Local, test.Output, done)

		go func() {
			test.Func(region)
			region.Draw()
			done <- true
		}()

		for _, chunk := range test.Input {
			if _, err := io.WriteString(pipe.Local, chunk); err != nil {
				t.Errorf("%s: write(%q): %s", desc, chunk, err)
			}
		}

		<-done

		pipe.Local.Close()
		<-done

		pipe.Remote.Close()
		<-done
	}
}
