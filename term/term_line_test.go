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

var termTests = []struct {
	Desc   string
	Chunks []string
	Echo   []string
	Output []string
}{
	{
		Desc:   "basic",
		Chunks: []string{"test"},
		Output: []string{"test"},
	},
	{
		Desc:   "lines",
		Chunks: []string{"one\ntwo"},
		Output: []string{"one", "\n", "two"},
	},
	{
		Desc:   "\\r\\n",
		Chunks: []string{"one\r\ntwo"},
		Output: []string{"one", "\r", "\n", "two"},
	},
	{
		Desc:   "echo",
		Chunks: []string{"o", "n", "e"},
		Echo:   []string{"o", "n", "e"},
	},
	{
		Desc:   "newline",
		Chunks: []string{"o", "n", "e", "\r", "\n"},
		Echo:   []string{"o", "n", "e", "\r\n", "\r\n"},
	},
	{
		Desc:   "word",
		Chunks: []string{"one\n"},
		Echo:   []string{"o", "n", "e", "\r\n"},
	},
	{
		Desc:   "backspace",
		Chunks: []string{"spee\bll"},
		Echo:   []string{"s", "p", "e", "e", "\b \b", "l", "l"},
		Output: []string{"spell"},
	},
	{
		Desc:   "bksp start",
		Chunks: []string{"\b\bbkx\bsp"},
		Echo:   []string{"b", "k", "x", "\b \b", "s", "p"},
		Output: []string{"bksp"},
	},
	{
		Desc:   "bksp lines",
		Chunks: []string{"\b\bbkx\bsp\ntext\b\bst"},
		Output: []string{"bksp", "\n", "test"},
	},
	{
		Desc:   "bksp start chars",
		Chunks: []string{"\b", "b", "k", "s", "p", "\n"},
		Output: []string{"bksp", "\n"},
	},
	{
		Desc:   "escape only",
		Chunks: []string{"\x1b"},
		Echo:   []string{}, // EOF before escape completes won't echo
		Output: []string{"\x1b"},
	},
	{
		Desc:   "escape non-CSI",
		Chunks: []string{"\x1b0"},
		Echo:   []string{"\x1b", "0"},
		Output: []string{"\x1b0"},
	},
	{
		Desc:   "escape embedded",
		Chunks: []string{"one\x1btwo"},
		Echo:   []string{"o", "n", "e", "\x1b", "t", "w", "o"},
		Output: []string{"one\x1btwo"},
	},
	{
		Desc:   "esc BS",
		Chunks: []string{"one\x1b\b\btwo"},
		Output: []string{"ontwo"},
	},
	{
		Desc:   "unknown seq",
		Chunks: []string{"\x1b[5G"}, // CHA[5]
		Echo:   []string{},          // Well-formed escapes, even unknown, aren't echoed
		Output: []string{"\x1b[5G"}, // but they are outputted
	},
	{
		Desc:   "unknown seq inline",
		Chunks: []string{"on\x1b[5Ge"},
		Echo:   []string{"o", "n", "e"},
		Output: []string{"on\x1b[5Ge"},
	},
	{
		Desc:   "up",
		Chunks: []string{"one\n\x1b[Atwo\n"},
		Echo: []string{
			"o", "n", "e", "\r\n",
			"one",
			"t", "w", "o", "\r\n",
		},
		Output: []string{"one", "\n", "onetwo", "\n"},
	},
	{
		Desc:   "zero up",
		Chunks: []string{"0\n\x1b[A1"},
		Echo:   []string{"0", "\r\n", "0", "1"},
		Output: []string{"0", "\n", "01"},
	},
	{
		Desc:   "up noop",
		Chunks: []string{"y\x1b[A", "x"},
		Echo:   []string{"y", "x"},
		Output: []string{"yx"},
	},
	{
		Desc:   "late up",
		Chunks: []string{"one\ntwo\x1b[A\n"},
		Echo: []string{
			"o", "n", "e", "\r\n",
			"t", "w", "o",
			"\b\b\bone", "\r\n",
		},
		Output: []string{"one", "\n", "one", "\n"},
	},
	{
		Desc:   "up up",
		Chunks: []string{"one\n\x1b[Atwo\x1b[Athree\n"},
		Echo: []string{
			"o", "n", "e", "\r\n",
			"one", "t", "w", "o",
			"\b\b\b\b\b\bone   \b\b\b",
			"t", "h", "r", "e", "e", "\r\n"},
		Output: []string{"one", "\n", "onethree", "\n"},
	},
	{
		Desc: "left",
		Chunks: []string{
			"abcde",
			"\x1b[D", // LEFT
		},
		Echo: []string{
			"a", "b", "c", "d", "e",
			"\x1b[D",
		},
		Output: []string{"abcde"},
	},
	{
		Desc: "left noop",
		Chunks: []string{
			"\x1b[D", // LEFT
			"abcde",
		},
		Echo: []string{
			"a", "b", "c", "d", "e",
		},
		Output: []string{"abcde"},
	},
	{
		Desc: "left insert",
		Chunks: []string{
			"abc",
			"\x1b[D", // LEFT
			"d",
		},
		Echo: []string{
			"a", "b", "c",
			"\x1b[D",
			"dc\b",
		},
		Output: []string{"abdc"},
	},
	{
		Desc: "left bksp",
		Chunks: []string{
			"abcd",
			"\x1b[D", // LEFT
			"\x1b[D", // LEFT
			"\b",
		},
		Echo: []string{
			"a", "b", "c", "d",
			"\x1b[D",
			"\x1b[D",
			"\bcd \b\b\b",
		},
		Output: []string{"acd"},
	},
	{
		Desc: "left noop insert",
		Chunks: []string{
			"a",
			"\x1b[D", // LEFT
			"\x1b[D", // LEFT
			"b",
		},
		Echo: []string{
			"a",
			"\x1b[D",
			"ba\b",
		},
		Output: []string{"ba"},
	},
	{
		Desc: "right noop",
		Chunks: []string{
			"abc",
			"\x1b[C", // RIGHT
		},
		Echo: []string{
			"a", "b", "c",
		},
		Output: []string{"abc"},
	},
	{
		Desc: "left right",
		Chunks: []string{
			"ab",
			"\x1b[D", // LEFT
			"\x1b[C", // RIGHT
			"c",
		},
		Echo: []string{
			"a", "b",
			"\x1b[D", // LEFT
			"\x1b[C", // RIGHT
			"c",
		},
		Output: []string{"abc"},
	},
	{
		Desc: "left right right",
		Chunks: []string{
			"01234",
			"\x1b[D", // LEFT
			"\x1b[D", // LEFT
			"\x1b[D", // LEFT
			"\x1b[C", // RIGHT
			"\x1b[C", // RIGHT
			"X",
		},
		Echo: []string{
			"0", "1", "2", "3", "4",
			"\x1b[D", // LEFT
			"\x1b[D", // LEFT
			"\x1b[D", // LEFT
			"\x1b[C", // RIGHT
			"\x1b[C", // RIGHT
			"X4\b",
		},
		Output: []string{"0123X4"},
	},
	{
		Desc: "left left down",
		Chunks: []string{
			"abc",
			"\x1b[D", // LEFT
			"\x1b[D", // LEFT
			"\x1b[B", // DOWN
		},
		Echo: []string{
			"a", "b", "c",
			"\x1b[D",
			"\x1b[D",
			"bc",
		},
		Output: []string{"abc"},
	},
	{
		Desc: "left up",
		Chunks: []string{
			"qwerty\nabc",
			"\x1b[D", // LEFT
			"\x1b[A", // UP
			"!",
		},
		Echo: []string{
			"q", "w", "e", "r", "t", "y", "\r\n",
			"a", "b", "c",
			"\x1b[D",
			"\b\bqwerty",
			"!",
		},
		Output: []string{"qwerty", "\n", "qwerty!"},
	},
}

// TestTerm test up to 1000 reads of up to 4096 bytes each per testcase.
func TestTerm(t *testing.T) {
	for _, test := range termTests {
		desc := test.Desc
		done := make(chan bool)
		pipe := NewDoublePipe()
		tty := NewTTY(pipe.Remote)

		go VerifyReads(t, desc, "read", tty, test.Output, done)
		go VerifyReads(t, desc, "echo", pipe.Local, test.Echo, done)

		for _, chunk := range test.Chunks {
			if _, err := io.WriteString(pipe.Local, chunk); err != nil {
				t.Errorf("%s: write(%q): %s", desc, chunk, err)
			}
		}

		pipe.Local.Close()
		<-done

		pipe.Remote.Close()
		<-done
	}
}
