package term

import (
	"io"
	"os"
	"sync"
)

// The following constants are provided for your own edification; they are the
// internal defaults and cannot be changed.
const (
	ReadBufferLength      = 32
	DefaultLineBufferSize = 32
	DefaultRawBufferSize  = 256
)

// A TTY is a simple interface for reading input from a user over a raw
// terminal emulation interface.
//
// All methods on TTY are goroutine-safe (though calling Read concurrently
// is probably not the most deterministic thing in the world to do).
type TTY struct {
	// IO
	console io.Reader
	intecho io.Writer

	// Synchronization and reading
	next    chan []byte    // Completed chunks (usually lines)
	partial []byte         // Store partial reads
	lock    sync.RWMutex   // Synchronize multiple readers (locks partial)
	error   os.Error       // The error when the reader closed
	update  chan chan bool // Take ownership of the IO and Settings data

	// Settings
	cooked bool // Enable line editing
	bsize  int  // Initial line buffer size

	// State
	buffer    []byte // The last read from console
	output    []byte // The pending line/chunk
	last      []byte // The last line/chunk (used for prevline)
	preescape []byte // The contents of output before the escape sequence

}

// NewTTY creates a new TTY for interacting with a user via a limited
// line-oriented interface.  If the given reader is also an io.Writer,
// interactive echo is enabled.
func NewTTY(console io.Reader) *TTY {
	t := &TTY{
		console: console,
		next:    make(chan []byte, ReadBufferLength),
		cooked:  true,
		bsize:   DefaultLineBufferSize,
		update:  make(chan chan bool),
	}

	t.intecho, _ = console.(io.Writer)

	go t.run()
	return t
}

// NewRawTTY creates a new TTY without line editing and with a larger potential
// input buffer size, and with no interactive echo.
func NewRawTTY(console io.Reader) *TTY {
	t := &TTY{
		console: console,
		next:    make(chan []byte, ReadBufferLength),
		bsize:   DefaultRawBufferSize,
		update:  make(chan chan bool),
	}

	go t.run()
	return t
}

// Echo enables interactive echo, sending all writes on the given writer.
// Whether the echo writer is specified here or inferred in NewTTY, any write
// error will disable echo.  Providing nil to Echo disables interactive echo.
func (t *TTY) Echo(echo io.Writer) {
	lock := make(chan bool, 1)
	t.update <- lock
	t.intecho = echo
	lock <- true
}

// BufferSize sets the initial line buffer size.  In general, you shouldn't
// need to change this, as the line buffer will continue to grow if the line is
// really long, but if you find that you have lots of really long lines it
// might help reduce garbage.
func (t *TTY) BufferSize(size int) {
	lock := make(chan bool, 1)
	t.update <- lock
	t.bsize = size
	lock <- true
}

// Cooked sets whether line buffering is performed.  If no line buffering is
// performed (e.g. cooked is false) data is written exactly as it is received.
// In the case of an interactive session, this will often be broken up into
// individual characters or control sequences, not lines or words.
//
// Switching to raw mode from cooked will suspend any line editing state and
// receive bytes directly.  Switching back to cooked mode will resume with the
// state where it was before cooked was enabled.  In most cases, it will not
// be necessary to switch between the two modes.
func (t *TTY) Cooked(cooked bool) {
	lock := make(chan bool, 1)
	t.update <- lock
	t.cooked = cooked
	lock <- true
}

// echo echoes the bytes if interactive editing is enabled
//
// Side effects:
// - If there is a write error, interactive editing is disabled
func (t *TTY) echo(b ...byte) {
	if t.intecho != nil {
		if _, err := t.intecho.Write(b); err != nil {
			t.intecho = nil
		}
	}
}

// hpush (history push) stores the line for later reuse if it
// is not an escape sequence and contains characters.
//
// Side effects: (only if output is nonzero and not an escape sequence)
// - t.last will contain a copy of output
func (t *TTY) hpush() {
	if len(t.output) == 0 || t.output[0] < 0x32 {
		return
	}
	t.last = make([]byte, len(t.output))
	copy(t.last, t.output)
}

