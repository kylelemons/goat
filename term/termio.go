package term

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

type inMode uint64
// Input Flags
const (
	IGNBRK  inMode = 0x00000001 // ignore BREAK condition
	BRKINT  inMode = 0x00000002 // map BREAK to SIGINTR
	IGNPAR  inMode = 0x00000004 // ignore (discard) parity errors
	PARMRK  inMode = 0x00000008 // mark parity and framing errors
	INPCK   inMode = 0x00000010 // enable checking of parity errors
	ISTRIP  inMode = 0x00000020 // strip 8th bit off chars
	INLCR   inMode = 0x00000040 // map NL into CR
	IGNCR   inMode = 0x00000080 // ignore CR
	ICRNL   inMode = 0x00000100 // map CR to NL (ala CRMOD)
	IXON    inMode = 0x00000200 // enable output flow control
	IXOFF   inMode = 0x00000400 // enable input flow control
	IXANY   inMode = 0x00000800 // any char will restart after stop
	IMAXBEL inMode = 0x00002000 // ring bell on input queue full
	IUTF8   inMode = 0x00004000 // maintain state for UTF-8 VERASE
)

type outMode uint64
// Output Flags
const (
	OPOST  outMode = 0x00000001 // enable following output processing
	ONLCR  outMode = 0x00000002 // map NL to CR-NL (ala CRMOD)
	OXTABS outMode = 0x00000004 // expand tabs to spaces
	ONOEOT outMode = 0x00000008 // discard EOT's (^D) on output)
	OCRNL  outMode = 0x00000010 // map CR to NL on output
	ONOCR  outMode = 0x00000020 // no CR output at column 0
	ONLRET outMode = 0x00000040 // NL performs CR function
	OFILL  outMode = 0x00000080 // use fill characters for delay
	NLDLY  outMode = 0x00000300 // \n delay
	TABDLY outMode = 0x00000c04 // horizontal tab delay
	CRDLY  outMode = 0x00003000 // \r delay
	FFDLY  outMode = 0x00004000 // form feed delay
	BSDLY  outMode = 0x00008000 // \b delay
	VTDLY  outMode = 0x00010000 // vertical tab delay
	OFDEL  outMode = 0x00020000 // fill is DEL, else NUL
)

type ctlMode uint64
// Control Flags
const (
	CIGNORE    ctlMode = 0x00000001 // ignore control flags
	CSIZE      ctlMode = 0x00000300 // character size mask
	CS6        ctlMode = 0x00000100 // 6 bits
	CS7        ctlMode = 0x00000200 // 7 bits
	CS8        ctlMode = 0x00000300 // 8 bits
	CSTOPB     ctlMode = 0x00000400 // send 2 stop bits
	CREAD      ctlMode = 0x00000800 // enable receiver
	PARENB     ctlMode = 0x00001000 // parity enable
	PARODD     ctlMode = 0x00002000 // odd parity, else even
	HUPCL      ctlMode = 0x00004000 // hang up on last close
	CLOCAL     ctlMode = 0x00008000 // ignore modem status lines
	CCTS_OFLOW ctlMode = 0x00010000 // CTS flow control of output
	CRTS_IFLOW ctlMode = 0x00020000 // RTS flow control of input
	CDTR_IFLOW ctlMode = 0x00040000 // DTR flow control of input
	CDSR_OFLOW ctlMode = 0x00080000 // DSR flow control of output
	CCAR_OFLOW ctlMode = 0x00100000 // DCD flow control of output
)

type locMode uint64
// Local flags
const (
	ECHOKE     locMode = 0x00000001 // visual erase for line kill
	ECHOE      locMode = 0x00000002 // visually erase chars
	ECHOK      locMode = 0x00000004 // echo NL after line kill
	ECHO       locMode = 0x00000008 // enable echoing
	ECHONL     locMode = 0x00000010 // echo NL even if ECHO is off
	ECHOPRT    locMode = 0x00000020 // visual erase mode for hardcopy
	ECHOCTL    locMode = 0x00000040 // echo control chars as ^(Char)
	ISIG       locMode = 0x00000080 // enable signals INTR, QUIT, [D]SUSP
	ICANON     locMode = 0x00000100 // canonicalize input lines
	ALTWERASE  locMode = 0x00000200 // use alternate WERASE algorithm
	IEXTEN     locMode = 0x00000400 // enable DISCARD and LNEXT
	EXTPROC    locMode = 0x00000800 // external processing
	TOSTOP     locMode = 0x00400000 // stop background jobs from output
	FLUSHO     locMode = 0x00800000 // output being flushed (state)
	NOKERNINFO locMode = 0x02000000 // no kernel output from VSTATUS
	PENDIN     locMode = 0x20000000 // XXX retype pending input (state)
	NOFLSH     locMode = 0x80000000 // don't flush after interrupt
)

