package logging

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/game_tool_box/internal/config"
)

var (
	mu     sync.Mutex
	logger *log.Logger
	file   *os.File
)

// Init initializes the app logger.
//
// Safe to call multiple times.
func Init() {
	mu.Lock()
	defer mu.Unlock()
	if logger != nil {
		return
	}

	dir, err := config.Dir()
	if err != nil {
		// fallback to stdout only
		logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
		return
	}
	_ = os.MkdirAll(dir, 0o755)

	p := filepath.Join(dir, "app.log")
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
		return
	}
	file = f
	mw := io.MultiWriter(os.Stdout, f)
	logger = log.New(mw, "", log.LstdFlags|log.Lmicroseconds)
	logger.Printf("logger initialized: %s", p)
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	if logger != nil {
		logger.Printf("logger closed")
	}
	logger = nil
	if file != nil {
		_ = file.Close()
		file = nil
	}
}

func Infof(format string, args ...any) {
	Init()
	mu.Lock()
	l := logger
	mu.Unlock()
	if l == nil {
		return
	}
	l.Printf("INFO "+format, args...)
}

func Errorf(format string, args ...any) {
	Init()
	mu.Lock()
	l := logger
	mu.Unlock()
	if l == nil {
		return
	}
	l.Printf("ERROR "+format, args...)
}
