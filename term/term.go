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
	"sync"
)

// The following constants are provided for your own edification; they are the
// internal defaults and cannot be changed.
const (
	ReadBufferLength       = 32
	DefaultLineBufferSize  = 32
	DefaultRawBufferSize   = 256
	DefaultFrameBufferSize = 8
)

type ttyMode int

// The following constants are the modes in which the TTY can be set
const (
	Raw   ttyMode = iota // All reads are passed through
	Line                 // Basic line-editing capabilities are provided
	Frame                // Basic screen-editing capabilities are provided
)

// A TTY is a simple interface for reading input from a user over a raw
// terminal emulation interface.
//
// All methods on TTY are goroutine-safe (though calling Read concurrently
// is probably not the most deterministic thing in the world to do).
type TTY struct {
	// IO
	console io.Reader
	screen  io.Writer

	// Synchronization and reading
	next    chan []byte    // Completed chunks (usually lines)
	partial []byte         // Store partial reads
	lock    sync.RWMutex   // Synchronize multiple readers (locks partial)
	error   error          // The error when the reader closed
	update  chan chan bool // Take ownership of the IO and Settings data

	// Settings
	mode  ttyMode // The current mode of the TTY
	bsize int     // Initial line buffer size

	// State (Line mode)
	buffer    []byte // The last read from console
	output    []byte // The pending line/chunk
	last      []byte // The last line/chunk (used for prevline)
	preescape []byte // The contents of output before the escape sequence
	linepos   int    // >= 0 if doing in-place line editing

	// State (Frame mode)
	regions []*Region
	active  int
}

// NewTTY creates a new TTY for interacting with a user via a limited
// line-oriented interface.  If the given reader is also an io.Writer,
// interactive echo is enabled.
func NewTTY(console io.Reader) *TTY {
	t := &TTY{
		console: console,
		next:    make(chan []byte, ReadBufferLength),
		mode:    Line,
		bsize:   DefaultLineBufferSize,
		update:  make(chan chan bool),
	}

	t.screen, _ = console.(io.Writer)

	go t.run()
	return t
}

// NewFrameTTY creates a new TTY for interacting with a user via a
// screen-oriented interface.  If the given reader is also an io.Writer,
// interactive echo is enabled.
//
// A TTY created with NewFrameTTY has synchronized reads, so further input is
// not processed until the chunk has been read.  The default read buffer size
// for a Frame TTY is much smaller than the others.
//
// The default region for a new Frame is an 80x24 region with the initial
// cursor placed in the upper right-hand corner.  This region is returned,
// but will not have been been drawn (e.g. its settings can be changed).
func NewFrameTTY(console io.ReadWriter) (*TTY, *Region) {
	t := &TTY{
		console: console,
		screen:  console,
		next:    make(chan []byte),
		mode:    Frame,
		bsize:   DefaultFrameBufferSize,
		update:  make(chan chan bool),
	}

	go t.run()
	r := t.NewRegion(80, 24, 0, 0)
	return t, r
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

// SetEcho enables or disables interactive echo, sending all writes on the
// given writer.  Whether the echo writer is specified here or inferred in
// NewTTY, any write error will disable echo.  Providing nil to SetEcho
// disables interactive echo.
func (t *TTY) SetEcho(echo io.Writer) {
	lock := make(chan bool, 1)
	t.update <- lock
	t.screen = echo
	lock <- true
}

// SetLineBuffer sets the initial line buffer size.  In general, you shouldn't
// need to change this, as the line buffer will continue to grow if the line is
// really long, but if you find that you have lots of really long lines it
// might help reduce garbage.
func (t *TTY) SetLineBuffer(size int) {
	lock := make(chan bool, 1)
	t.update <- lock
	t.bsize = size
	lock <- true
}

// SetMode sets the TTY mode.
//
// Raw: No line buffering is performed, and data is written exactly as it is
// received, including control sequences.  The input is not broken up into
// logical units, reads are passed through directly.  In the case of an
// interactive session, this will often be broken up into individual characters
// or control sequences, not lines or words.
//
// Line: Basic line-buffering is performed.  See the package comment.
//
// Frame: Basic screen-editing is enabled.  Currently the same as Line.
//
// Switching modes will suspend any state tracking for the old mode.  Switching
// back will resume with the state where it was before the mode was changed,
// but this may result in unforseen side effects.  Changing modes does not
// effect the line buffer size or whether reads are synchronous, as is the case
// for TTYs created explicitly in a certain mode.  It should not usually be
// necessary to change modes.
func (t *TTY) SetMode(mode ttyMode) {
	lock := make(chan bool, 1)
	t.update <- lock
	t.mode = mode
	lock <- true
}

// echo echoes the bytes if interactive editing is enabled
//
// Side effects:
// - If there is a write error, interactive editing is disabled
func (t *TTY) echo(b ...byte) {
	if t.screen != nil {
		if _, err := t.screen.Write(b); err != nil {
			t.screen = nil
		}
	}
}

// yield gives the chance for an update to proceed
//
// Side effects:
// - anything
func (t *TTY) yield() {
	select {
	case done := <-t.update:
		<-done
	default:
	}
}

// emit sends the contents of t.output over the t.next channel, optionally
// prefixing it with the preescape if any.  Nothing is done if the length of
// output (including preescape) is zero.
//
// Side effects:
// - t.output refers to a newly allocated zero-length slice (with capacity t.bsize)
// - t.preescape is nil
// - the output is written to t.next
func (t *TTY) emit() {
	if len(t.preescape) > 0 {
		t.output = append(t.preescape, t.output...)
		t.preescape = nil
	}
	if len(t.output) > 0 {
		t.next <- t.output
		t.output = make([]byte, 0, t.bsize)
		t.linepos = -1
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
	t.linepos = -1

	for {
		t.yield()
		n, err := t.console.Read(t.buffer)
		if err != nil {
			t.emit()
			t.error = err
			return
		}
		t.yield()

		switch t.mode {
		case Raw:
			t.next <- t.buffer[:n]
		case Line, Frame:
			// Process each character that was read
			for _, ch := range t.buffer[:n] {
				if len(t.output) > 0 && t.output[0] == ESC {
					t.lineesc(ch)
				} else {
					t.linechar(ch)
				}
			}
		}
	}
}

// Read reads the next line, chunk, control sequence, etc from the console.
func (t *TTY) Read(b []byte) (n int, err error) {
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
func (t *TTY) Write(b []byte) (n int, err error) {
	w := t.screen
	if w == nil {
		return 0, io.EOF
	}
	return w.Write(b)
}
