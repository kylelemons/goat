package term

// Terminal Control Codes
const (
	NUL = iota
	SOH // Start of Header
	STX // Start of Text
	ETX // End of Text
	EOT // End of Transmission
	ENQ // Enquire
	ACK // Acknowledge
	BEL // Bell
	BS  // Backspace
	TAB // Horizontal tab
	LF  // Line feed
	VT  // Vertical tab
	FF  // Form feed
	CR  // Carriage return
	SO  // Shift out
	SI  // Shift in
	DLE // Data link escape
	DC1 // Device Control 1
	DC2 // Device Control 2
	DC3 // Device Control 3
	DC4 // Device Control 4
	NAK // Negative Acknowledge
	SYN // Synchronize
	ETB // End Transmission Block
	CAN // CANCEL
	EM  // End of Medium
	SUB // Substitute
	ESC // Escape
	FS  // File separator
	GS  // Group separator
	RS  // Record separator
	US  // Unit separator
)

// Control Constants
//
// These are all emitted by themselves to easily discern them from the rest of
// the sequence of chunks.
const (
	Interrupt = "\x03" // ^C
	EndOfFile = "\x04" // ^D
	Suspend   = "\x1a" // ^Z
	Quit      = "\x1c" // ^\

	CarriageReturn = "\r"
	NewLine        = "\n"
)
