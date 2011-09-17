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
	"log"
	"os"

	"github.com/kylelemons/goat/term"
)

func main() {
	tio, err := term.NewTermSettings(0)
	if err != nil {
		log.Fatalf("terminfo: %s", err)
	}
	if err := tio.Raw(); err != nil {
		log.Fatalf("rawterm: %s", err)
	}
	defer tio.Reset()

	ch := make([]byte, 10)
	tty := term.NewTTY(os.Stdin)
	line := ""
	for {
		n, err := tty.Read(ch)
		if err != nil {
			log.Printf("read: %s", err)
			return
		}
		switch str := string(ch[:n]); str {
		case "quit", term.Interrupt, term.EndOfFile:
			return
		case term.CarriageReturn, term.NewLine:
			log.Printf("read: %q\r", line)
			line = ""
		default:
			line += str
		}
	}
}
