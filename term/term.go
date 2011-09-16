package term

import (
	"io"
	"os"
	"sync"

	"log"
)

var _ = log.Printf

type TTY struct {
	tty io.Reader

	etty io.Writer
	echo bool

	next chan []byte

	lock    sync.RWMutex
	error   os.Error
	partial []byte

	line bool
	size int
	hist int
}

func NewTTY(user io.Reader) *TTY {
	t := &TTY{
		tty:  user,
		next: make(chan []byte, 32),
		line: true,
		size: 32,
		hist: 3,
	}

	t.etty, t.echo = user.(io.Writer)

	go t.run()

	return t
}

func (t *TTY) run() {
	defer close(t.next)

	buffer := make([]byte, t.size)
	output := make([]byte, 0, t.size)

	// TODO(kevlar): These can become methods on TTY
	echo := func(b ...byte) {
		if t.echo {
			if _, err := t.etty.Write(b); err != nil {
				t.echo = false
			}
		}
	}

	var hist []byte
	var preescape []byte

	hpush := func() {
		if len(output) == 0 || output[0] < 0x32 {
			return
		}
		hist = make([]byte, len(output))
		copy(hist, output)
	}

	hprev := func() {
		if len(hist) == 0 {
			return
		}
		width := len(preescape)
		output = hist
		preescape = nil
		if t.echo {
			size, delta := 1+len(output), width-len(output)
			if delta > 0 {
				size += 2 * delta
			}
			overwrite := make([]byte, size)
			overwrite[0] = '\r'
			copy(overwrite[1:], output)
			for i := len(output); i < width; i++ {
				overwrite[1+i] = ' '
				overwrite[1+i+delta] = '\b'
			}
			echo(overwrite...)
		}
	}

	char := func(ch byte) {
		switch ch {
		case ESC:
			if len(output) > 0 {
				preescape = output
				output = make([]byte, 0, 8)
			}
			output = append(output, ESC)
		case '\r', '\n':
			echo('\r', '\n')
			hpush()
			fallthrough
		case SOH, STX, ETX, EOT, ENQ, ACK, BEL, VT, FF, SO, SI, DLE, DC1,
			DC2, DC3, DC4, NAK, SYN, ETB, CAN, EM, SUB, FS, GS, RS, US:
			if len(output) > 0 {
				t.next <- output
				output = make([]byte, 0, t.size)
			}
			t.next <- []byte{ch}
		case BS:
			if len(output) > 0 {
				echo(BS, ' ', BS)
				output = output[:len(output)-1]
			}
		default:
			echo(ch)
			output = append(output, ch)
		}
	}

	esc := func(ch byte) {
		if len(output) == 1 {
			if ch != '[' {
				echo(output...)
				output = append(preescape, output...)
				preescape = nil
				char(ch)
			} else {
				output = append(output, ch)
			}
			return
		}
		output = append(output, ch)
		if ch >= '@' && ch <= '~' {
			switch ch {
			case 'A': // up
				hprev()
				return
			case 'B': // down
			case 'C': // right
			case 'D': // left
			case '~': // pgup(5~)/dn(6~)
			default:
				output = append(preescape, output...)
				preescape = nil
				return
			}
			output = preescape
			preescape = nil
		}
	}

	for {
		n, err := t.tty.Read(buffer)
		if err != nil {
			if len(output) > 0 {
				t.next <- append(preescape, output...)
			}
			t.error = err
			return
		}
		for _, ch := range buffer[:n] {
			if len(output) > 0 && output[0] == ESC {
				esc(ch)
			} else {
				char(ch)
			}
		}
	}
}

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
