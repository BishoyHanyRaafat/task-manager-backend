package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type InitResult struct {
	Writer io.Writer
	Close  func() error
}

// Init configures:
// - stdlib log output (log.Printf, etc.)
// - gin default writers (Logger/Recovery)
//
// It always logs to stdout. If logFilePath is non-empty, it also appends to that file.
func Init(logFilePath string) (*InitResult, error) {
	out := os.Stdout

	res := &InitResult{
		Writer: out,
		Close:  func() error { return nil },
	}

	if logFilePath == "" {
		gin.DefaultWriter = out
		gin.DefaultErrorWriter = out
		log.SetOutput(out)
		log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC)
		return res, nil
	}

	if err := os.MkdirAll(filepath.Dir(logFilePath), 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	mw := io.MultiWriter(out, f)
	res.Writer = mw
	res.Close = f.Close

	gin.DefaultWriter = mw
	gin.DefaultErrorWriter = mw
	log.SetOutput(mw)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC)

	return res, nil
}
