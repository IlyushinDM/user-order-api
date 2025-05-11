package logger_util

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

// --- Tests for asyncWriter ---

func TestNewAsyncWriter_NilWriter(t *testing.T) {
	aw, err := NewAsyncWriter(nil, 10)
	if err == nil || aw != nil {
		t.Errorf("Expected error for nil writer, got aw=%v, err=%v", aw, err)
	}
}

func TestAsyncWriter_WriteAndClose(t *testing.T) {
	var buf bytes.Buffer
	aw, err := NewAsyncWriter(&buf, 2)
	if err != nil {
		t.Fatalf("Failed to create asyncWriter: %v", err)
	}

	msg := []byte("hello async log\n")
	n, err := aw.Write(msg)
	if err != nil {
		t.Errorf("Unexpected error on Write: %v", err)
	}
	if n != len(msg) {
		t.Errorf("Expected written bytes %d, got %d", len(msg), n)
	}

	// Close and ensure all logs are flushed
	if err := aw.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if !strings.Contains(buf.String(), "hello async log") {
		t.Errorf("Expected log message in buffer, got: %q", buf.String())
	}
}

// --- Tests for SetupLogger ---

func TestSetupLogger_DefaultLevel(t *testing.T) {
	// Unset LOG_LEVEL to test default
	os.Unsetenv("LOG_LEVEL")
	log, closeFunc := SetupLogger()
	defer closeFunc()

	if log.GetLevel() != logrus.InfoLevel {
		t.Errorf("Expected default log level info, got %v", log.GetLevel())
	}
}

func TestSetupLogger_InvalidLevel(t *testing.T) {
	os.Setenv("LOG_LEVEL", "notalevel")
	defer os.Unsetenv("LOG_LEVEL")
	log, closeFunc := SetupLogger()
	defer closeFunc()

	if log.GetLevel() != logrus.InfoLevel {
		t.Errorf("Expected fallback to info level, got %v", log.GetLevel())
	}
}

// --- Tests for LogrusGormWriter ---

type testHook struct {
	mu   sync.Mutex
	msgs []string
}

func (h *testHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *testHook) Fire(e *logrus.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.msgs = append(h.msgs, e.Message)
	return nil
}

func TestLogrusGormWriter_Printf(t *testing.T) {
	logger := logrus.New()
	hook := &testHook{}
	logger.AddHook(hook)
	logger.SetLevel(logrus.TraceLevel)

	gormWriter := &LogrusGormWriter{Logger: logger}
	gormWriter.Printf("gorm query: %s", "SELECT 1")

	found := false
	hook.mu.Lock()
	for _, msg := range hook.msgs {
		if strings.Contains(msg, "gorm query: SELECT 1") {
			found = true
			break
		}
	}
	hook.mu.Unlock()
	if !found {
		t.Errorf("Expected gorm query log in hook messages, got: %v", hook.msgs)
	}
}

// --- Test LoggerConfig defaults ---

func TestLoggerConfig_Default(t *testing.T) {
	var cfg LoggerConfig
	if cfg.LogLevel != "" {
		t.Errorf("Expected empty LogLevel before env/config load, got %q", cfg.LogLevel)
	}
}

// --- Test asyncWriter closes underlying writer if io.Closer and not os.Stdout ---

type closerBuffer struct {
	bytes.Buffer
	closed bool
}

func (c *closerBuffer) Close() error {
	c.closed = true
	return nil
}

func TestAsyncWriter_Close_ClosesUnderlyingWriter(t *testing.T) {
	cb := &closerBuffer{}
	aw, err := NewAsyncWriter(cb, 2)
	if err != nil {
		t.Fatalf("Failed to create asyncWriter: %v", err)
	}
	aw.Write([]byte("test\n"))
	aw.Close()
	if !cb.closed {
		t.Errorf("Expected underlying writer to be closed")
	}
}

func TestAsyncWriter_Close_DoesNotCloseStdout(t *testing.T) {
	aw, err := NewAsyncWriter(os.Stdout, 2)
	if err != nil {
		t.Fatalf("Failed to create asyncWriter: %v", err)
	}
	aw.Write([]byte("test\n"))
	// Should not panic or close os.Stdout
	if err := aw.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// --- Test processQueue handles done signal with empty queue ---

func TestAsyncWriter_ProcessQueue_DoneEmptyQueue(t *testing.T) {
	var buf bytes.Buffer
	aw, err := NewAsyncWriter(&buf, 1)
	if err != nil {
		t.Fatalf("Failed to create asyncWriter: %v", err)
	}
	aw.Close()
	// Should not deadlock or panic
}
