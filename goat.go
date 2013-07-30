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

// goat
//
// It is a basic example of terminal emulation with the "goat/term" package.
// It reads chunks in and writes them to standard output.  Try typing a line
// and then hitting the up key on the next line.  Try editing a previous line
// and hitting the up key again.
//
// Press ^C, ^D, or type "quit" to exit.
//
// If something happens and you can't exit, try "killall goat" from another
// terminal; this shouldn't happen, but it's possible.
package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/kylelemons/goat/term"
	"github.com/kylelemons/goat/termios"
)

var frame = flag.Bool("frame", false, "Do a frame demo instead of line editing")

func main() {
	flag.Parse()

	tio, err := termios.NewTermSettings(0)
	if err != nil {
		log.Fatalf("terminal: %s", err)
	}
	if err := tio.Raw(); err != nil {
		log.Fatalf("rawterm: %s", err)
	}
	defer tio.Reset()

	if *frame {
		frameDemo(tio)
	} else {
		lineDemo()
	}
}

func lineDemo() {
	tty := term.NewTTY(os.Stdin)

	// Prompt after each newline
	prompt := func() {
		io.WriteString(tty, "> ")
	}
	prompt()

	// Allocate the line buffer and accumulator
	linebuf := make([]byte, 128)
	line := ""

	for {
		// Read from the TTY
		n, err := tty.Read(linebuf)
		if err != nil {
			log.Printf("read: %s", err)
			return
		}

		// Examine the chunk
		switch str := string(linebuf[:n]); str {
		case "quit", term.Interrupt, term.EndOfFile:
			// Quit on "quit", ^C, and ^D
			io.WriteString(os.Stdout, "Goodbye!\r\n")
			return
		case term.CarriageReturn, term.NewLine:
			// Print out lines
			log.Printf("read: %q\r\n", line)
			prompt()
			line = ""
		default:
			// Accumulate lines
			line += str
		}
	}
}

func frameDemo(tio *termios.TermSettings) {
	// Allocate a TTY connected to standard input
	tty, region := term.NewFrameTTY(os.Stdin)
	tty.Clear()
	region.SetBorder(term.SimpleBorder)

	width, height, err := tio.GetSize()
	if err == nil && width > 0 && height > 0 {
		region.SetSize(width, height)
	}

	region.Draw()

	// Allocate the line buffer and accumulator
	linebuf := make([]byte, 128)

	for {
		// Read from the TTY
		n, err := tty.Read(linebuf)
		if err != nil {
			log.Printf("read: %s", err)
			return
		}

		// Examine the chunk
		switch str := string(linebuf[:n]); str {
		case "quit", term.Interrupt, term.EndOfFile:
			// Quit on "quit", ^C, and ^D
			tty.Clear()
			tty.SetCursor(0, 0)
			io.WriteString(tty, "Goodbye!\r\n")
			log.Printf("%dx%d\r", width, height)
			return
		}
	}

}
