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

// hpush (history push) stores the line for later reuse if it
// is not an escape sequence and contains characters.
//
// Side effects: (only if output is nonzero and not an escape sequence)
// - t.last will contain a copy of output
func (t *TTY) hpush() {
	if len(t.output) == 0 || t.output[0] < 32 {
		return
	}
	t.last = make([]byte, len(t.output))
	copy(t.last, t.output)
}

// hprev (history previous) replaces the current output with the last
// saved line (unless no line has been saved).
//
// To echo the new line, the following is written:
//   <home><line><spaces><backspaces>
// Where <line> is the new output <spaces> and <backspaces> are present if the
// previous line was long enough to require them to not leave dangling letters,
// and <home> is enough backspace characters to get to the beginning of the
// current line of text.
//
// Preconditions:
// - Must be called within an escape sequence
// Side effects:
// - t.output will contain a copy of t.last or will contain preescape
// - t.preescape will be nil
func (t *TTY) hprev() {
	if len(t.last) == 0 {
		t.output = t.preescape
		t.preescape = nil
		return
	}

	t.output = make([]byte, len(t.last))
	copy(t.output, t.last)

	width := len(t.preescape)
	t.preescape = nil

	home := width
	if t.linepos >= 0 {
		home = t.linepos
	}
	t.linepos = -1

	if t.screen != nil {
		size, delta := home+len(t.output), width-len(t.output)
		if delta > 0 {
			size += 2 * delta
		}
		overwrite := make([]byte, size)
		for i := 0; i < home; i++ {
			overwrite[i] = '\b'
		}
		copy(overwrite[home:], t.output)
		for i := len(t.output); i < width; i++ {
			overwrite[home+i] = ' '
			overwrite[home+i+delta] = '\b'
		}
		t.echo(overwrite...)
	}
}

// linechar processes the next character of input in line mode.
//
// If ch is ESC, it begins a new escape sequence by storing the current output
// into preescape and creating a new 8-cap byte slice for the escape sequence.
//
// If ch is a low nonprinting character, the current output is written and then
// the control character is written by itself.  This is to allow easy detection
// of things like ^C and ^D.
//
// If ch is BS (and there are characters in output), the length of output is
// shortened by one and a "\b \b" sequence is echoed to blank the space on the
// console.
//
// If ch is carriage return or newline (some terminals emit one, some emit the
// other), the output is written and then a the character is written, but in
// both cases a CRLF is echoed.
//
// If ch is anything else (basicaly a printing character), it is echoed and
// appended to output.
//
// Side Effects (possible):
// - t.preescape points to a new/different slice
// - t.output points to a new/different slice or has changed
// - t.next has data sent over it
// - hpush() is called
func (t *TTY) linechar(ch byte) {
	switch ch {
	case ESC:
		if len(t.output) > 0 {
			t.preescape = t.output
			t.output = make([]byte, 0, 8)
		}
		t.output = append(t.output, ESC)
	case '\r', '\n':
		t.echo('\r', '\n')
		t.hpush()
		fallthrough
	case SOH, STX, ETX, EOT, ENQ, ACK, BEL, VT, FF, SO, SI, DLE, DC1,
		DC2, DC3, DC4, NAK, SYN, ETB, CAN, EM, SUB, FS, GS, RS, US:
		t.emit()
		t.next <- []byte{ch}
	case BS, DEL:
		if len(t.output) == 0 || t.linepos == 0 {
			break
		}
		if t.linepos > 0 {
			// Delete onscreen
			if t.screen != nil {
				delta := len(t.output) - t.linepos
				overwrite := make([]byte, 1+1+2*delta+1)
				overwrite[0] = ch
				copy(overwrite[1:], t.output[t.linepos:])
				overwrite[1+delta] = ' '
				for i := 0; i < delta+1; i++ {
					overwrite[2+delta+i] = '\b'
				}
				t.echo(overwrite...)
			}
			// Delete from output
			t.output = append(t.output[:t.linepos-1], t.output[t.linepos:]...)
			t.linepos--
			break
		}
		t.echo(ch, ' ', ch)
		t.output = t.output[:len(t.output)-1]
	default:
		if t.linepos >= 0 {
			// Insert on screen
			if t.screen != nil {
				delta := len(t.output) - t.linepos
				overwrite := make([]byte, 1+2*delta)
				overwrite[0] = ch
				copy(overwrite[1:], t.output[t.linepos:])
				for i := 0; i < delta; i++ {
					overwrite[1+delta+i] = '\b'
				}
				t.echo(overwrite...)
			}
			// Insert into output
			t.output = append(t.output[:t.linepos],
				append([]byte{ch}, t.output[t.linepos:]...)...)
			t.linepos++
			break
		}
		t.echo(ch)
		t.output = append(t.output, ch)
	}
}

// lineesc processes the next character from a potential escape sequence in
// line mode.
//
// If the second character is not [, then the original output is restored and
// the queued bytes are echoed and the character is processed by char()
//
// The escape sequence ends with the first "printing" character (@ to ~) after
// the <ESC>[ sequence, and that character indicates the action.  The following
// actions are known:
//   A - Up
//   B - Down
//   C - Right
//   D - Left
//   ~ - PageUp/PageDown
// These have optional arguments before them, which are all currently ignored.
// Most of them don't do anything, but these known escape sequences are not
// written out.  If the escape sequence is not known, however, the original
// output is restored with the escape sequence appended.
//   Up    - loads the last saved line
//   Down  - goes to the end of the current line
//   Left  - goes one character closer to the beginning of the line
//   Right - goes one character closer to the end of the line
//
// Side Effects: (possible)
// - t.output refers to a new/different slice
// - t.preescape refers to a new/different slice or nil
// - char() is called
func (t *TTY) lineesc(ch byte) {
	if len(t.output) == 1 {
		if ch != '[' {
			t.echo(t.output...)
			t.output = append(t.preescape, t.output...)
			t.preescape = nil
			t.linechar(ch)
		} else {
			t.output = append(t.output, ch)
		}
		return
	}
	t.output = append(t.output, ch)
	if ch >= '@' && ch <= '~' {
		switch ch {
		case 'A': // up
			t.hprev()
			return
		case 'B': // down
			if t.linepos < 0 {
				break
			}
			t.echo(t.preescape[t.linepos:]...)
			t.linepos = -1
		case 'C': // right
			if len(t.preescape) == 0 {
				break
			}
			if t.linepos < 0 {
				break
			}
			t.echo(t.output...)
			t.linepos++
			if t.linepos == len(t.preescape) {
				t.linepos = -1
			}
		case 'D': // left
			if len(t.preescape) == 0 {
				break
			}
			if t.linepos < 0 {
				t.linepos = len(t.preescape)
			}
			if t.linepos > 0 {
				t.echo(t.output...)
				t.linepos--
			}
		case '~': // pgup(5~)/dn(6~)
		default:
			t.output = append(t.preescape, t.output...)
			t.preescape = nil
			return
		}
		t.output = t.preescape
		t.preescape = nil
	}
}
