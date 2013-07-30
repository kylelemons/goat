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

	DEL = 127 // Delete
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
