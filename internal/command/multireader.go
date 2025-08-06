package command

import (
	"bufio"
	"bytes"
	"io"
	"sync"
)

type concurrentMultiReader struct {
	readers []io.Reader
	data    chan []byte
	err     error
	once    sync.Once
	buffer  []byte
}

func multiReader(readers ...io.Reader) *concurrentMultiReader {
	return &concurrentMultiReader{
		readers: readers,
		data:    make(chan []byte, len(readers)),
	}
}

func (cmr *concurrentMultiReader) spin() {
	var wg sync.WaitGroup

	// launch a goroutine for each reader
	for _, r := range cmr.readers {
		wg.Add(1)
		go func(r io.Reader) {
			defer wg.Done()

			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				line := scanner.Bytes()
				data := make([]byte, len(line)+1)
				copy(data, line)
				data[len(line)] = '\n'
				cmr.data <- data
			}

			if err := scanner.Err(); err != nil {
				cmr.err = err
			}
		}(r)
	}

	go func() {
		wg.Wait()
		close(cmr.data)
	}()
}

func (cmr *concurrentMultiReader) Read(p []byte) (n int, err error) {
	cmr.once.Do(cmr.spin)

	if cmr.err != nil {
		return 0, cmr.err
	}

	// if we have a partial line in the buffer, try to complete it
	if len(cmr.buffer) > 0 {
		if n, ok := cmr.line(p); ok {
			return n, nil
		}
		// if no full line is found, let's read more data
	}

	for {
		chunk, ok := <-cmr.data
		if !ok {
			// return any remaining data if the channel is closed
			if len(cmr.buffer) > 0 {
				n = copy(p, cmr.buffer)
				cmr.buffer = nil
				return n, nil
			}
			return 0, io.EOF
		}

		// append the new data to the buf
		cmr.buffer = append(cmr.buffer, chunk...)

		// check if we have a full line
		if n, ok := cmr.line(p); ok {
			return n, nil
		}
		// if no full line is found, jump back to the top of the loop and keep trying
	}
}

func (cmr *concurrentMultiReader) line(p []byte) (int, bool) {
	// try to detect a full line
	i := bytes.IndexByte(cmr.buffer, '\n')
	if i >= 0 {
		// we have a full line!
		n := copy(p, cmr.buffer[:i+1])
		cmr.buffer = cmr.buffer[i+1:]
		return n, true
	}
	return 0, false
}
