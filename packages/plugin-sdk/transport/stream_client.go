package transport

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// MaxMessageSize bounds a single message so a corrupt header cannot make the
// reader allocate without limit.
const MaxMessageSize = 64 << 20

// Stream reads and writes framed messages. Writes are safe for concurrent use;
// reads must come from one goroutine.
type Stream struct {
	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex
}

// NewStream wraps a reader and a writer, typically a child process's stdout
// and stdin (host side) or the process's own stdin and stdout (plugin side).
func NewStream(r io.Reader, w io.Writer) *Stream {
	return &Stream{reader: bufio.NewReader(r), writer: w}
}

// Write sends one message body with its header.
func (s *Stream) Write(body []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := fmt.Fprintf(s.writer, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		return err
	}
	_, err := s.writer.Write(body)
	return err
}

// Read returns the next message body. It returns io.EOF when the peer closed
// the stream cleanly between messages.
func (s *Stream) Read() ([]byte, error) {
	length := -1
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) && line == "" && length < 0 {
				return nil, io.EOF
			}
			return nil, fmt.Errorf("reading header: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, found := strings.Cut(line, ":")
		if !found {
			return nil, fmt.Errorf("malformed header %q", line)
		}
		if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			n, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil || n < 0 || n > MaxMessageSize {
				return nil, fmt.Errorf("invalid Content-Length %q", value)
			}
			length = n
		}
	}
	if length < 0 {
		return nil, errors.New("message without Content-Length")
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(s.reader, body); err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}
	return body, nil
}
