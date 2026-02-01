package dns

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"
)

type DnsLogger struct {
	mu  sync.Mutex
	f   *os.File
	w   *bufio.Writer
	enc *json.Encoder
}

func NewDnsLogger(path string) (*DnsLogger, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	w := bufio.NewWriterSize(f, 256*1024)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	return &DnsLogger{
		f:   f,
		w:   w,
		enc: enc,
	}, nil
}

func (l *DnsLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = l.w.Flush()
	return l.f.Close()
}

func (l *DnsLogger) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Flush()
}

func (l *DnsLogger) Write(v any) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.enc.Encode(v); err != nil {
		return err
	}
	return nil
}

func Timestamp() string {
	return time.Now().Format(time.RFC3339Nano)
}
