// +build linux darwin

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

// Package termios implements low-level terminal settings.
package termios

import (
	"fmt"
	"syscall"
	"unsafe"
)

/*
#include <termios.h>
#include <sys/ioctl.h>
*/
import "C"

type (
	inMode    C.tcflag_t
	outMode   C.tcflag_t
	ctlMode   C.tcflag_t
	locMode   C.tcflag_t
	charIndex C.cc_t
)

// Input Flags
const (
	IGNBRK  inMode = C.IGNBRK  // ignore BREAK condition
	BRKINT  inMode = C.BRKINT  // map BREAK to SIGINTR
	IGNPAR  inMode = C.IGNPAR  // ignore (discard) parity errors
	PARMRK  inMode = C.PARMRK  // mark parity and framing errors
	INPCK   inMode = C.INPCK   // enable checking of parity errors
	ISTRIP  inMode = C.ISTRIP  // strip 8th bit off chars
	INLCR   inMode = C.INLCR   // map NL into CR
	IGNCR   inMode = C.IGNCR   // ignore CR
	ICRNL   inMode = C.ICRNL   // map CR to NL (ala CRMOD)
	IXON    inMode = C.IXON    // enable output flow control
	IXOFF   inMode = C.IXOFF   // enable input flow control
	IXANY   inMode = C.IXANY   // any char will restart after stop
	IMAXBEL inMode = C.IMAXBEL // ring bell on input queue full
	IUTF8   inMode = C.IUTF8   // maintain state for UTF-8 VERASE
)

// Output Flags
const (
	OPOST  outMode = C.OPOST  // enable following output processing
	ONLCR  outMode = C.ONLCR  // map NL to CR-NL (ala CRMOD)
	OCRNL  outMode = C.OCRNL  // map CR to NL on output
	ONOCR  outMode = C.ONOCR  // no CR output at column 0
	ONLRET outMode = C.ONLRET // NL performs CR function
	OFILL  outMode = C.OFILL  // use fill characters for delay
	NLDLY  outMode = C.NLDLY  // \n delay
	TABDLY outMode = C.TABDLY // horizontal tab delay
	CRDLY  outMode = C.CRDLY  // \r delay
	FFDLY  outMode = C.FFDLY  // form feed delay
	BSDLY  outMode = C.BSDLY  // \b delay
	VTDLY  outMode = C.VTDLY  // vertical tab delay
	OFDEL  outMode = C.OFDEL  // fill is DEL, else NUL
)

// Control Flags
const (
	CSIZE  ctlMode = C.CSIZE  // character size mask
	CS6    ctlMode = C.CS6    // 6 bits
	CS7    ctlMode = C.CS7    // 7 bits
	CS8    ctlMode = C.CS8    // 8 bits
	CSTOPB ctlMode = C.CSTOPB // send 2 stop bits
	CREAD  ctlMode = C.CREAD  // enable receiver
	PARENB ctlMode = C.PARENB // parity enable
	PARODD ctlMode = C.PARODD // odd parity, else even
	HUPCL  ctlMode = C.HUPCL  // hang up on last close
	CLOCAL ctlMode = C.CLOCAL // ignore modem status lines
)

// Local flags
const (
	ECHOKE  locMode = C.ECHOKE  // visual erase for line kill
	ECHOE   locMode = C.ECHOE   // visually erase chars
	ECHOK   locMode = C.ECHOK   // echo NL after line kill
	ECHO    locMode = C.ECHO    // enable echoing
	ECHONL  locMode = C.ECHONL  // echo NL even if ECHO is off
	ECHOPRT locMode = C.ECHOPRT // visual erase mode for hardcopy
	ECHOCTL locMode = C.ECHOCTL // echo control chars as ^(Char)
	ISIG    locMode = C.ISIG    // enable signals INTR, QUIT, [D]SUSP
	ICANON  locMode = C.ICANON  // canonicalize input lines
	IEXTEN  locMode = C.IEXTEN  // enable DISCARD and LNEXT
	EXTPROC locMode = C.EXTPROC // external processing
	TOSTOP  locMode = C.TOSTOP  // stop background jobs from output
	FLUSHO  locMode = C.FLUSHO  // output being flushed (state)
	PENDIN  locMode = C.PENDIN  // XXX retype pending input (state)
	NOFLSH  locMode = C.NOFLSH  // don't flush after interrupt
)

