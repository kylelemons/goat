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

// Package term provides a basic terminal emulation framework.
//
// The TTY class abstracts the logic of providing line editing and line
// buffering, as well as escape sequence recognition.  When creating a TTY with
// the NewTTY function, interactive echo will automatically be enabled if the
// provided io.Reader's underlying object also implements io.Writer.
//
// Line editing capabilities (Line mode)
//
// The line editing facilities are very basic; you can type, and you can
// backspace out characters up to the beginning of the line.  Note that for all
// internal purposes, typing a control character (e.g. ^D or ^C) starts a new
// line, including for line history below.
//
// You can also use the arrow keys for editing:
//   LEFT   Move back one character
//   RIGHT  Move forward one character
//   DOWN   Move to the end of the line
//   UP     Restore previous line (see below)
//
// Line history (Line mode)
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