// hprev (history previous) replaces the current output with the last
// saved line (unless no line has been saved).
//
// To echo the new line, the following is written:
//   \r<line><spaces><backspaces>
// Where <line> is the new output <spaces> and <backspaces> are present if the
// previous line was long enough to require them to not leave dangling
// characters.
//
// Preconditions:
// - Must be called within an escape sequence
// Side effects: (only if t.last has been set)
// - t.output will contain a copy of t.last
// - t.preescape will be nil
func (t *TTY) hprev() {
	if len(t.last) == 0 {
		return
	}

	t.output = make([]byte, len(t.last))
	copy(t.output, t.last)

	width := len(t.preescape)
	t.preescape = nil

	if t.intecho != nil {
		size, delta := 1+len(t.output), width-len(t.output)
		if delta > 0 {
			size += 2 * delta
		}
		overwrite := make([]byte, size)
		overwrite[0] = '\r'
		copy(overwrite[1:], t.output)
		for i := len(t.output); i < width; i++ {
			overwrite[1+i] = ' '
			overwrite[1+i+delta] = '\b'
		}
		t.echo(overwrite...)
	}
}

// char processes the next character of input.
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
func (t *TTY) char(ch byte) {
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
		if len(t.output) > 0 {
			t.next <- t.output
			t.output = make([]byte, 0, t.bsize)
		}
		t.next <- []byte{ch}
	case BS, DEL:
		if len(t.output) > 0 {
			t.echo(ch, ' ', ch)
			t.output = t.output[:len(t.output)-1]
		}
	default:
		t.echo(ch)
		t.output = append(t.output, ch)
	}
}

// esc processes the next character from a potential escape sequence.
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
//
// Side Effects: (possible)
// - t.output refers to a new/different slice
// - t.preescape refers to a new/different slice or nil
// - char() is called
func (t *TTY) esc(ch byte) {
	if len(t.output) == 1 {
		if ch != '[' {
			t.echo(t.output...)
			t.output = append(t.preescape, t.output...)
			t.preescape = nil
			t.char(ch)
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
		case 'C': // right
		case 'D': // left
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

// yield gives the chance for an update to proceed
func (t *TTY) yield() {
	select {
	case done := <-t.update: <-done
	default:
	}
}

// run is the primary reading goroutine.  It reads chunks from the console, and processes them
// or (if not in cooked mode) outputs them directly.  Before each read, it gives the setter
// methods the opportunity to pause it while they poke at the TTY internals.  This is not
// necessary for reading, which takes data directly from the next channel.
func (t *TTY) run() {
	defer close(t.next)

	t.buffer = make([]byte, t.bsize)
	t.output = make([]byte, 0, t.bsize)

	for {
		t.yield()
		n, err := t.console.Read(t.buffer)
		if err != nil {
			if len(t.output) > 0 {
				t.next <- append(t.preescape, t.output...)
			}
			t.error = err
			return
		}
		t.yield()

		// Bypass line editing if we're not in cooked mode
		if !t.cooked {
			t.next <- t.buffer[:n]
			continue
		}

		// Process each character that was read
		for _, ch := range t.buffer[:n] {
			if len(t.output) > 0 && t.output[0] == ESC {
				t.esc(ch)
			} else {
				t.char(ch)
			}
		}
	}
}

// Read reads the next line, chunk, control sequence, etc from the console.
func (t *TTY) Read(b []byte) (n int, err os.Error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	var ok bool
	if len(t.partial) == 0 {
		if t.partial, ok = <-t.next; !ok {
			return 0, t.error
		}
	}

	n = copy(b, t.partial)
	t.partial = t.partial[n:]
	return
}

// Write writes to the same io.Writer that is handing the interactive echo.  If
// interactive echo is disabled (either directly or because an echo write
// failed) Write will return EOF.
func (t *TTY) Write(b []byte) (n int, err os.Error) {
	w := t.intecho
	if w == nil {
		return 0, os.EOF
	}
	return w.Write(b)
}
