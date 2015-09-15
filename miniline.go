package miniline

import (
	"bufio"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

type InterruptedError struct{}

func (_ InterruptedError) Error() string {
	return "Interrupted"
}

var ErrInterrupted error = InterruptedError{}

type lineReader struct {
	reader *bufio.Reader
	writer *bufio.Writer
	buf    []byte
	pos    int
}

func (lr *lineReader) esc(s string) (err error) {
	_, err = lr.writer.WriteString("\x1b[" + s)
	return
}

func (lr *lineReader) pbuf() (err error) {
	err = lr.esc("s")
	if err == nil {
		_, err = lr.writer.Write(lr.buf[lr.pos:])
	}
	if err == nil {
		err = lr.esc("u")
	}
	return
}

func (lr *lineReader) backspace() (err error) {
	if len(lr.buf) == 0 {
		return
	}

	err = lr.esc("D")
	if err == nil {
		err = lr.esc("K")
	}
	if lr.pos < len(lr.buf) {
		copy(lr.buf[lr.pos-1:], lr.buf[lr.pos:])
		lr.buf = lr.buf[:len(lr.buf)-1]
	} else {
		lr.buf = lr.buf[:len(lr.buf)-1]
	}
	lr.pos--

	if err == nil {
		err = lr.pbuf()
	}
	return
}

func (lr *lineReader) readEscape() (err error) {
	b, err := lr.reader.ReadByte()
	if err != nil {
		return
	}
	if b != byte('[') {
		err = lr.writer.WriteByte(0x7)
		return
	}
	b, err = lr.reader.ReadByte()
	if err != nil {
		return
	}
	switch b {
	case byte('A'), byte('B'): // up and down, noop
	case byte('C'): // right
		if lr.pos == len(lr.buf) {
			return
		}
		err = lr.esc("C")
		lr.pos++
	case byte('D'): // left
		if lr.pos == 0 {
			return
		}
		err = lr.esc("D")
		lr.pos--
	}
	return
}

func (lr *lineReader) readLine() error {
	for {
		if err := lr.writer.Flush(); err != nil {
			return err
		}

		b, err := lr.reader.ReadByte()
		if err != nil {
			return err
		}
		switch b {
		case 0x4, byte('\r'): // EOF or newline
			return nil
		case 0x7F: // backspace
			if err := lr.backspace(); err != nil {
				return err
			}
			continue
		case 0x03: // ^C
			return ErrInterrupted
		case 0x1b: // ESC
			if err := lr.readEscape(); err != nil {
				return err
			}
			continue
		}
		lr.writer.WriteByte(b)

		if lr.pos == len(lr.buf) {
			lr.buf = append(lr.buf, b)
			lr.pos++
		} else {
			lr.buf = append(lr.buf, 0)
			copy(lr.buf[lr.pos+1:], lr.buf[lr.pos:])
			lr.buf[lr.pos] = b
			lr.pos++
			lr.pbuf()
		}
	}
}

func WithRaw(fd uintptr, fn func() error) (err error) {
	origState, err := terminal.MakeRaw(int(fd))
	if err != nil {
		return
	}
	defer terminal.Restore(int(fd), origState)
	err = fn()
	return
}

func ReadLine(prompt string) (line string, err error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return
	}
	defer tty.Close()

	tty.WriteString(prompt)
	reader := &lineReader{
		reader: bufio.NewReader(tty),
		writer: bufio.NewWriter(tty),
	}

	err = WithRaw(tty.Fd(), func() error {
		return reader.readLine()
	})

	line = string(reader.buf)

	tty.WriteString("\n")
	return
}
