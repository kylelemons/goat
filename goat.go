package main

import (
	"log"
	"os"

	"github.com/kylelemons/goat/term"
)

func main() {
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
		case "quit", "":
			return
		case "\r", "\n":
			log.Printf("read: %q\r", line)
			line = ""
		default:
			line += str
		}
	}
}
