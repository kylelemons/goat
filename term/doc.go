// Package term provides a basic terminal emulation framework.
//
// The TTY class abstracts the logic of providing line editing and line
// buffering, as well as escape sequence recognition.  When creating a TTY with
// the NewTTY function, interactive echo will automatically be enabled if the
// provided io.Reader's underlying object also implements io.Writer.
//
// Line editing capabilities
//
// The line editing facilities are very basic; you can type, and you can
// backspace out characters up to the beginning of the line.  Note that for all
// internal purposes, typing a control character (e.g. ^D or ^C) starts a new
// line, including for line history below.
//
// Line history
//
// Currently the TTY only has a single-line history.  Pressing the return key
// will save the current line in that history, and pressing the "up" arrow at
// any time will restore the previous line.
//
// Example
//
// The following example reads from standard input, calling runCommand with
// the complete lines it accumulates.
//
//   tty := term.NewTTY(os.Stdin)
//
//   line := ""
//   for {
//       n, err := tty.Read(raw)
//       if err != nil {
//           log.Printf("read: %s", err)
//           return
//       }
//
//       switch str := string(raw[:n]); str {
//       case "quit", term.Interrupt, term.EndOfFile:
//           fmt.Println("Goodbye!")
//           return
//       case term.CarriageReturn, term.NewLine:
//           runCommand(line)
//           line = ""
//       default:
//           line += str
//       }
//   }
//
// In order for the above example to work, the terminal must be in raw mode,
// which can be done by running your binary like so (on a unix-like operating
// system like linux or darwin):
//
//   stty raw; cmd; stty cooked
//
package term
