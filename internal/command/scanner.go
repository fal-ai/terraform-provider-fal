package command

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

const (
	maxBufferSize = 128 * 1024
)

var (
	cr  = []byte(" \r")
	crr = []byte("\r\n")
)

func scan(r io.Reader, callback func(line []byte)) error {
	br := bufio.NewReaderSize(r, 4096)

	var buf bytes.Buffer
	for {
		buf.Reset()
		var err error
		var line []byte
		var isPrefix bool

		for {
			line, isPrefix, err = br.ReadLine()

			// certain apps output absolutely random carriage returns in the output
			// seemingly in line with that it thinks is the terminal size. those returns
			// break a lot of output handling, so we'll just replace them with proper new
			// lines and then split it later
			line = bytes.Replace(line, cr, crr, -1)
			ns := buf.Len() + len(line)

			// if the length of the line value and the current value in the buffer will
			// exceed the maximum buffer size, chop it down to hit the maximum size and then
			// write that data before ending this loop. its kinda a re-implementation of the
			// scanner logic without triggering an error if you exceed the buffer size
			if ns > maxBufferSize {
				buf.Write(line[:len(line)-(ns-maxBufferSize)])
				break
			} else {
				buf.Write(line)
			}

			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			// finish this loop and begin outputting the line if there is no prefix (the
			// line fit into the default buffer), or if we hit the end of the line
			if !isPrefix || errors.Is(err, io.EOF) {
				break
			}
		}

		if buf.Len() > 0 {
			// you need to make a copy of the buffer here to avoid a race condition
			c := make([]byte, buf.Len())
			copy(c, buf.Bytes())
			callback(c)
		}

		if errors.Is(err, io.EOF) {
			break
		}
	}
	return nil
}