// Control Characters
const (
	vEOF     = iota // ICANON
	vEOL            // ICANON
	vEOL2           // ICANON together with IEXTEN
	vERASE          // ICANON
	vWERASE         // ICANON together with IEXTEN
	vKILL           // ICANON
	vREPRINT        // ICANON together with IEXTEN
	spareCC1        // Spare 1
	vINTR           // ISIG
	vQUIT           // ISIG
	vSUSP           // ISIG
	vDSUSP          // ISIG together with IEXTEN
	vSTART          // IXON, IXOFF
	vSTOP           // IXON, IXOFF
	vLNEXT          // IEXTEN
	vDISCARD        // IEXTEN
	vMIN            // !ICANON
	vTIME           // !ICANON
	vSTATUS         // ICANON together with IEXTEN
	spareCC2        // Spare 2
	nCC             // Number of control chars
)

// Standard speeds
const (
	b0      = 0
	b50     = 50
	b75     = 75
	b110    = 110
	b134    = 134
	b150    = 150
	b200    = 200
	b300    = 300
	b600    = 600
	b1200   = 1200
	b1800   = 1800
	b2400   = 2400
	b4800   = 4800
	b7200   = 7200
	b9600   = 9600
	b19200  = 19200
	bEXTA   = 19200
	b38400  = 38400
	bEXTB   = 38400
	b14400  = 14400
	b28800  = 28800
	b57600  = 57600
	b76800  = 76800
	b115200 = 115200
	b230400 = 230400
)

type termios [0 +
	8 + //   unsigned long c_iflag   // input flags
	8 + //   unsigned long c_oflag   // output flags
	8 + //   unsigned long c_cflag   // control flags
	8 + //   unsigned long c_lflag   // local flags
	nCC + // unsigned char c_cc[nCC] // control chars
	8 + //   unsigned long c_ispeed  // input speed
	8 + //   unsigned long c_ospeed  // output speed
	0]byte

type TermSettings struct {
	fd       int
	original termios
	current  termios
	i, o     []byte
	c, l     []byte
	cc       []byte
	is, os   []byte
}

func NewTermSettings(fd int) (*TermSettings, os.Error) {
	tio := &TermSettings{fd: fd}
	_, _, errno := syscall.RawSyscall(syscall.SYS_IOCTL,
		uintptr(tio.fd),
		uintptr(syscall.TIOCGETA),
		uintptr(unsafe.Pointer(&tio.original)))
	if errno != 0 {
		return nil, os.Errno(errno)
	}
	copy(tio.current[:], tio.original[:])
	tio.i, tio.o = tio.current[000:010], tio.current[010:020]
	tio.c, tio.l = tio.current[020:030], tio.current[030:040]
	tio.cc = tio.current[040 : 040+nCC]
	tio.is, tio.os = tio.current[040+nCC:050+nCC], tio.current[050+nCC:060+nCC]
	return tio, nil
}

func (tio *TermSettings) String() string {
	return fmt.Sprintf(`Terminal[%d]:
  Input   = 0x%X
  Output  = 0x%X
  Control = 0x%X
  Local   = 0x%X
  Chars   = %v
  ISpeed  = 0x%X
  OSpeed  = 0x%X
`, tio.fd, tio.i, tio.o, tio.c, tio.l, tio.cc, tio.is, tio.os)
}

// Raw sets the terminal to a very minimal raw mode suitable for simulating a
// terminal emulator or doing raw line editing.
//
// The changes are applied immediately.
//
// I recommend this being done early on in main() and having a deferred call to
// tio.Reset so that the changes will be reverted when everything exits
// cleanly.
func (tio *TermSettings) Raw() os.Error {
	tio.SetInput(IGNBRK | IXANY)
	tio.SetOutput(0)
	tio.SetLocal(0)
	return tio.Apply()
}

// Reset sets the terminal settings to match those that were in effect when the
// call to NewTermSettings was made.
func (tio *TermSettings) Reset() os.Error {
	copy(tio.current[:], tio.original[:])
	return tio.Apply()
}

var le = binary.LittleEndian

func (tio *TermSettings) SetInput(mode inMode)    { le.PutUint64(tio.i, uint64(mode)) }
func (tio *TermSettings) SetOutput(mode outMode)  { le.PutUint64(tio.o, uint64(mode)) }
func (tio *TermSettings) SetControl(mode ctlMode) { le.PutUint64(tio.c, uint64(mode)) }
func (tio *TermSettings) SetLocal(mode locMode)   { le.PutUint64(tio.l, uint64(mode)) }

// Apply applies the settings currently stored in tio.  This is mostly useful
// for maintaining multiple TerminalSettings for different modes, and you can
// simply Apply whichever you need.
func (tio *TermSettings) Apply() os.Error {
	_, _, errno := syscall.RawSyscall(syscall.SYS_IOCTL,
		uintptr(tio.fd),
		uintptr(syscall.TIOCSETA),
		uintptr(unsafe.Pointer(&tio.current)))
	if errno != 0 {
		return os.Errno(errno)
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
