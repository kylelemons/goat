package term

import (
	"io"
	"os"
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
		Echo:   []string{"o", "n", "e", "\r\n", "\rone", "t", "w", "o", "\r\n"},
		Output: []string{"one", "\n", "onetwo", "\n"},
	},
	{
		Desc:   "late up",
		Chunks: []string{"one\ntwo\x1b[A\n"},
		Echo:   []string{"o", "n", "e", "\r\n", "t", "w", "o", "\rone", "\r\n"},
		Output: []string{"one", "\n", "one", "\n"},
	},
	{
		Desc:   "up up",
		Chunks: []string{"one\n\x1b[Atwo\x1b[Athree\n"},
		Echo: []string{
			"o", "n", "e", "\r\n",
			"\rone", "t", "w", "o",
			"\rone   \b\b\b",
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
}

// TestTerm test up to 1000 reads of up to 4096 bytes each per testcase.
func TestTerm(t *testing.T) {
	for _, test := range termTests {
		desc := test.Desc

		done := make(chan bool)
		checkRead := func(what string, r io.Reader, chunks []string) {
			raw := make([]byte, 4096)
			var idx int
			for idx = 0; idx < 1000; idx++ {
				n, err := r.Read(raw)
				if err == os.EOF {
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

		type rw struct {
			io.Reader
			io.Writer
		}

		userR, userW := io.Pipe()
		echoR, echoW := io.Pipe()

		tty := NewTTY(rw{userR, echoW})

		go checkRead("read", tty, test.Output)
		go checkRead("echo", echoR, test.Echo)
		for _, chunk := range test.Chunks {
			if _, err := io.WriteString(userW, chunk); err != nil {
				t.Errorf("%s: write(%q): %s", desc, chunk, err)
				return
			}
		}
		userW.Close()
		<-done

		echoW.Close()
		<-done
	}
}
