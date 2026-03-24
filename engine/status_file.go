package engine

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const StatusFileName = "status.json"

const defaultStatusFileWriteInterval = 3 * time.Second

type statusFileWriter struct {
	path     string
	snapshot func() StatusSnapshot

	writeInterval time.Duration
	writeFile     func(path string, snapshot StatusSnapshot) error
	logf          func(format string, args ...any)

	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
	doneCh  chan struct{}
}

func newStatusFileWriter(path string, snapshot func() StatusSnapshot) *statusFileWriter {
	return &statusFileWriter{
		path:          path,
		snapshot:      snapshot,
		writeInterval: defaultStatusFileWriteInterval,
		writeFile:     writeStatusSnapshotFile,
		logf:          log.Printf,
	}
}

func (w *statusFileWriter) Start() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})
	w.running = true
	w.stopCh = stopCh
	w.doneCh = doneCh
	writeInterval := w.writeInterval
	if writeInterval <= 0 {
		writeInterval = defaultStatusFileWriteInterval
	}
	w.mu.Unlock()

	go w.run(stopCh, doneCh, writeInterval)
}

func (w *statusFileWriter) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	stopCh := w.stopCh
	doneCh := w.doneCh
	w.running = false
	w.stopCh = nil
	w.doneCh = nil
	w.mu.Unlock()

	close(stopCh)
	<-doneCh
	_ = os.Remove(w.path)
}

func (w *statusFileWriter) run(stopCh <-chan struct{}, doneCh chan<- struct{}, writeInterval time.Duration) {
	defer close(doneCh)

	writeFailed := false
	lastWriteErr := ""
	w.writeSnapshot(&writeFailed, &lastWriteErr)

	ticker := time.NewTicker(writeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			w.writeSnapshot(&writeFailed, &lastWriteErr)
		}
	}
}

func (w *statusFileWriter) writeSnapshot(writeFailed *bool, lastWriteErr *string) {
	err := w.writeFile(w.path, w.snapshot())
	if err != nil {
		errText := err.Error()
		if !*writeFailed || *lastWriteErr != errText {
			w.logf("status: writing %s failed: %v", filepath.Base(w.path), err)
		}
		*writeFailed = true
		*lastWriteErr = errText
		return
	}

	if *writeFailed {
		w.logf("status: writing %s recovered", filepath.Base(w.path))
	}
	*writeFailed = false
	*lastWriteErr = ""
}

func writeStatusSnapshotFile(path string, snapshot StatusSnapshot) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".status-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	cleanup = false
	return nil
}