// Control Character Indices
const (
	VEOF     charIndex = C.VEOF     // ICANON
	VEOL     charIndex = C.VEOL     // ICANON
	VEOL2    charIndex = C.VEOL2    // ICANON together with IEXTEN
	VERASE   charIndex = C.VERASE   // ICANON
	VWERASE  charIndex = C.VWERASE  // ICANON together with IEXTEN
	VKILL    charIndex = C.VKILL    // ICANON
	VREPRINT charIndex = C.VREPRINT // ICANON together with IEXTEN
	VINTR    charIndex = C.VINTR    // ISIG
	VQUIT    charIndex = C.VQUIT    // ISIG
	VSUSP    charIndex = C.VSUSP    // ISIG
	VSTART   charIndex = C.VSTART   // IXON, IXOFF
	VSTOP    charIndex = C.VSTOP    // IXON, IXOFF
	VLNEXT   charIndex = C.VLNEXT   // IEXTEN
	VDISCARD charIndex = C.VDISCARD // IEXTEN
	VMIN     charIndex = C.VMIN     // !ICANON
	VTIME    charIndex = C.VTIME    // !ICANON
	NCC      charIndex = C.NCCS     // Number of control chars
)

// TermSettings contain both the original settings from when it was created
// and the current settings being manipulated.  At any time, Reset will
// restore the terminal to its original state.
type TermSettings struct {
	fd       int
	original C.struct_termios
	current  C.struct_termios
}

// NewTermSettings examines the state of the current terminal and
// stores it in a fresh TermSettings.
func NewTermSettings(fd int) (*TermSettings, error) {
	tio := &TermSettings{fd: fd}

	if ret, errno := C.tcgetattr(C.int(fd), &tio.current); ret != 0 {
		return nil, errno
	}
	tio.original = tio.current
	return tio, nil
}

// Char returns the rune associated with the given control
// character.  These will generally be ASCII control characters.
func (tio *TermSettings) Char(idx charIndex) rune {
	return rune(tio.current.c_cc[int(idx)])
}

// String returns a debugging string which contains low-level
// information about the terminal.
func (tio *TermSettings) String() string {
	return fmt.Sprintf(`Terminal[%d]:
  Input   = 0x%X
  Output  = 0x%X
  Control = 0x%X
  Local   = 0x%X
  Chars   = %v
`,
		tio.fd,
		tio.current.c_iflag,
		tio.current.c_oflag,
		tio.current.c_cflag,
		tio.current.c_lflag,
		tio.current.c_cc)
}

// GetSize attempts to determine the size of the terminal with which
// this TermSettings is associated and return the number of rows (the height)
// and the number of columns (width).
func (tio *TermSettings) GetSize() (width, height int, err error) {
	var ws C.struct_winsize
	_, _, errno := syscall.RawSyscall(syscall.SYS_IOCTL,
		uintptr(tio.fd),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return 0, 0, syscall.Errno(errno)
	}
	height = int(ws.ws_row)
	width = int(ws.ws_col)
	return
}

// Raw sets the terminal to a very minimal raw mode suitable for simulating a
// terminal emulator or doing raw line editing.
//
// The changes are applied immediately.
//
// I recommend this being done early on in main() and having a deferred call to
// tio.Reset so that the changes will be reverted when everything exits
// cleanly.
func (tio *TermSettings) Raw() error {
	//tio.SetInput(IGNBRK | IXANY)
	//tio.SetOutput(0)
	//tio.SetLocal(0)
	C.cfmakeraw(&tio.current)
	return tio.Apply()
}

// Reset sets the terminal settings to match those that were in effect when the
// call to NewTermSettings was made.
func (tio *TermSettings) Reset() error {
	tio.current = tio.original
	return tio.Apply()
}

func (tio *TermSettings) SetInput(mode inMode)    { tio.current.c_iflag = C.tcflag_t(mode) }
func (tio *TermSettings) SetOutput(mode outMode)  { tio.current.c_oflag = C.tcflag_t(mode) }
func (tio *TermSettings) SetControl(mode ctlMode) { tio.current.c_cflag = C.tcflag_t(mode) }
func (tio *TermSettings) SetLocal(mode locMode)   { tio.current.c_lflag = C.tcflag_t(mode) }

// Apply applies the settings currently stored in tio.  This is mostly useful
// for maintaining multiple TerminalSettings for different modes, and you can
// simply Apply whichever you need.
func (tio *TermSettings) Apply() error {
	const when = C.TCSANOW
	if ret, errno := C.tcsetattr(C.int(tio.fd), when, &tio.current); ret != 0 {
		return errno
	}
	return nil
}

/*
Cooked:
Input   = 0x00002B02
Output  = 0x00000003
Control = 0x00004B00
Local   = 0x200005CB

Raw:

iIGNBRK  + 0x00000001 // ignore BREAK condition
iIXANY   + 0x00000800 // any char will restart after stop
Input    = 0x00000801

oONLCR  + 0x00000002 // map NL to CR-NL (ala CRMOD)
Output  = 0x00000002

lECHOKE     + 0x00000001 // visual erase for line kill
lECHOCTL    + 0x00000040 // echo control chars as ^(Char)
Local       = 0x00000041
*/
