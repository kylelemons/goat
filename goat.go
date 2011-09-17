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
	"io"
	"log"
	"os"

	"github.com/kylelemons/goat/term"
)

func main() {
	// Set the terminal to RAW mode
	tio, err := term.NewTermSettings(0)
	if err != nil {
		log.Fatalf("terminfo: %s", err)
	}
	if err := tio.Raw(); err != nil {
		log.Fatalf("rawterm: %s", err)
	}

	// Restore cooked settings on exit
	defer tio.Reset()

	// Allocate a TTY connected to standard input
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
