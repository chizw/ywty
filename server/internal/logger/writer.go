package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// lumberjackWriter 提供轻量的日志文件滚动写入
// 避免引入额外 lumberjack 依赖；P0 阶段够用，后续可替换
type lumberjackWriter struct {
	filename   string
	maxSize    int // MB
	maxAge     int // day
	maxBackups int
	compress   bool

	mu       sync.Mutex
	file     *os.File
	size     int64
	openTime time.Time
}

func (w *lumberjackWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.open(); err != nil {
			return 0, err
		}
	}

	if w.maxSize > 0 && w.size+int64(len(p)) > int64(w.maxSize)*1024*1024 {
		_ = w.rotate()
	}

	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

func (w *lumberjackWriter) open() error {
	if err := os.MkdirAll(filepath.Dir(w.filename), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(w.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	st, _ := f.Stat()
	w.file = f
	w.size = st.Size()
	w.openTime = time.Now()
	return nil
}

func (w *lumberjackWriter) rotate() error {
	if w.file != nil {
		_ = w.file.Close()
		w.file = nil
	}

	ts := w.openTime.Format("20060102-150405")
	base := w.filename + "." + ts
	_ = os.Rename(w.filename, base)

	if w.maxAge > 0 {
		go w.cleanup(base)
	}

	return w.open()
}

func (w *lumberjackWriter) cleanup(skip string) {
	dir := filepath.Dir(w.filename)
	base := filepath.Base(w.filename)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -w.maxAge)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == base || e.Name() == skip {
			continue
		}
		if !strings.HasPrefix(e.Name(), base+".") {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		if fi.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

func (w *lumberjackWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}

// 确保实现了 io.WriteCloser
var _ io.WriteCloser = (*lumberjackWriter)(nil)

// strings 引用避免编译器剔除
var _ = strings.HasPrefix
